package exit_test

import (
	"testing"

	"github.com/Panterrich/MetricCollector/pkg/checkers/exit"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestExitCheckAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), exit.ExitCheckAnalyzer, "./...")
}
