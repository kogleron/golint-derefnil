package analyzer

import "go/ast"

type Receiver struct {
	Name     string
	TypeName string
	FuncDecl *ast.FuncDecl
}

func (r Receiver) GetName() string {
	return r.Name
}
