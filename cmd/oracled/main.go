package main

import (
	"github.com/medibloc/panacea-oracle/cmd/oracled/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
