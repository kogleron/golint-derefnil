package analyzer

import "go/ast"

type Dereference struct {
	varbl Varbl
	Name  string
	Expr  ast.Expr
}
