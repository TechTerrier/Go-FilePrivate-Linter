package main

import (
	"golang.org/x/tools/go/packages"
	"log"
	"os"
)

func main() {
	dir := os.Args[1]

	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes,
		Dir:   dir,
		Tests: false,
	}

	packageList, err := packages.Load(cfg, "./...")
	if err != nil {
		log.Fatalf("Failed to load packages: %v", err)
	}

	for _, pkg := range packageList {
		for _, file := range pkg.Syntax {
			filePath := pkg.Fset.Position(file.Pos()).Filename
			fileViolations := checkFile(filePath)
			printViolations(fileViolations)
		}
	}

	if foundViolation {
		os.Exit(1)
	}
}
