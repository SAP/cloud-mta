package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/SAP/cloud-mta/internal/logs"
	"github.com/SAP/cloud-mta/mta"
)

var createMtaCmdPath string
var createMtaCmdData string
var deleteMtaCmdPath string
var copyCmdSourcePath string
var copyCmdTargetPath string
var deleteFileCmdPath string
var existCmdName string
var existCmdPath string
var updateBuildParametersCmdPath string
var updateBuildParametersCmdData string
var updateBuildParametersCmdForce bool
var updateBuildParametersCmdHashcode int
var getMtaIDCmdPath string

func init() {

	// set flags of commands
	createMtaCmd.Flags().StringVarP(&createMtaCmdPath, "path", "p", "",
		"the path to the yaml file")
	createMtaCmd.Flags().StringVarP(&createMtaCmdData, "data", "d", "",
		"data in JSON format")
	deleteMtaCmd.Flags().StringVarP(&deleteMtaCmdPath, "path", "p", "",
		"the path to the MTA project")
	copyCmd.Flags().StringVarP(&copyCmdSourcePath, "source", "s", "",
		"the path to the source file")
	copyCmd.Flags().StringVarP(&copyCmdTargetPath, "target", "t", "",
		"the path to the target file")
	deleteFileCmd.Flags().StringVarP(&deleteFileCmdPath, "path", "p", "",
		"the path to the file")
	existCmd.Flags().StringVarP(&existCmdPath, "path", "p", "",
		"the path to the file")
	existCmd.Flags().StringVarP(&existCmdName, "name", "n", "",
		"the name to check")
	updateBuildParametersCmd.Flags().StringVarP(&updateBuildParametersCmdPath, "path", "p", "",
		"the path to the file")
	updateBuildParametersCmd.Flags().StringVarP(&updateBuildParametersCmdData, "data", "d", "",
		"data in JSON format")
	updateBuildParametersCmd.Flags().BoolVarP(&updateBuildParametersCmdForce, "force", "f", false,
		"force action")
	updateBuildParametersCmd.Flags().IntVarP(&updateBuildParametersCmdHashcode, "hashcode", "c", 0,
		"data hashcode")
	getMtaIDCmd.Flags().StringVarP(&getMtaIDCmdPath, "path", "p", "",
		"the path to the file")
}

// createMtaCmd Create new MTA project
var createMtaCmd = &cobra.Command{
	Use:   "create",
	Short: "Create new MTA project",
	Long:  "Create new MTA project",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mta.RunModifyAndWriteHash("create MTA project", createMtaCmdPath, false, func() error {
			return mta.CreateMta(createMtaCmdPath, createMtaCmdData, os.MkdirAll)
		}, 0, true)
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// deleteMtaCmd Delete MTA project
var deleteMtaCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete MTA project",
	Long:  "Delete MTA project",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logs.Logger.Info("delete MTA in path: " + deleteMtaCmdPath)
		err := mta.DeleteMta(deleteMtaCmdPath)
		writeErr := mta.WriteResult(nil, 0, err)
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

// copyCmd copy from source path to target path
var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy from source path to target path",
	Long:  "Copy from source path to target path",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mta.RunAndWriteResultAndHash(
			fmt.Sprintf("copy from source path: %s to target path: %s", copyCmdSourcePath, copyCmdTargetPath),
			copyCmdTargetPath,
			func() (interface{}, error) {
				return nil, mta.CopyFile(copyCmdSourcePath, copyCmdTargetPath, os.Create)
			},
		)
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
		writeErr := mta.WriteResult(nil, 0, err)
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

// existCmd check if name exists in file
var existCmd = &cobra.Command{
	Use:   "exist",
	Short: "Check exists",
	Long:  "Check exists",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mta.RunAndWriteResultAndHash(
			fmt.Sprintf("check if name %s exists in %s file", existCmdName, existCmdPath),
			existCmdPath,
			func() (interface{}, error) {
				return mta.IsNameUnique(existCmdPath, existCmdName)
			},
		)
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// updateBuildParametersCmd update build parameters in mta
var updateBuildParametersCmd = &cobra.Command{
	Use:   "buildParameters",
	Short: "Update build parameters",
	Long:  "Update build parameters",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mta.RunModifyAndWriteHash("update build parameters", updateBuildParametersCmdPath, updateBuildParametersCmdForce, func() error {
			return mta.UpdateBuildParameters(updateBuildParametersCmdPath, updateBuildParametersCmdData)
		}, updateBuildParametersCmdHashcode, false)
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// getMtaIDCmd - get MTA ID.
var getMtaIDCmd = &cobra.Command{
	Use:   "id",
	Short: "Get MTA ID",
	Long:  "Get MTA ID",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mta.RunAndWriteResultAndHash("get MTA ID", getMtaIDCmdPath, func() (interface{}, error) {
			return mta.GetMtaID(getMtaIDCmdPath)
		})
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}
