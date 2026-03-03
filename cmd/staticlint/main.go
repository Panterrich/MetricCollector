// Package main provides a command-line tool that runs multiple analyzers on Go source code.
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"

	"github.com/Panterrich/MetricCollector/pkg/checkers/exit"
	"github.com/gostaticanalysis/loopdefer"
	"github.com/ultraware/whitespace"
)

// main - initializes and runs multiple analyzers from the staticcheck package,
// as well as additional analyzers such as ExitCheckAnalyzer, PrintfAnalyzer, ShadowAnalyzer, and StructTagAnalyzer.
//
// The multichecker.Main function is used to run all these analyzers together.
func main() {
	checks := make([]*analysis.Analyzer, 0, len(staticcheck.Analyzers)+4)

	for _, v := range staticcheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	checks = append(checks, exit.ExitCheckAnalyzer)
	checks = append(checks, printf.Analyzer)
	checks = append(checks, shadow.Analyzer)
	checks = append(checks, structtag.Analyzer)

	checks = append(checks, loopdefer.Analyzer)
	checks = append(checks, whitespace.NewAnalyzer(&whitespace.Settings{
		MultiIf: true,
	}))

	multichecker.Main(checks...)
}
