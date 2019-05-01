package commands

import (
	"github.com/SAP/cloud-mta/internal/logs"
	"github.com/spf13/cobra"
	"github.com/SAP/cloud-mta/mta"
)

var createMtaCmdPath string
var createMtaCmdData string

func init() {

	// set flags of mtad command
	createMtaCmd.Flags().StringVarP(&createMtaCmdPath, "path", "p", "",
		"the path to the yaml file")
	createMtaCmd.Flags().StringVarP(&createMtaCmdData, "data", "d",
		"", "data in JSON format")

}

// AddModule Add new module
var createMtaCmd = &cobra.Command{
	Use:   "create",
	Short: "Create new MTA project",
	Long:  "Create new MTA project",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logs.Logger.Info("Create MTA project")
		err := mta.CreateMta(createMtaCmdPath, createMtaCmdData)
		if err != nil{
			logs.Logger.Error(err)
		}
		return err
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// AddModule Add new module
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

// AddResources - Add new resource
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
