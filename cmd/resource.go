package commands

import (
	"github.com/spf13/cobra"

	"github.com/SAP/cloud-mta/mta"
)

var addResourceMtaCmdPath string
var addResourceCmdData string
var addResourceCmdForce bool
var addResourceCmdHashcode int
var getResourcesCmdPath string
var updateResourceMtaCmdPath string
var updateResourceCmdData string
var updateResourceCmdHashcode int

func init() {
	// set flags of commands
	addResourceCmd.Flags().StringVarP(&addResourceMtaCmdPath, "path", "p", "",
		"the path to the yaml file")
	addResourceCmd.Flags().StringVarP(&addResourceCmdData, "data", "d", "",
		"data in JSON format")
	addResourceCmd.Flags().BoolVarP(&addResourceCmdForce, "force", "f", false,
		"force action")
	addResourceCmd.Flags().IntVarP(&addResourceCmdHashcode, "hashcode", "c", 0,
		"data hashcode")
	getResourcesCmd.Flags().StringVarP(&getResourcesCmdPath, "path", "p", "",
		"the path to the yaml file")
	updateResourceCmd.Flags().StringVarP(&updateResourceMtaCmdPath, "path", "p", "",
		"the path to the yaml file")
	updateResourceCmd.Flags().StringVarP(&updateResourceCmdData, "data", "d", "",
		"data in JSON format")
	updateResourceCmd.Flags().IntVarP(&updateResourceCmdHashcode, "hashcode", "c", 0,
		"data hashcode")
}

// addResourceCmd - Add new resource
var addResourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Add new resources",
	Long:  "Add new resources",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mta.RunModifyAndWriteHash("add new resource", addResourceMtaCmdPath, addResourceCmdForce, func() error {
			return mta.AddResource(addResourceMtaCmdPath, addResourceCmdData, mta.Marshal)
		}, addResourceCmdHashcode, false)
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// getResourcesCmd - Get all resources
var getResourcesCmd = &cobra.Command{
	Use:   "resources",
	Short: "Get all resources",
	Long:  "Get all resources",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mta.RunAndWriteResultAndHash("get resources", getResourcesCmdPath, func() (interface{}, error) {
			return mta.GetResources(getResourcesCmdPath)
		})
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// updateModuleCmd updates an existing module
var updateResourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Update existing resource",
	Long:  "Update existing resource",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mta.RunModifyAndWriteHash("update existing resource", updateResourceMtaCmdPath, false, func() error {
			return mta.UpdateResource(updateResourceMtaCmdPath, updateResourceCmdData, mta.Marshal)
		}, updateResourceCmdHashcode, false)
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}
