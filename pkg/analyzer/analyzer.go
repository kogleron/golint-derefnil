package analyzer

import (
	"flag"
	"go/ast"
	"sync"

	"golang.org/x/tools/go/analysis"
)

var params = &ProcParams{
	DumpIgnoreLock: &sync.Mutex{},
}

var Analyzer = &analysis.Analyzer{
	Name:  "recvnil",
	Doc:   "Checks that there is a check for nil for the dereferenced receiver in a method",
	Flags: *getFlags(params),
	Run: func(pass *analysis.Pass) (interface{}, error) {
		derefAnalyzers := []DerefAnalyzer{
			NewRecvDerefAnalyzer(
				newDerefFinder,
				newNilcheckFinder,
			),
			NewArgDerefAnalyzer(
				newDerefFinder,
				newNilcheckFinder,
				newCallSelectorFinder,
			),
		}

		proc, err := NewProcessor(params, derefAnalyzers)
		if err != nil {
			return nil, err
		}

		return proc.Run(pass)
	},
}

func getFlags(params *ProcParams) *flag.FlagSet {
	var flags *flag.FlagSet = flag.NewFlagSet("golint-derefnil", flag.ExitOnError)

	flags.BoolVar(&params.DumpIgnore, "dump-ignore", false, "Dumps errors into the given file.")
	flags.StringVar(&params.DumpIgnoreFile, "ignore-file", "./.derefnil.ignore", "Path to the file with ingored lines")

	return flags
}

func newDerefFinder(varbl Varbl) DerefFinder {
	return &derefFinder{
		varbl:  varbl,
		derefs: []Dereference{},
	}
}

func newNilcheckFinder(varbl Varbl) NilcheckFinder {
	return &nilcheckFinder{
		varbl: varbl,
	}
}

func newCallSelectorFinder() CallSelectorFinder {
	return &callSelectorFinder{
		exprs: map[ast.Expr]struct{}{},
	}
}
