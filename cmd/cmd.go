package commands

import (
	"github.com/spf13/cobra"
)

func init() {

	rootCmd.Flags().BoolP("version", "v", false, "version for MTA")

	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(createMtaCmd)
	rootCmd.AddCommand(copyCmd)
	rootCmd.AddCommand(deleteFileCmd)
	rootCmd.AddCommand(existCmd)
	addCmd.AddCommand(addModuleCmd, addResourceCmd)
	getCmd.AddCommand(getModulesCmd, getResourcesCmd)
	updateCmd.AddCommand(updateModuleCmd, updateResourceCmd, updateBuildParametersCmd)
}

// Parent command add any artifacts
var addCmd = &cobra.Command{
	Use:    "add",
	Short:  "Add artifacts",
	Long:   "Add artifacts",
	Hidden: true,
	Run:    nil,
}

// Parent command get any artifacts
var getCmd = &cobra.Command{
	Use:    "get",
	Short:  "Get artifacts",
	Long:   "Get artifacts",
	Hidden: true,
	Run:    nil,
}

// Parent command to update artifacts
var updateCmd = &cobra.Command{
	Use:    "update",
	Short:  "Update artifact",
	Long:   "Update artifact",
	Hidden: true,
	Run:    nil,
}
