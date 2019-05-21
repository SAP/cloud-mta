package commands

import (
	"github.com/spf13/cobra"

	"github.com/SAP/cloud-mta/internal/logs"
	"github.com/SAP/cloud-mta/mta"
)

var addModuleMtaCmdPath string
var addModuleCmdData string
var addModuleCmdHashcode int
var getModulesCmdPath string
var updateModuleMtaCmdPath string
var updateModuleCmdData string
var updateModuleCmdHashcode int

func init() {
	// set flags of commands
	addModuleCmd.Flags().StringVarP(&addModuleMtaCmdPath, "path", "p", "",
		"the path to the yaml file")
	addModuleCmd.Flags().StringVarP(&addModuleCmdData, "data", "d", "",
		"data in JSON format")
	addModuleCmd.Flags().IntVarP(&addModuleCmdHashcode, "hashcode", "c", 0,
		"data hashcode")
	getModulesCmd.Flags().StringVarP(&getModulesCmdPath, "path", "p", "",
		"the path to the yaml file")
	updateModuleCmd.Flags().StringVarP(&updateModuleMtaCmdPath, "path", "p", "",
		"the path to the yaml file")
	updateModuleCmd.Flags().StringVarP(&updateModuleCmdData, "data", "d", "",
		"data in JSON format")
	updateModuleCmd.Flags().IntVarP(&updateModuleCmdHashcode, "hashcode", "c", 0,
		"data hashcode")
}

// addModuleCmd Add new module
var addModuleCmd = &cobra.Command{
	Use:   "module",
	Short: "Add new module",
	Long:  "Add new module",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logs.Logger.Info("add new module")
		hash, err := mta.ModifyMta(addModuleMtaCmdPath, func() error {
			return mta.AddModule(addModuleMtaCmdPath, addModuleCmdData, mta.Marshal)
		}, addModuleCmdHashcode, false)
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

// getModulesCmd Get all modules
var getModulesCmd = &cobra.Command{
	Use:   "modules",
	Short: "Get all modules",
	Long:  "Get all modules",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logs.Logger.Info("get modules")
		modules, err := mta.GetModules(getModulesCmdPath)
		hash, _, hashErr := mta.GetMtaHash(getModulesCmdPath)
		if err == nil && hashErr != nil {
			// Return an error if we couldn't get the hash
			err = hashErr
		}
		writeErr := mta.WriteResult(modules, hash, err)
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
var updateModuleCmd = &cobra.Command{
	Use:   "module",
	Short: "Update existing module",
	Long:  "Update existing module",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logs.Logger.Info("update existing module")
		hash, err := mta.ModifyMta(addModuleMtaCmdPath, func() error {
			return mta.UpdateModule(updateModuleMtaCmdPath, updateModuleCmdData, mta.Marshal)
		}, updateModuleCmdHashcode, false)
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
