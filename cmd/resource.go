package commands

import (
	"github.com/SAP/cloud-mta/internal/logs"
	"github.com/SAP/cloud-mta/mta"
	"github.com/spf13/cobra"
)

var addResourceMtaCmdPath string
var addResourceCmdData string

func init() {
	// set flags of commands
	addResourceCmd.Flags().StringVarP(&addResourceMtaCmdPath, "path", "p", "",
		"the path to the yaml file")
	addResourceCmd.Flags().StringVarP(&addResourceCmdData, "data", "d", "",
		"data in JSON format")
}

// addResourceCmd - Add new resource
var addResourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Add new resources",
	Long:  "Add new resources",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logs.Logger.Info("Adding new resource")
		err := mta.AddResource(addResourceMtaCmdPath, addResourceCmdData)
		if err != nil {
			logs.Logger.Error(err)
		}
		return err
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}
