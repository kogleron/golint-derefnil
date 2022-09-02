package analyzer

import "go/ast"

type CallSelectorFinder interface {
	Visit(node ast.Node) (w ast.Visitor)
	Found(selectorExpr ast.Expr) bool
}

type NewCallSelectorFinder func() CallSelectorFinder

type callSelectorFinder struct {
	exprs map[ast.Expr]struct{}
}

func (f *callSelectorFinder) Visit(node ast.Node) (w ast.Visitor) {
	if f == nil {
		return nil
	}

	callExpr, ok := node.(*ast.CallExpr)
	if !ok {
		return f
	}

	selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return f
	}

	f.exprs[selectorExpr] = struct{}{}

	return f
}

func (f *callSelectorFinder) Found(expr ast.Expr) bool {
	if f == nil {
		return false
	}

	_, found := f.exprs[expr]

	return found
}
