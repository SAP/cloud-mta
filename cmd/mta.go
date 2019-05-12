package commands

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/SAP/cloud-mta/internal/logs"
	"github.com/SAP/cloud-mta/mta"
)

var createMtaCmdPath string
var createMtaCmdData string
var sourceCmdPath string
var targetCmdPath string
var deleteFileCmdPath string

func init() {

	// set flags of commands
	createMtaCmd.Flags().StringVarP(&createMtaCmdPath, "path", "p", "",
		"the path to the yaml file")
	createMtaCmd.Flags().StringVarP(&createMtaCmdData, "data", "d", "",
		"data in JSON format")
	copyCmd.Flags().StringVarP(&sourceCmdPath, "source", "s", "",
		"the path to the source file")
	copyCmd.Flags().StringVarP(&targetCmdPath, "target", "t", "",
		"the path to the target file")
	deleteFileCmd.Flags().StringVarP(&deleteFileCmdPath, "path", "p", "",
		"the path to the file")
}

// createMtaCmd Create new MTA project
var createMtaCmd = &cobra.Command{
	Use:   "create",
	Short: "Create new MTA project",
	Long:  "Create new MTA project",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logs.Logger.Info("create MTA project")
		err := mta.CreateMta(createMtaCmdPath, createMtaCmdData, os.MkdirAll)
		if err != nil {
			logs.Logger.Error(err)
		}
		return err
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// copyCmd copy from source path to target path
var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy from source path to target path",
	Long:  "Copy from source path to target path",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logs.Logger.Info("copy from source path to target path")
		err := mta.CopyFile(sourceCmdPath, targetCmdPath, os.Create)
		if err != nil {
			logs.Logger.Error(err)
		}
		return err
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// deleteFileCmd delete file in given path
var deleteFileCmd = &cobra.Command{
	Use:   "deleteFile",
	Short: "Delete file",
	Long:  "Delete file",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logs.Logger.Info("delete file in path: " + deleteFileCmdPath)
		err := mta.DeleteFile(deleteFileCmdPath)
		if err != nil {
			logs.Logger.Error(err)
		}
		return err
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}
