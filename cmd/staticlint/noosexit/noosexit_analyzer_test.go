package noosexit

import (
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestNoOsExitAnalyzer(t *testing.T) {
	testdata := filepath.Join(filepath.Dir(analysistest.TestData()), "testdata")
	analysistest.Run(t, testdata, NoOsExitAnalyzer)

}
