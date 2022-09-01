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

type Processor struct {
	params      *ProcParams
	analyzers   []DerefAnalyzer
	ignoreLines map[string]struct{}
}

func NewProcessor(params *ProcParams, analyzers []DerefAnalyzer) (*Processor, error) {
	p := &Processor{
		ignoreLines: map[string]struct{}{},
		params:      params,
		analyzers:   analyzers,
	}

	err := p.loadIgnore()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Processor) loadIgnore() error {
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

func (p *Processor) Run(pass *analysis.Pass) (interface{}, error) {
	if p == nil {
		return nil, nil
	}

	visitor := NewChainVisitor()

	for j := range p.analyzers {
		visitor.AddVisitor(p.analyzers[j])
	}

	for _, f := range pass.Files {
		ast.Walk(visitor, f)
	}

	return nil, p.report(pass)
}

func (p *Processor) report(pass *analysis.Pass) error {
	if p == nil {
		return nil
	}

	if p.params.DumpIgnore {
		return p.dumpIgnore(pass)
	}

	return p.reportPass(pass)
}

func (p *Processor) dumpIgnore(pass *analysis.Pass) error {
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

	for j := range p.analyzers {
		var errItrt error
		err = p.analyzers[j].IterateDerefs(func(deref *Dereference) bool {
			line := p.buildIgnoreLine(wd, pass, deref.SelectorExpr.Pos())
			_, errItrt = w.WriteString(line + "\n")

			return errItrt == nil
		})

		if errItrt != nil {
			return errItrt
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Processor) buildIgnoreLine(wd string, pass *analysis.Pass, pos token.Pos) string {
	file := pass.Fset.File(pos)
	return fmt.Sprintf(
		"%s %d",
		strings.Replace(file.Name(), wd+"/", "", 1),
		file.Position(pos).Line,
	)
}

func (p *Processor) reportPass(pass *analysis.Pass) error {
	if p == nil {
		return nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	for j := range p.analyzers {
		var errItrt error
		err = p.analyzers[j].IterateDerefs(func(deref *Dereference) bool {
			if p.ignoreDeref(wd, pass, deref) {
				return true
			}

			errItrt = p.analyzers[j].ReportDeref(pass, deref)
			return errItrt == nil
		})

		if err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func (p Processor) ignoreDeref(wd string, pass *analysis.Pass, deref *Dereference) bool {
	ignoreLine := p.buildIgnoreLine(wd, pass, deref.SelectorExpr.Pos())
	_, ignored := p.ignoreLines[ignoreLine]

	return ignored
}
