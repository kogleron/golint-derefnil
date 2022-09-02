package analyzer

import "go/ast"

type Argument struct {
	Name     string
	TypeName string
	FuncDecl *ast.FuncDecl
}

func (a Argument) GetName() string {
	return a.Name
}
