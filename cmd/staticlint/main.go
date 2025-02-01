package main

import (
	"github.com/kisielk/errcheck/errcheck"
	"github.com/ultraware/whitespace"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck/st1005"
	"honnef.co/go/tools/stylecheck/st1012"
)

func main() {
	whitespaceAnalyzer := whitespace.NewAnalyzer(nil)

	countAnalyzers := len(staticcheck.Analyzers) + 9

	mychecks := make([]*analysis.Analyzer, 0, countAnalyzers)

	mychecks = append(mychecks, []*analysis.Analyzer{
		printf.Analyzer,    // check consistency of Printf format strings and arguments
		shadow.Analyzer,    // check for possible unintended shadowing of variables
		structtag.Analyzer, // checks struct field tags are well formed
		shift.Analyzer,     // checks for shifts that exceed the width of an integer

		st1005.Analyzer, // incorrectly formatted error string
		st1012.Analyzer, // poorly chosen name for error variable

		errcheck.Analyzer,  // check for unchecked errors
		whitespaceAnalyzer, // whitespace is a linter that checks for unnecessary newlines at the start and end of functions, if, for, etc

		OSExitCheckAnalyzer, // check os.Exit() in main()
	}...)

	// Appending SA analyzers
	for _, v := range staticcheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}

	multichecker.Main(
		mychecks...,
	)
}
