package analyzer

import "go/ast"

type receiver struct {
	Name     string
	TypeName string
	FuncDecl *ast.FuncDecl
}

type dereference struct {
	Name         string
	SelectorExpr *ast.SelectorExpr
}
