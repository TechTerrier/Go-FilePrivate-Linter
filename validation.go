package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var foundViolation = false

type variableUsage struct {
	Name      string
	Position  token.Position
	UsageType string
	DeclFile  string // File where variable is declared (if known)
	UsageFile string
}

// hasFilePrivateComment - Checks if the fileprivate comment is present.
func hasFilePrivateComment(cg *ast.CommentGroup) bool {
	if cg == nil {
		return false
	}
	for _, comment := range cg.List {
		if strings.Contains(strings.ToLower(comment.Text), "fileprivate") {
			return true
		}
	}
	return false
}

func getFilePrivateVariablesFromFile(file string) (vars []string, err error) {
	fileSet := token.NewFileSet()
	node, err := parser.ParseFile(fileSet, file, nil, parser.ParseComments)
	if err != nil {
		log.Println("Failed to parse file:", err)
		return nil, err
	}

	for _, decl := range node.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}

		// Check for "fileprivate" comment above the entire declaration
		blockHasFilePrivate := hasFilePrivateComment(gen.Doc)

		for _, spec := range gen.Specs {
			valSpec := spec.(*ast.ValueSpec)
			// Check for "fileprivate" comment above this var spec or inline
			specHasFilePrivate := hasFilePrivateComment(valSpec.Doc) || hasFilePrivateComment(valSpec.Comment)

			if blockHasFilePrivate || specHasFilePrivate {
				for _, name := range valSpec.Names {
					vars = append(vars, name.Name)
				}
			}
		}
	}
	return vars, nil
}

func getUsages(filePath string) []variableUsage {

	if len(os.Args) < 2 {
		log.Println("Usage: ./FilePrivateLinter <file-to-analyze>")
		os.Exit(1)
	}

	pkgDir := filepath.Dir(filePath)
	fileSet := token.NewFileSet()

	// Step 1: Parse the entire package to find all declarations
	pkgVars := make(map[string]string) // var name -> declaring file
	packageList, err := parser.ParseDir(fileSet, pkgDir, nil, parser.AllErrors)
	if err != nil {
		log.Fatal("Error parsing package:", err)
	}

	// Collect all package-level variables and local declarations
	localVars := make(map[string]bool)
	for _, pkg := range packageList {
		for filename, file := range pkg.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				if decl, ok := n.(*ast.GenDecl); ok && decl.Tok == token.VAR {
					for _, spec := range decl.Specs {
						if valueSpec, ok := spec.(*ast.ValueSpec); ok {
							for _, name := range valueSpec.Names {
								if name.Name != "_" {
									pkgVars[name.Name] = filename
								}
							}
						}
					}
				}
				if assign, ok := n.(*ast.AssignStmt); ok {
					for _, expr := range assign.Lhs {
						if ident, ok := expr.(*ast.Ident); ok {
							localVars[ident.Name] = true
						}
					}
				}
				return true
			})
		}
	}

	// Step 2: Parse the target file and find all variable usages
	file, err := parser.ParseFile(fileSet, filePath, nil, parser.AllErrors)
	if err != nil {
		log.Fatal("Error parsing file:", err)
	}

	var usages []variableUsage

	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Ident:
			// Skip blank identifiers and package names
			if x.Name == "_" || (x.Obj != nil && x.Obj.Kind == ast.Pkg) {
				return true
			}

			// Check if this is a known variable (package or local)
			_, isPkgVar := pkgVars[x.Name]
			_, isLocalVar := localVars[x.Name]
			isPotentialVar := !ast.IsExported(x.Name) && x.Obj == nil

			if isPkgVar || isLocalVar || isPotentialVar {
				usageType := "direct"
				switch parent := n.(type) {
				case *ast.SelectorExpr:
					if parent.X == x {
						usageType = "method receiver"
					}
				case *ast.CallExpr:
					if parent.Fun == x {
						usageType = "function call"
					}
				}

				declFile := ""
				if isPkgVar {
					declFile = pkgVars[x.Name]
				}

				usages = append(usages, variableUsage{
					Name:      x.Name,
					Position:  fileSet.Position(x.Pos()),
					UsageType: usageType,
					DeclFile:  declFile,
					UsageFile: filePath,
				})
			}

		case *ast.CallExpr:
			// Check function arguments for variables
			for _, arg := range x.Args {
				if ident, ok := arg.(*ast.Ident); ok && ident.Name != "_" {
					_, isPkgVar := pkgVars[ident.Name]
					_, isLocalVar := localVars[ident.Name]
					isPotentialVar := !ast.IsExported(ident.Name) && ident.Obj == nil

					if isPkgVar || isLocalVar || isPotentialVar {
						declFile := ""
						if isPkgVar {
							declFile = pkgVars[ident.Name]
						}

						usages = append(usages, variableUsage{
							Name:      ident.Name,
							Position:  fileSet.Position(ident.Pos()),
							UsageType: "function argument",
							DeclFile:  declFile,
							UsageFile: filePath,
						})
					}
				}
			}

		case *ast.SelectorExpr:
			// Check for method calls on variables (x.Method())
			if ident, ok := x.X.(*ast.Ident); ok && ident.Name != "_" {
				_, isPkgVar := pkgVars[ident.Name]
				_, isLocalVar := localVars[ident.Name]
				isPotentialVar := !ast.IsExported(ident.Name) && ident.Obj == nil

				if isPkgVar || isLocalVar || isPotentialVar {
					declFile := ""
					if isPkgVar {
						declFile = pkgVars[ident.Name]
					}

					usages = append(usages, variableUsage{
						Name:      ident.Name,
						Position:  fileSet.Position(ident.Pos()),
						UsageType: "method receiver",
						DeclFile:  declFile,
						UsageFile: filePath,
					})

				}
			}
		}
		return true
	})

	return usages
}

func checkFile(filePath string) []variableUsage {
	usages := getUsages(filePath)
	var filteredUsages []variableUsage
	for _, usage := range usages {

		if usage.DeclFile == "" {
			continue
		}

		privateVars, privateErr := getFilePrivateVariablesFromFile(usage.DeclFile)
		if privateErr != nil {
			continue
		}

		if slices.Contains(privateVars, usage.Name) && usage.DeclFile != usage.UsageFile {
			foundViolation = true
			filteredUsages = append(filteredUsages, usage)
		}
	}

	return filteredUsages
}

func printViolations(violations []variableUsage) {
	for _, usage := range violations {
		log.Printf("- %s (declared in %s) at %s used as %s\n",
			usage.Name,
			usage.DeclFile,
			usage.Position,
			usage.UsageType)
	}
}
