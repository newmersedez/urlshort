package analyzer

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "urlshortlint",
	Doc:      "reports panic calls and os.Exit/log.Fatal usage outside of main function in main package",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.CallExpr)(nil)}

	insp.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			return false
		}

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		file := fileOf(stack)
		if file == nil {
			return true
		}

		if isGenerated(file) || isTestFile(pass.Fset.Position(file.Pos()).Filename) {
			return true
		}

		checkPanic(pass, call)
		checkExitOutsideMain(pass, call, stack)

		return true
	})

	return nil, nil
}

func checkPanic(pass *analysis.Pass, call *ast.CallExpr) {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return
	}
	if ident.Name == "panic" {
		pass.Reportf(call.Pos(), "use of built-in panic is not allowed")
	}
}

func checkExitOutsideMain(pass *analysis.Pass, call *ast.CallExpr, stack []ast.Node) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return
	}

	pkg := ident.Name
	fn := sel.Sel.Name
	if !isForbiddenCall(pkg, fn) {
		return
	}

	enclosing := enclosingFuncDecl(stack)
	isMainFunc := pass.Pkg.Name() == "main" && enclosing != nil && enclosing.Name.Name == "main"
	if !isMainFunc {
		pass.Reportf(call.Pos(), "%s.%s must not be called outside of main function in main package", pkg, fn)
	}
}

func isForbiddenCall(pkg, fn string) bool {
	return (pkg == "os" && fn == "Exit") || (pkg == "log" && fn == "Fatal")
}

func enclosingFuncDecl(stack []ast.Node) *ast.FuncDecl {
	for i := len(stack) - 1; i >= 0; i-- {
		if fd, ok := stack[i].(*ast.FuncDecl); ok {
			return fd
		}
	}
	return nil
}

func fileOf(stack []ast.Node) *ast.File {
	for _, n := range stack {
		if f, ok := n.(*ast.File); ok {
			return f
		}
	}
	return nil
}

func isGenerated(file *ast.File) bool {
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			if strings.HasPrefix(c.Text, "// Code generated") {
				return true
			}
		}
	}
	return false
}

func isTestFile(filename string) bool {
	return strings.HasSuffix(filename, "_test.go")
}
