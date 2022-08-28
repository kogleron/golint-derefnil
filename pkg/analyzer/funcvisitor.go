package analyzer

import "go/ast"

type FuncVisitor struct {
	funcVisitor func(v ast.Visitor, node ast.Node) ast.Visitor
}

func NewFuncVisitor(funcVisitor func(v ast.Visitor, node ast.Node) ast.Visitor) *FuncVisitor {
	return &FuncVisitor{
		funcVisitor: funcVisitor,
	}
}

func (v *FuncVisitor) Visit(node ast.Node) ast.Visitor {
	if v == nil {
		return nil
	}

	return v.funcVisitor(v, node)
}
