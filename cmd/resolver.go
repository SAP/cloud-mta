package commands

import (
	"github.com/SAP/cloud-mta/internal/logs"
	"github.com/SAP/cloud-mta/internal/resolver"
	"github.com/spf13/cobra"
)

var resolvePath string
var resolveWorkspaceDir string
var resolveModule string
var resolveEnvFileName string

func init() {
	resolveMtaCmd.Flags().StringVarP(&resolvePath, "path", "p", "",
		"the path to the mta.yaml file")
	resolveMtaCmd.Flags().StringVarP(&resolveWorkspaceDir, "workspace", "w", "",
		"the path to the project folder; the default path is the folder of the mta.yaml file")
	resolveMtaCmd.Flags().StringVarP(&resolveModule, "module", "m", "",
		"the module name")
	resolveMtaCmd.Flags().StringVarP(&resolveEnvFileName, "envFile", "e", "",
		"the environment file path, relative to the module folder; the default file path is \".env\"")

}

// createMtaCmd Create new MTA project
var resolveMtaCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Resolve variables and placeholders in an MTA file",
	Long: `The MTA file typically contains variables in the form ~{var-name} and placeholders in the form ${placeholder}.
The resolve command prints the module's properties from the MTA file to stdout, with variables and placeholders replaced with concrete values, based on environment variables and an environment file.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logs.Logger.Info("Resolve MTA")
		err := resolver.Resolve(resolveWorkspaceDir, resolveModule, resolvePath, resolveEnvFileName)
		if err != nil {
			logs.Logger.Error(err)
		}
		return err
	},
	Hidden:        false,
	SilenceUsage:  true,
	SilenceErrors: true,
}
