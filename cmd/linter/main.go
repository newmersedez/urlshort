package main

import (
	"github.com/newmersedez/urlshort/cmd/linter/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
