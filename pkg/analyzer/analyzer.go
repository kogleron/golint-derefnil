package analyzer

import (
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "recvnil",
	Doc:  "Checks that there is a check for nil for the dereferenced receiver in a method",
	Run: func(pass *analysis.Pass) (interface{}, error) {
		proc := newProcessor()

		return proc.Run(pass)
	},
}
