package analyzer

import "go/ast"

type ChainVisitor struct {
	visitors []ast.Visitor
}

func NewChainVisitor(visitors ...ast.Visitor) *ChainVisitor {
	return &ChainVisitor{
		visitors: visitors,
	}
}

func (v *ChainVisitor) AddVisitor(visitor ast.Visitor) {
	if v == nil {
		return
	}

	v.visitors = append(v.visitors, visitor)
}

func (v *ChainVisitor) Visit(node ast.Node) (w ast.Visitor) {
	if v == nil {
		return nil
	}

	if len(v.visitors) == 0 {
		return nil
	}

	leftVisitors := []ast.Visitor{}
	for _, visitor := range v.visitors {
		res := visitor.Visit(node)
		if res == nil {
			continue
		}

		leftVisitors = append(leftVisitors, visitor)
	}

	v.visitors = leftVisitors

	if len(v.visitors) == 0 {
		return nil
	}

	return v
}
