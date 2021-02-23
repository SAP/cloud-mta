package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/SAP/cloud-mta/internal/logs"
	"github.com/SAP/cloud-mta/mta"
	"github.com/SAP/cloud-mta/validations"
)

var createMtaCmdPath string
var createMtaCmdData string
var deleteMtaCmdPath string
var copyCmdSourcePath string
var copyCmdTargetPath string
var deleteFileCmdPath string
var existCmdName string
var existCmdPath string
var getBuildParametersCmdPath string
var getBuildParametersCmdExtensions []string
var getParametersCmdPath string
var getParametersCmdExtensions []string
var updateBuildParametersCmdPath string
var updateBuildParametersCmdData string
var updateBuildParametersCmdForce bool
var updateBuildParametersCmdHashcode int
var updateParametersCmdPath string
var updateParametersCmdData string
var updateParametersCmdHashcode int
var getMtaIDCmdPath string
var validateMtaCmdPath string
var validateMtaCmdExtensions []string

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

	getBuildParametersCmd.Flags().StringVarP(&getBuildParametersCmdPath, "path", "p", "",
		"the path to the yaml file")
	getBuildParametersCmd.Flags().StringSliceVarP(&getBuildParametersCmdExtensions, "extensions", "x", nil,
		"the paths to the MTA extension descriptors")

	getParametersCmd.Flags().StringVarP(&getParametersCmdPath, "path", "p", "",
		"the path to the yaml file")
	getParametersCmd.Flags().StringSliceVarP(&getParametersCmdExtensions, "extensions", "x", nil,
		"the paths to the MTA extension descriptors")

	updateBuildParametersCmd.Flags().StringVarP(&updateBuildParametersCmdPath, "path", "p", "",
		"the path to the file")
	updateBuildParametersCmd.Flags().StringVarP(&updateBuildParametersCmdData, "data", "d", "",
		"data in JSON format")
	updateBuildParametersCmd.Flags().BoolVarP(&updateBuildParametersCmdForce, "force", "f", false,
		"force action")
	updateBuildParametersCmd.Flags().IntVarP(&updateBuildParametersCmdHashcode, "hashcode", "c", 0,
		"data hashcode")

	updateParametersCmd.Flags().StringVarP(&updateParametersCmdPath, "path", "p", "",
		"the path to the file")
	updateParametersCmd.Flags().StringVarP(&updateParametersCmdData, "data", "d", "",
		"data in JSON format")
	updateParametersCmd.Flags().IntVarP(&updateParametersCmdHashcode, "hashcode", "c", 0,
		"data hashcode")

	getMtaIDCmd.Flags().StringVarP(&getMtaIDCmdPath, "path", "p", "",
		"the path to the file")

	validateMtaCmd.Flags().StringVarP(&validateMtaCmdPath, "path", "p", "",
		"the path to the yaml file")
	validateMtaCmd.Flags().StringSliceVarP(&validateMtaCmdExtensions, "extensions", "x", nil,
		"the paths to the MTA extension descriptors")

}

// createMtaCmd Create new MTA project
var createMtaCmd = &cobra.Command{
	Use:   "create",
	Short: "Create new MTA project",
	Long:  "Create new MTA project",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mta.RunModifyAndWriteHash("create MTA project", createMtaCmdPath, false, func() ([]string, error) {
			return nil, mta.CreateMta(createMtaCmdPath, createMtaCmdData, os.MkdirAll)
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
		writeErr := mta.WriteResult(nil, nil, 0, err)
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
			nil, // Extensions are not relevant in copy
			func() (interface{}, []string, error) {
				return nil, nil, mta.CopyFile(copyCmdSourcePath, copyCmdTargetPath, os.Create)
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
		writeErr := mta.WriteResult(nil, nil, 0, err)
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
			nil, // Extensions don't change, remove or add names
			func() (interface{}, []string, error) {
				return mta.IsNameUnique(existCmdPath, existCmdName)
			},
		)
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// getBuildParametersCmd get build parameters from mta
var getBuildParametersCmd = &cobra.Command{
	Use:   "buildParameters",
	Short: "Get build parameters",
	Long:  "Get build parameters",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mta.RunAndWriteResultAndHash("get build parameters", getBuildParametersCmdPath, getBuildParametersCmdExtensions, func() (interface{}, []string, error) {
			return mta.GetBuildParameters(getBuildParametersCmdPath, getBuildParametersCmdExtensions)
		})
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// getParametersCmd get parameters from mta
var getParametersCmd = &cobra.Command{
	Use:   "parameters",
	Short: "Get parameters",
	Long:  "Get parameters",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mta.RunAndWriteResultAndHash("get parameters", getParametersCmdPath, getParametersCmdExtensions, func() (interface{}, []string, error) {
			return mta.GetParameters(getParametersCmdPath, getParametersCmdExtensions)
		})
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
		return mta.RunModifyAndWriteHash("update build parameters", updateBuildParametersCmdPath, updateBuildParametersCmdForce, func() ([]string, error) {
			return mta.UpdateBuildParameters(updateBuildParametersCmdPath, updateBuildParametersCmdData)
		}, updateBuildParametersCmdHashcode, false)
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// updateParametersCmd update parameters in mta
var updateParametersCmd = &cobra.Command{
	Use:   "parameters",
	Short: "Update parameters",
	Long:  "Update parameters",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mta.RunModifyAndWriteHash("update parameters", updateParametersCmdPath, false, func() ([]string, error) {
			return mta.UpdateParameters(updateParametersCmdPath, updateParametersCmdData)
		}, updateParametersCmdHashcode, false)
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
		// Extensions don't change the mta ID
		return mta.RunAndWriteResultAndHash("get MTA ID", getMtaIDCmdPath, nil, func() (interface{}, []string, error) {
			return mta.GetMtaID(getMtaIDCmdPath)
		})
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// validateMtaCmd - validate mta.yaml file and its extensions
var validateMtaCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate MTA",
	Long:  "Validate mta.yaml file and MTA extension files",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Extensions don't change the mta ID
		return mta.RunAndWriteResultAndHash("validate MTA", validateMtaCmdPath, validateMtaCmdExtensions, func() (interface{}, []string, error) {
			return validate.Validate(validateMtaCmdPath, validateMtaCmdExtensions), nil, nil
		})
	},
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
}
