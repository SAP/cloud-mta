package commands

import (
	"github.com/spf13/cobra"

	"github.com/SAP/cloud-mta/internal/logs"
	"github.com/SAP/cloud-mta/mta"
)

var addResourceMtaCmdPath string
var addResourceCmdData string
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
		logs.Logger.Info("add new resource")
		hash, err := mta.ModifyMta(addResourceMtaCmdPath, func() error {
			return mta.AddResource(addResourceMtaCmdPath, addResourceCmdData, mta.Marshal)
		}, addResourceCmdHashcode, false)
		writeErr := mta.WriteResult(nil, hash, err)
		if err != nil {
			// The original error is more important
			return err
		}
		return writeErr
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
		logs.Logger.Info("get resources")
		resources, err := mta.GetResources(getResourcesCmdPath)
		hash, _, hashErr := mta.GetMtaHash(getResourcesCmdPath)
		if err == nil && hashErr != nil {
			// Return an error if we couldn't get the hash
			err = hashErr
		}
		writeErr := mta.WriteResult(resources, hash, err)
		if err != nil {
			// The original error is more important
			return err
		}
		return writeErr
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
		logs.Logger.Info("update existing resource")
		hash, err := mta.ModifyMta(addResourceMtaCmdPath, func() error {
			return mta.UpdateResource(updateResourceMtaCmdPath, updateResourceCmdData, mta.Marshal)
		}, addResourceCmdHashcode, false)
		writeErr := mta.WriteResult(nil, hash, err)
		if err != nil {
			// The original error is more important
			return err
		}
		return writeErr
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}
