package main

import (
	"fmt"
	"os"
)

func main() {
	if err := RunSplunkAIAnalyzer(); err != nil {
		fmt.Printf("❌ Splunk Analyzer Failed: %v\n", err)
		os.Exit(1)
	}
}