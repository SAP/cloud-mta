package commands

import (
	"fmt"
	"github.com/SAP/cloud-mta/internal/logs"
	"github.com/SAP/cloud-mta/internal/resolver"
	"github.com/SAP/cloud-mta/mta"
	"github.com/spf13/cobra"
)

var resolveCmdPath string
var resolveCmdExtensions []string
var resolveCmdWorkspaceDir string
var resolveCmdModule string
var resolveCmdEnvFileName string
var resolveCmdOutputFormat string

func init() {
	resolveMtaCmd.Flags().StringVarP(&resolveCmdPath, "path", "p", "",
		"the path to the mta.yaml file")
	resolveMtaCmd.Flags().StringSliceVarP(&resolveCmdExtensions, "extensions", "x", nil,
		"the paths to the MTA extension descriptors")
	resolveMtaCmd.Flags().StringVarP(&resolveCmdWorkspaceDir, "workspace", "w", "",
		"the path to the project folder; the default path is the folder of the mta.yaml file")
	resolveMtaCmd.Flags().StringVarP(&resolveCmdModule, "module", "m", "",
		"the module name")
	resolveMtaCmd.Flags().StringVarP(&resolveCmdEnvFileName, "envFile", "e", "",
		"the environment file path, relative to the module folder; the default file path is \".env\"")
	resolveMtaCmd.Flags().StringVarP(&resolveCmdOutputFormat, "output", "o", "",
		"the output format; use \"json\" for json-formatted output")
	_ = resolveMtaCmd.Flags().MarkHidden("output")

}

// createMtaCmd Create new MTA project
var resolveMtaCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Resolve variables and placeholders in an MTA file",
	Long: `The MTA file typically contains variables in the form ~{var-name} and placeholders in the form ${placeholder}.
The resolve command prints the module's properties from the MTA file to stdout, with variables and placeholders replaced with concrete values, based on environment variables and an environment file.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if resolveCmdOutputFormat == "json" {
			return mta.RunAndWriteResultAndHash(
				"Resolve MTA",
				resolveCmdPath,
				resolveCmdExtensions,
				func() (interface{}, []string, error) {
					return resolver.Resolve(resolveCmdWorkspaceDir, resolveCmdModule, resolveCmdPath, resolveCmdExtensions, resolveCmdEnvFileName)
				},
			)
		}

		// Just write to the output (this option is here for backwards compatibility)
		logs.Logger.Info("Resolve MTA")
		result, messages, err := resolver.Resolve(resolveCmdWorkspaceDir, resolveCmdModule, resolveCmdPath, resolveCmdExtensions, resolveCmdEnvFileName)
		if err != nil {
			logs.Logger.Error(err)
		} else {
			for key, val := range result.Properties {
				fmt.Println(key + "=" + val)
			}
			for _, message := range messages {
				logs.Logger.Warn(message)
			}
			for _, message := range result.Messages {
				logs.Logger.Warn(message)
			}
		}
		return err
	},
	Hidden:        false,
	SilenceUsage:  true,
	SilenceErrors: true,
}
