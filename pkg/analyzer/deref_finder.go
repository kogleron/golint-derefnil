package analyzer

import "go/ast"

type DerefFinder interface {
	Visit(node ast.Node) (w ast.Visitor)
	GetDerefs() []Dereference
}

type NewDerefFinder func(varbl Varbl) DerefFinder

type derefFinder struct {
	varbl  Varbl
	derefs []Dereference
}

func (f *derefFinder) Visit(node ast.Node) (w ast.Visitor) {
	if f == nil {
		return nil
	}

	selectorExpr, ok := node.(*ast.SelectorExpr)
	if !ok {
		return f
	}

	f.processDeref(selectorExpr)

	return f
}

func (f *derefFinder) processDeref(selectorExpr *ast.SelectorExpr) {
	if f == nil {
		return
	}

	ident, ok := selectorExpr.X.(*ast.Ident)
	if !ok {
		return
	}

	if ident.Name != f.varbl.GetName() {
		return
	}

	f.derefs = append(f.derefs, Dereference{
		varbl:        f.varbl,
		Name:         selectorExpr.Sel.Name,
		SelectorExpr: selectorExpr,
	})
}

func (f *derefFinder) GetDerefs() []Dereference {
	if f == nil {
		return nil
	}

	return f.derefs
}
