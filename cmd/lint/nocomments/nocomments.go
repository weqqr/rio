package nocomments

import (
	"go/ast"
	"strings"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

const Name = "nocomments"

func init() {
	register.Plugin(Name, New)
}

type plugin struct{}

func New(any) (register.LinterPlugin, error) {
	return plugin{}, nil
}

func (plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{{
		Name: Name,
		Doc:  "reports comments that are not compiler directives or Go example output",
		Run:  run,
	}}, nil
}

func (plugin) GetLoadMode() string {
	return register.LoadModeSyntax
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, group := range file.Comments {
			if isExampleOutput(group) {
				continue
			}
			for _, comment := range group.List {
				if isDirective(comment.Text) {
					continue
				}
				pass.Reportf(comment.Pos(), "comments are not allowed")
			}
		}
	}
	return nil, nil
}

func isExampleOutput(group *ast.CommentGroup) bool {
	for _, comment := range group.List {
		if strings.HasPrefix(comment.Text, "// Output:") ||
			strings.HasPrefix(comment.Text, "// Unordered output:") {
			return true
		}
	}
	return false
}

func isDirective(text string) bool {
	return strings.HasPrefix(text, "//go:") ||
		strings.HasPrefix(text, "//line ") ||
		strings.HasPrefix(text, "//export ") ||
		strings.HasPrefix(text, "//nolint") ||
		strings.HasPrefix(text, "//lint:") ||
		strings.HasPrefix(text, "// +build") ||
		strings.HasPrefix(text, "//#cgo") ||
		strings.HasPrefix(text, "// #cgo")
}
