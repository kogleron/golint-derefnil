package analyzer

import (
	"bufio"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/tools/go/analysis"
)

var dumpedLines map[string]struct{} = map[string]struct{}{}

type ProcParams struct {
	DumpIgnore     bool
	DumpIgnoreLock *sync.Mutex
	DumpIgnoreFile string
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

	_, err := os.Open(p.params.DumpIgnoreFile)
	if err != nil {
		return nil
	}

	f, err := os.OpenFile(p.params.DumpIgnoreFile, os.O_RDONLY, 0o755)
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

	var openFlags int
	_, noexists := os.Open(p.params.DumpIgnoreFile)
	if noexists != nil {
		openFlags = os.O_WRONLY | os.O_CREATE
	} else {
		openFlags = os.O_WRONLY | os.O_APPEND
	}

	f, err := os.OpenFile(p.params.DumpIgnoreFile, openFlags, 0o755)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	for j := range p.analyzers {
		var errItrt error
		err = p.analyzers[j].IterateDerefs(func(deref *Dereference) bool {
			line := p.buildIgnoreLine(pass, deref.Expr.Pos())

			if _, dumped := dumpedLines[line]; dumped {
				return true
			}

			_, errItrt = w.WriteString(line + "\n")
			if errItrt != nil {
				return false
			}

			dumpedLines[line] = struct{}{}

			return true
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

func (p *Processor) buildIgnoreLine(pass *analysis.Pass, pos token.Pos) string {
	file := pass.Fset.File(pos)
	r := regexp.MustCompile("^.*?([^/]+)$")
	return fmt.Sprintf(
		"%s/%s %d",
		pass.Pkg.Path(),
		r.ReplaceAllString(file.Name(), "$1"),
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
	if deref == nil {
		return true
	}

	ignoreLine := p.buildIgnoreLine(pass, deref.Expr.Pos())
	_, ignored := p.ignoreLines[ignoreLine]

	return ignored
}
