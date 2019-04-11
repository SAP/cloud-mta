package commands

import (
	"github.com/SAP/cloud-mta/internal/logs"
	"github.com/spf13/cobra"
)

// Add new module
var AddModule = &cobra.Command{
	Use:   "modules",
	Short: "Add new module",
	Long:  "Add new module",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logs.Logger.Info("Adding new module...")
		return nil
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Add new resource
var AddResources = &cobra.Command{
	Use:   "resource",
	Short: "Add new resources",
	Long:  "Add new resources",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logs.Logger.Info("Adding new Resource...")
		return nil
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}
