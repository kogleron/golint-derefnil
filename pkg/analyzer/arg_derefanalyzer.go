package analyzer

import (
	"go/ast"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/analysis"
)

type ArgDerefAnalyzer struct {
	derefFinderFactory        NewDerefFinder
	nilcheckFinderFactory     NewNilcheckFinder
	callselectorFinderFactory NewCallSelectorFinder

	derefs []Dereference
}

func NewArgDerefAnalyzer(
	derefFinderFactory NewDerefFinder,
	nilcheckFinderFactory NewNilcheckFinder,
	callselectorFinderFactory NewCallSelectorFinder,
) *ArgDerefAnalyzer {
	return &ArgDerefAnalyzer{
		derefFinderFactory:        derefFinderFactory,
		nilcheckFinderFactory:     nilcheckFinderFactory,
		callselectorFinderFactory: callselectorFinderFactory,

		derefs: []Dereference{},
	}
}

func (a *ArgDerefAnalyzer) Visit(node ast.Node) (w ast.Visitor) {
	funcDecl, ok := node.(*ast.FuncDecl)
	if ok {
		a.collectDerefs(funcDecl)
	}

	return a
}

func (a *ArgDerefAnalyzer) IterateDerefs(v DerefReportVisitor) error {
	if a == nil {
		return nil
	}

	for j := range a.derefs {
		if !v(&a.derefs[j]) {
			break
		}
	}

	return nil
}

func (a *ArgDerefAnalyzer) ReportDeref(pass *analysis.Pass, deref *Dereference) error {
	if deref == nil {
		return nil
	}

	arg, ok := deref.varbl.(*Argument)
	if !ok {
		return errors.New("expect argument")
	}

	pass.Reportf(
		deref.Expr.Pos(),
		"no nil check for the arg '%s' of '%s' before dereferencing",
		arg.Name,
		arg.FuncDecl.Name.Name,
	)

	return nil
}

func (a *ArgDerefAnalyzer) collectDerefs(funcDecl *ast.FuncDecl) {
	if a == nil {
		return
	}

	args := a.getArgs(funcDecl)
	if len(args) == 0 {
		return
	}

	callselFinder := a.callselectorFinderFactory()
	chainVisitor := NewChainVisitor(callselFinder)
	derefFinders := map[*Argument]DerefFinder{}
	nilcheckFinders := map[*Argument]NilcheckFinder{}

	for _, arg := range args {
		derefFinders[arg] = a.derefFinderFactory(arg)
		nilcheckFinders[arg] = a.nilcheckFinderFactory(arg)
		chainVisitor.AddVisitor(derefFinders[arg])
		chainVisitor.AddVisitor(nilcheckFinders[arg])
	}

	ast.Walk(chainVisitor, funcDecl.Body)

	for arg, derefFinder := range derefFinders {
		nilcheckFinder, ok := nilcheckFinders[arg]
		if !ok || nilcheckFinder.Found() {
			continue
		}

		derefs := derefFinder.GetDerefs()
		if len(derefs) == 0 {
			return
		}

		for j := range derefs {
			if callselFinder.Found(derefs[j].Expr) {
				continue
			}

			a.derefs = append(a.derefs, derefs[j])
		}
	}
}

func (a *ArgDerefAnalyzer) getArgs(funcDecl *ast.FuncDecl) []*Argument {
	if funcDecl.Type == nil || funcDecl.Type.Params == nil || len(funcDecl.Type.Params.List) == 0 {
		return nil
	}

	result := []*Argument{}

	for _, field := range funcDecl.Type.Params.List {
		arg := a.getArg(funcDecl, field)
		if arg == nil {
			continue
		}

		result = append(result, arg)
	}

	return result
}

func (a *ArgDerefAnalyzer) getArg(funcDecl *ast.FuncDecl, field *ast.Field) *Argument {
	if field == nil || field.Type == nil {
		return nil
	}

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
			return &Argument{
				Name:     name.Name,
				TypeName: ident.Name,
				FuncDecl: funcDecl,
			}
		}
	}

	return nil
}
