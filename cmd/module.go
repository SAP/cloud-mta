package commands

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"github.com/SAP/cloud-mta/internal/logs"
	"github.com/SAP/cloud-mta/mta"
)

var addModuleMtaCmdPath string
var addModuleCmdData string
var getModulesCmdPath string

func init() {
	// set flags of commands
	addModuleCmd.Flags().StringVarP(&addModuleMtaCmdPath, "path", "p", "",
		"the path to the yaml file")
	addModuleCmd.Flags().StringVarP(&addModuleCmdData, "data", "d", "",
		"data in JSON format")
	getModulesCmd.Flags().StringVarP(&getModulesCmdPath, "path", "p", "",
		"the path to the yaml file")
}

// addModuleCmd Add new module
var addModuleCmd = &cobra.Command{
	Use:   "module",
	Short: "Add new module",
	Long:  "Add new module",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logs.Logger.Info("add new module")
		err := mta.AddModule(addModuleMtaCmdPath, addModuleCmdData, yaml.Marshal)
		if err != nil {
			logs.Logger.Error(err)
		}
		return err
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
		if err != nil {
			logs.Logger.Error(err)
		}
		if modules != nil {
			fmt.Print(string(modules))
		}
		return err
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}
