package analyzer

import "go/ast"

type NilcheckFinder interface {
	Visit(node ast.Node) ast.Visitor
	Found() bool
}

type NewNilcheckFinder func(varbl Varbl) NilcheckFinder

type nilcheckFinder struct {
	varbl         Varbl
	hasIfNilCheck bool
}

func (v *nilcheckFinder) Visit(node ast.Node) ast.Visitor {
	if v == nil {
		return nil
	}

	t, ok := node.(*ast.IfStmt)
	if !ok {
		return v
	}

	if !v.isIfNilCheck(t) {
		return v
	}

	v.hasIfNilCheck = true
	return nil
}

func (v *nilcheckFinder) isIfNilCheck(ifStmt *ast.IfStmt) bool {
	if v == nil {
		return false
	}

	result := false

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

		nilCheck := (identX.Name == v.varbl.GetName() && identY.Name == "nil") ||
			(identY.Name == v.varbl.GetName() && identX.Name == "nil")

		if !nilCheck {
			return av
		}

		result = true

		return nil
	})

	ast.Walk(fv, ifStmt)

	return result
}

func (v *nilcheckFinder) Found() bool {
	if v == nil {
		return false
	}

	return v.hasIfNilCheck
}
