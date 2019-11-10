package main

import (
	"fmt"
	"os"

	"github.com/spotinst/spotctl/internal/cmd"
)

func main() {
	if err := cmd.New(os.Stdin, os.Stdout, os.Stderr).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Exited with error: %v\n", err)
		os.Exit(1)
	}
}
