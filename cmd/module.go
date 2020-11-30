package commands

import (
	"github.com/spf13/cobra"

	"github.com/SAP/cloud-mta/mta"
)

var addModuleMtaCmdPath string
var addModuleCmdData string
var addModuleCmdForce bool
var addModuleCmdHashcode int
var getModulesCmdPath string
var getModulesCmdExtensions []string
var updateModuleMtaCmdPath string
var updateModuleCmdData string
var updateModuleCmdHashcode int

func init() {
	// Sets the flags of the commands.
	addModuleCmd.Flags().StringVarP(&addModuleMtaCmdPath, "path", "p", "",
		"the path to the yaml file")
	addModuleCmd.Flags().StringVarP(&addModuleCmdData, "data", "d", "",
		"data in JSON format")
	addModuleCmd.Flags().BoolVarP(&addModuleCmdForce, "force", "f", false,
		"force action")
	addModuleCmd.Flags().IntVarP(&addModuleCmdHashcode, "hashcode", "c", 0,
		"data hashcode")

	getModulesCmd.Flags().StringVarP(&getModulesCmdPath, "path", "p", "",
		"the path to the yaml file")
	getModulesCmd.Flags().StringSliceVarP(&getModulesCmdExtensions, "extensions", "x", nil,
		"the paths to the MTA extension descriptors")

	updateModuleCmd.Flags().StringVarP(&updateModuleMtaCmdPath, "path", "p", "",
		"the path to the yaml file")
	updateModuleCmd.Flags().StringVarP(&updateModuleCmdData, "data", "d", "",
		"data in JSON format")
	updateModuleCmd.Flags().IntVarP(&updateModuleCmdHashcode, "hashcode", "c", 0,
		"data hashcode")
}

// addModuleCmd - adds a new module.
var addModuleCmd = &cobra.Command{
	Use:   "module",
	Short: "Add new module",
	Long:  "Add new module",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mta.RunModifyAndWriteHash("add new module", addModuleMtaCmdPath, addModuleCmdForce, func() ([]string, error) {
			return mta.AddModule(addModuleMtaCmdPath, addModuleCmdData, mta.Marshal)
		}, addModuleCmdHashcode, false)
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// getModulesCmd - gets all the modules.
var getModulesCmd = &cobra.Command{
	Use:   "modules",
	Short: "Get all modules",
	Long:  "Get all modules",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mta.RunAndWriteResultAndHash("get modules", getModulesCmdPath, getModulesCmdExtensions, func() (interface{}, []string, error) {
			return mta.GetModules(getModulesCmdPath, getModulesCmdExtensions)
		})
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// updateModuleCmd - updates an existing module.
var updateModuleCmd = &cobra.Command{
	Use:   "module",
	Short: "Update existing module",
	Long:  "Update existing module",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mta.RunModifyAndWriteHash("update existing module", updateModuleMtaCmdPath, false, func() ([]string, error) {
			return mta.UpdateModule(updateModuleMtaCmdPath, updateModuleCmdData, mta.Marshal)
		}, updateModuleCmdHashcode, false)
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}
