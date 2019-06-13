package commands

import (
	"github.com/SAP/cloud-mta/internal/logs"
	"github.com/SAP/cloud-mta/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/x-cray/logrus-prefixed-formatter"
)

var cfgFile string

func init() {
	logs.Logger = logs.NewLogger()
	formatter, ok := logs.Logger.Formatter.(*prefixed.TextFormatter)
	if ok {
		formatter.DisableColors = true
	}
	cobra.OnInitialize(initConfig)
}

// rootCmd - represents the base command.
var rootCmd = &cobra.Command{
	Use:     "MTA",
	Short:   "MTA tools",
	Long:    "MTA tools",
	Version: cliVersion(),
	Args:    cobra.MaximumNArgs(1),
}

// Execute - adds all the child commands to the root command and sets the flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv() // Reads in the environment variables that match.
	// If a configs file is found, reads it in.
	if err := viper.ReadInConfig(); err == nil {
		logs.Logger.Println("Using configs file:", viper.ConfigFileUsed())
	}
}

func cliVersion() string {
	v, _ := version.GetVersion()
	return v.CliVersion
}
