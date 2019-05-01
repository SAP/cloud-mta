package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(createMtaCmd)
	addCmd.AddCommand(AddModule, AddResources)
}

// Parent command add any artifacts
var addCmd = &cobra.Command{
	Use:    "add",
	Short:  "Add artifacts",
	Long:   "Add artifacts",
	Hidden: true,
	Run:    nil,
}
