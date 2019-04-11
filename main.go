package main

import (
	"os"

	"github.com/SAP/cloud-mta/cmd"
)

func main() {
	// Execute CLI Root commands
	err := commands.Execute()
	if err != nil {
		os.Exit(1)
	}
}
