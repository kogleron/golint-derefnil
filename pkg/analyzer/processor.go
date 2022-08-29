package analyzer

import (
	"bufio"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"strings"
	"sync"

	"golang.org/x/tools/go/analysis"
)

const (
	FILE_IGNORE = ".recvnil.ignore"
)

type ProcParams struct {
	DumpIgnore     bool
	DumpIgnoreLock *sync.Mutex
}

type processor struct {
	recvDerefs   map[*receiver][]dereference
	structFields map[string]map[string]struct{}
	ignoreLines  map[string]struct{}
	params       *ProcParams
}

func newProcessor(params *ProcParams) (*processor, error) {
	p := &processor{
		recvDerefs:   make(map[*receiver][]dereference),
		structFields: make(map[string]map[string]struct{}),
		ignoreLines:  map[string]struct{}{},
		params:       params,
	}

	err := p.loadIgnore()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *processor) loadIgnore() error {
	if p == nil {
		return nil
	}

	_, err := os.Open(FILE_IGNORE)
	if err != nil {
		return nil
	}

	f, err := os.OpenFile(FILE_IGNORE, os.O_RDONLY, 0o755)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ignoreLine := strings.Trim(scanner.Text(), " ")
		if len(ignoreLine) == 0 {
			continue
		}

		p.ignoreLines[ignoreLine] = struct{}{}
	}

	err = scanner.Err()
	if err != nil {
		return err
	}

	return nil
}

func (p *processor) Run(pass *analysis.Pass) (interface{}, error) {
	inspect := p.getInspect(pass)

	for _, f := range pass.Files {
		ast.Inspect(f, inspect)
	}

	return nil, p.report(pass)
}

func (p *processor) report(pass *analysis.Pass) error {
	if p == nil {
		return nil
	}

	if p.params.DumpIgnore {
		return p.dumpIgnore(pass)
	}

	return p.reportPass(pass)
}

func (p *processor) dumpIgnore(pass *analysis.Pass) error {
	if p == nil {
		return nil
	}

	if p.params.DumpIgnoreLock == nil {
		return errors.New("no dump ignore lock")
	}

	p.params.DumpIgnoreLock.Lock()
	defer p.params.DumpIgnoreLock.Unlock()

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	var openFlags int
	_, noexists := os.Open(FILE_IGNORE)
	if noexists != nil {
		openFlags = os.O_WRONLY | os.O_CREATE
	} else {
		openFlags = os.O_WRONLY | os.O_APPEND
	}

	f, err := os.OpenFile(FILE_IGNORE, openFlags, 0o755)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	for recv, derefs := range p.recvDerefs {
		for _, deref := range derefs {
			_, isField := p.structFields[recv.TypeName][deref.Name]
			if !isField {
				continue
			}
			_, err = w.WriteString(p.buildIgnoreLine(wd, pass, deref.SelectorExpr.Pos()) + "\n")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *processor) buildIgnoreLine(wd string, pass *analysis.Pass, pos token.Pos) string {
	file := pass.Fset.File(pos)
	return fmt.Sprintf(
		"%s %d",
		strings.Replace(file.Name(), wd+"/", "", 1),
		file.Position(pos).Line,
	)
}

func (p *processor) reportPass(pass *analysis.Pass) error {
	if p == nil {
		return nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	for recv, derefs := range p.recvDerefs {
		for _, deref := range derefs {
			_, isField := p.structFields[recv.TypeName][deref.Name]
			if !isField {
				continue
			}

			ignoreLine := p.buildIgnoreLine(wd, pass, deref.SelectorExpr.Pos())
			_, ignored := p.ignoreLines[ignoreLine]
			if ignored {
				continue
			}

			pass.Reportf(
				deref.SelectorExpr.Pos(),
				"no nil check for the receiver '%s' of '%s' before accessing '%s'",
				recv.Name,
				recv.FuncDecl.Name.Name,
				deref.Name,
			)
		}
	}

	return nil
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

	derefs := recvVisitor.GetDerefs()

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
