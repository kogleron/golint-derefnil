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

	switch tn := node.(type) {
	case *ast.SelectorExpr:
		f.processDotDeref(tn)
	case *ast.StarExpr:
		f.processVarDeref(tn)
	}

	return f
}

func (f *derefFinder) processVarDeref(starExpr *ast.StarExpr) {
	if f == nil {
		return
	}

	ident, ok := starExpr.X.(*ast.Ident)
	if !ok || ident.Name != f.varbl.GetName() {
		return
	}

	f.derefs = append(f.derefs, Dereference{
		varbl: f.varbl,
		Name:  f.varbl.GetName(),
		Expr:  starExpr,
	})
}

func (f *derefFinder) processDotDeref(selectorExpr *ast.SelectorExpr) {
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
		varbl: f.varbl,
		Name:  selectorExpr.Sel.Name,
		Expr:  selectorExpr,
	})
}

func (f *derefFinder) GetDerefs() []Dereference {
	if f == nil {
		return nil
	}

	return f.derefs
}
