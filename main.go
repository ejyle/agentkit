package main

import (
	"os"

	"github.com/ejyle/agentkit/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
