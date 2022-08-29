package analyzer

import "go/ast"

type RecvVisitor struct {
	recv          *receiver
	derefs        []dereference
	hasIfNilCheck bool
}

func NewRecvVisitor(recv *receiver) *RecvVisitor {
	return &RecvVisitor{
		recv:   recv,
		derefs: []dereference{},
	}
}

func (v *RecvVisitor) Visit(node ast.Node) (w ast.Visitor) {
	if v == nil {
		return nil
	}

	switch t := node.(type) {
	case *ast.SelectorExpr:
		v.processDeref(t)
	case *ast.IfStmt:
		if !v.hasIfNilCheck {
			v.processIfNilCheck(t)
		}
	}

	return v
}

func (v *RecvVisitor) processIfNilCheck(ifStmt *ast.IfStmt) {
	if v == nil {
		return
	}

	fv := NewFuncVisitor(func(av ast.Visitor, node ast.Node) ast.Visitor {
		binaryExpr, ok := node.(*ast.BinaryExpr)
		if !ok || binaryExpr.Op.String() != "==" {
			return av
		}

		identX, ok := binaryExpr.X.(*ast.Ident)
		if !ok {
			return av
		}
		identY, ok := binaryExpr.Y.(*ast.Ident)
		if !ok {
			return av
		}

		nilCheck := (identX.Name == v.recv.Name && identY.Name == "nil") ||
			(identY.Name == v.recv.Name && identX.Name == "nil")

		if !nilCheck {
			return av
		}

		v.hasIfNilCheck = true

		return nil
	})
	ast.Walk(fv, ifStmt)
}

func (v *RecvVisitor) processDeref(selectorExpr *ast.SelectorExpr) {
	if v == nil {
		return
	}

	ident, ok := selectorExpr.X.(*ast.Ident)
	if !ok {
		return
	}

	if ident.Name != v.recv.Name {
		return
	}

	v.derefs = append(v.derefs, dereference{
		Name:         selectorExpr.Sel.Name,
		SelectorExpr: selectorExpr,
	})
}

func (v *RecvVisitor) GetDerefs() []dereference {
	if v == nil {
		return nil
	}

	return v.derefs
}

func (v *RecvVisitor) FoundNilCheck() bool {
	if v == nil {
		return false
	}

	return v.hasIfNilCheck
}
