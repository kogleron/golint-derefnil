package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

type processor struct {
	recvDerefs   map[*receiver][]string
	structFields map[string]map[string]struct{}
}

func newProcessor() *processor {
	return &processor{
		recvDerefs:   make(map[*receiver][]string),
		structFields: make(map[string]map[string]struct{}),
	}
}

func (p *processor) Run(pass *analysis.Pass) (interface{}, error) {
	inspect := p.getInspect(pass)

	for _, f := range pass.Files {
		ast.Inspect(f, inspect)
	}

	p.report(pass)

	return nil, nil
}

func (p *processor) report(pass *analysis.Pass) {
	if p == nil {
		return
	}

	for recv, derefs := range p.recvDerefs {
		for _, deref := range derefs {
			_, isField := p.structFields[recv.TypeName][deref]
			if isField {
				pass.Reportf(
					recv.FuncDecl.Pos(),
					"no nil check for the receiver '%s' of '%s' before accessing '%s'",
					recv.Name,
					recv.FuncDecl.Name.Name,
					deref,
				)
			}
		}
	}
}

func (p *processor) getInspect(pass *analysis.Pass) func(node ast.Node) bool {
	return func(node ast.Node) bool {
		switch t := node.(type) {
		case *ast.FuncDecl:
			p.collectDerefs(t)
		case *ast.TypeSpec:
			p.collectStructFields(t)
		}
		return true
	}
}

func (p *processor) collectStructFields(typeSpec *ast.TypeSpec) {
	if p == nil {
		return
	}

	tp, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return
	}
	structName := typeSpec.Name.Name
	if p.structFields[structName] == nil {
		p.structFields[structName] = make(map[string]struct{})
	}

	for _, field := range tp.Fields.List {
		for _, fdName := range field.Names {
			p.structFields[structName][fdName.Name] = struct{}{}
		}
	}
}

func (p *processor) collectDerefs(funcDecl *ast.FuncDecl) {
	if p == nil {
		return
	}

	recv, ok := p.hasLinkReceiver(funcDecl)
	if !ok {
		return
	}

	recvVisitor := NewRecvVisitor(recv)

	ast.Walk(recvVisitor, funcDecl.Body)

	derefs := recvVisitor.GetDerefNames()

	if !recvVisitor.FoundNilCheck() && len(derefs) > 0 {
		p.recvDerefs[recv] = derefs
	}
}

func (p *processor) hasLinkReceiver(funcDecl *ast.FuncDecl) (*receiver, bool) {
	if funcDecl.Recv == nil {
		return nil, false
	}

	for _, field := range funcDecl.Recv.List {
		starExpr, ok := field.Type.(*ast.StarExpr)
		if !ok {
			return nil, false
		}

		ident, ok := starExpr.X.(*ast.Ident)
		if !ok {
			return nil, false
		}

		for _, name := range field.Names {
			if name.Name == "" {
				return nil, false
			}

			recv := &receiver{
				Name:     name.Name,
				TypeName: ident.Name,
				FuncDecl: funcDecl,
			}

			return recv, true //nolint
		}

		break //nolint
	}

	return nil, false
}
