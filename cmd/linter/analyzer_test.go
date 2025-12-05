package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzerPanic(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "./panic_check")
}

func TestAnalyzerCall(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "./call_check")
	analysistest.Run(t, analysistest.TestData(), Analyzer, "./call_check/main")
}
