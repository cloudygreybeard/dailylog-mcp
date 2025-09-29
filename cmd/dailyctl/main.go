package main

import (
	"fmt"
	"os"

	"dailylog/cmd/dailyctl/cmd"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	if err := cmd.Execute(version, commit, date); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
