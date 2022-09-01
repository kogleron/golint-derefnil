package analyzer

import "go/ast"

type receiver struct {
	Name     string
	TypeName string
	FuncDecl *ast.FuncDecl
}

func (r receiver) GetName() string {
	return r.Name
}
