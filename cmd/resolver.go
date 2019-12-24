package commands

import (
	"github.com/SAP/cloud-mta/internal/logs"
	"github.com/SAP/cloud-mta/internal/resolver"
	"github.com/spf13/cobra"
)

var resolvePath string
var workspaceDir string
var resolveModule string
var resolveEnvFileName string

func init() {
	resolveMtaCmd.Flags().StringVarP(&resolvePath, "path", "p", "",
		"the path to the yaml file")
	resolveMtaCmd.Flags().StringVarP(&workspaceDir, "workspace", "w", "",
		"the path to workspace-folder")
	resolveMtaCmd.Flags().StringVarP(&resolveModule, "module", "m", "",
		"module-name")
	resolveMtaCmd.Flags().StringVarP(&resolveEnvFileName, "envFile", "e", "",
		"the environment file name. The default file name is .env")

}

// createMtaCmd Create new MTA project
var resolveMtaCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Resolve variables and placeholders in an MTA file",
	Long: `MTA file typically contains variables in the form ~{var-name} and placeholders in the form ${placeholder}, 
resolve command print to stdout the MTA fil contents with as much as possible variables and placeholders replaced 
with concrete values, based on environment variables provided and environment files in the modules' folders`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logs.Logger.Info("Resolve MTA")
		err := resolver.Resolve(workspaceDir, resolveModule, resolvePath, resolveEnvFileName)
		if err != nil {
			logs.Logger.Error(err)
		}
		return err
	},
	Hidden:        false,
	SilenceUsage:  true,
	SilenceErrors: true,
}
