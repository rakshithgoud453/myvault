// Package main is the entry point for the MyVault password manager.
package main

import (
	"os"

	"github.com/rakshithgoud453/myvault/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
