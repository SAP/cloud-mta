package commands

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"

	"github.com/SAP/cloud-mta/internal/logs"
	"github.com/SAP/cloud-mta/mta"
)

var addResourceMtaCmdPath string
var addResourceCmdData string
var addResourceCmdHashcode int
var getResourcesCmdPath string

func init() {
	// set flags of commands
	addResourceCmd.Flags().StringVarP(&addResourceMtaCmdPath, "path", "p", "",
		"the path to the yaml file")
	addResourceCmd.Flags().StringVarP(&addResourceCmdData, "data", "d", "",
		"data in JSON format")
	addResourceCmd.Flags().IntVarP(&addResourceCmdHashcode, "hashcode", "h", 0,
		"data hashcode")
	getResourcesCmd.Flags().StringVarP(&getResourcesCmdPath, "path", "p", "",
		"the path to the yaml file")
}

// addResourceCmd - Add new resource
var addResourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Add new resources",
	Long:  "Add new resources",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logs.Logger.Info("add new resource")
		err := mta.ModifyMta(addResourceMtaCmdPath, func() error {
			return mta.AddResource(addResourceMtaCmdPath, addResourceCmdData, yaml.Marshal)
		}, addResourceCmdHashcode, false)
		if err != nil {
			logs.Logger.Error(err)
		}
		return err
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
		if err != nil {
			logs.Logger.Error(err)
		}
		if resources != nil {
			output, rerr := json.Marshal(resources)
			if rerr != nil {
				logs.Logger.Error(rerr)
				return rerr
			}
			fmt.Print(string(output))
		}
		return err
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}
