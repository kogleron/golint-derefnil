package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

type DerefReportVisitor func(deref *Dereference) bool

type DerefAnalyzer interface {
	Visit(node ast.Node) (w ast.Visitor)
	IterateDerefs(v DerefReportVisitor) error
	ReportDeref(pass *analysis.Pass, deref *Dereference) error
}
