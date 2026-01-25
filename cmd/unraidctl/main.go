package main

import (
	"os"

	"github.com/jwmoss/unraidctl/cmd/unraidctl/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
