package exit

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for exit in main funcs",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	isOsExit := func(node ast.Node) bool {
		if c, ok := node.(*ast.CallExpr); ok {
			if s, ok := c.Fun.(*ast.SelectorExpr); ok {
				if p, ok := s.X.(*ast.Ident); ok {
					if p.Name == "os" && s.Sel.Name == "Exit" {
						pass.Reportf(s.Pos(), `direct calling os.Exit in main func`)
					}
				}
			}
		}

		return true
	}

	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}

		isSkip := false

		for _, comment := range file.Comments {
			if strings.Contains(comment.Text(), `Code generated by 'go test'`) {
				isSkip = true
				break
			}
		}

		if isSkip {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			if fd, ok := node.(*ast.FuncDecl); ok {
				if fd.Name.Name == "main" {
					for _, v := range fd.Body.List {
						ast.Inspect(v, isOsExit)
					}

					return false
				}
			}

			return true
		})
	}

	return nil, nil
}
