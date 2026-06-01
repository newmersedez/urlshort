package analyzer_test

import (
	"testing"

	"github.com/newmersedez/urlshort/cmd/linter/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestPanicCheck(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), analyzer.Analyzer, "paniccheck")
}

func TestExitOutsideMainCheck(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), analyzer.Analyzer, "exitcheck")
}
