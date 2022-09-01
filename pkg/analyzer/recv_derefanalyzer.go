package analyzer

import (
	"go/ast"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/analysis"
)

type RecvDerefAnalyzer struct {
	derefFinderFactory    NewDerefFinder
	nilcheckFinderFactory NewNilcheckFinder

	derefs       []Dereference
	structFields map[string]map[string]struct{}
}

func NewRecvDerefAnalyzer(
	derefFinderFactory NewDerefFinder,
	nilcheckFinderFactory NewNilcheckFinder,
) *RecvDerefAnalyzer {
	return &RecvDerefAnalyzer{
		derefFinderFactory:    derefFinderFactory,
		nilcheckFinderFactory: nilcheckFinderFactory,
		derefs:                []Dereference{},
		structFields:          make(map[string]map[string]struct{}),
	}
}

func (a *RecvDerefAnalyzer) Visit(node ast.Node) (w ast.Visitor) {
	switch t := node.(type) {
	case *ast.FuncDecl:
		a.collectDerefs(t)
	case *ast.TypeSpec:
		a.collectStructFields(t)
	}

	return a
}

func (a *RecvDerefAnalyzer) IterateDerefs(v DerefReportVisitor) error {
	if a == nil {
		return nil
	}

	for j := range a.derefs {
		deref := &a.derefs[j]
		recv, ok := deref.varbl.(*receiver)
		if !ok {
			return errors.New("expected *receiver")
		}
		if recv == nil {
			return errors.New("need receiver")
		}

		_, isField := a.structFields[recv.TypeName][deref.Name]
		if !isField {
			continue
		}

		if !v(deref) {
			break
		}
	}

	return nil
}

func (a *RecvDerefAnalyzer) collectDerefs(funcDecl *ast.FuncDecl) {
	if a == nil {
		return
	}

	recv := a.getRefReceiver(funcDecl)
	if recv == nil {
		return
	}

	derefFinder := a.derefFinderFactory(recv)
	nilcheckFinder := a.nilcheckFinderFactory(recv)
	chainVisitor := NewChainVisitor(
		derefFinder,
		nilcheckFinder,
	)

	ast.Walk(chainVisitor, funcDecl.Body)

	if nilcheckFinder.Found() {
		return
	}

	derefs := derefFinder.GetDerefs()
	if len(derefs) == 0 {
		return
	}

	a.derefs = append(a.derefs, derefs...)
}

func (a *RecvDerefAnalyzer) getRefReceiver(funcDecl *ast.FuncDecl) *receiver {
	if funcDecl.Recv == nil {
		return nil
	}

	for _, field := range funcDecl.Recv.List {
		starExpr, ok := field.Type.(*ast.StarExpr)
		if !ok {
			return nil
		}

		ident, ok := starExpr.X.(*ast.Ident)
		if !ok {
			return nil
		}

		for _, name := range field.Names {
			if name.Name != "" {
				return &receiver{
					Name:     name.Name,
					TypeName: ident.Name,
					FuncDecl: funcDecl,
				}
			}
		}

		break //nolint
	}

	return nil
}

func (a *RecvDerefAnalyzer) ReportDeref(pass *analysis.Pass, deref *Dereference) error {
	recv, ok := deref.varbl.(*receiver)
	if !ok {
		return errors.New("expected *receiver")
	}
	if recv == nil {
		return errors.New("need receiver")
	}

	pass.Reportf(
		deref.SelectorExpr.Pos(),
		"no nil check for the receiver '%s' of '%s' before accessing '%s'",
		recv.Name,
		recv.FuncDecl.Name.Name,
		deref.Name,
	)

	return nil
}

func (a *RecvDerefAnalyzer) collectStructFields(typeSpec *ast.TypeSpec) {
	if a == nil {
		return
	}

	tp, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return
	}
	structName := typeSpec.Name.Name
	if a.structFields[structName] == nil {
		a.structFields[structName] = make(map[string]struct{})
	}

	for _, field := range tp.Fields.List {
		for _, fdName := range field.Names {
			a.structFields[structName][fdName.Name] = struct{}{}
		}
	}
}
