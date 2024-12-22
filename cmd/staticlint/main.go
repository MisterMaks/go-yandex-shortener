package main

import (
	"github.com/MisterMaks/go-yandex-shortener/cmd/errcheckanalyzer"
	"github.com/kisielk/errcheck/errcheck"
	"go.uber.org/nilaway"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck/st1005"
	"honnef.co/go/tools/stylecheck/st1012"
)

func main() {
	mychecks := make([]*analysis.Analyzer, 0, 4+len(staticcheck.Analyzers)+2+2+1)

	mychecks = []*analysis.Analyzer{
		errcheckanalyzer.ErrCheckAnalyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
	}

	for _, v := range staticcheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}

	mychecks = append(
		mychecks,
		st1005.Analyzer,
		st1012.Analyzer,
	)

	mychecks = append(
		mychecks,
		nilaway.Analyzer,
		errcheck.Analyzer,
	)

	mychecks = append(
		mychecks,
		OSExitCheckAnalyzer,
	)

	multichecker.Main(
		mychecks...,
	)
}
