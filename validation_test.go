package main

import (
	"go/ast"
	"reflect"
	"slices"
	"testing"
)

func TestHasFilePrivateComment(t *testing.T) {
	tests := []struct {
		name     string
		input    *ast.CommentGroup
		expected bool
	}{
		{
			name:     "nil comment group",
			input:    nil,
			expected: false,
		},
		{
			name: "no comments",
			input: &ast.CommentGroup{
				List: []*ast.Comment{},
			},
			expected: false,
		},
		{
			name: "comment without 'fileprivate'",
			input: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "// this is a test"},
					{Text: "// TODO: something"},
				},
			},
			expected: false,
		},
		{
			name: "single comment with 'fileprivate'",
			input: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "// fileprivate"},
				},
			},
			expected: true,
		},
		{
			name: "multiple comments with 'fileprivate' in one",
			input: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "// some text"},
					{Text: "// fileprivate"},
					{Text: "// other"},
				},
			},
			expected: true,
		},
		{
			name: "comment with 'fileprivate' in the middle",
			input: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "// some text with fileprivate tag"},
				},
			},
			expected: true,
		},
		{
			name: "comment with 'FilePrivate' different casing",
			input: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "// FilePrivate"},
				},
			},
			expected: true, // Case-sensitive check
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasFilePrivateComment(tt.input)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSampleSubProjectHasViolation(t *testing.T) {
	violations := checkFile("sampleProject/package1/file2.go")
	if len(violations) < 1 {
		t.Errorf("got %v, want at least 1 violation", len(violations))
	}

	var violatingStrs []string

	for _, violation := range violations {
		// Prevent duplicates
		if slices.Contains(violatingStrs, violation.Name) {
			return
		}
		violatingStrs = append(violatingStrs, violation.Name)
	}

	expectedStrs := []string{"a", "b", "d", "e"}

	if !reflect.DeepEqual(violatingStrs, expectedStrs) {
		t.Errorf("got %v, want %v", violatingStrs, expectedStrs)
	}

}

func TestSampleSubProjectHasNoViolation(t *testing.T) {
	violations := checkFile("sampleProject/package2/test.go")
	if len(violations) != 0 {
		t.Errorf("got %v, want %v", len(violations), 0)
	}
}
