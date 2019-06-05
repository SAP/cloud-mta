package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/SAP/cloud-mta/internal/logs"
	"github.com/SAP/cloud-mta/mta"
	"github.com/spf13/cobra"
)

var resolvePath string
var workspaceDir string
var resolveModule string

func init() {
	resolveMtaCmd.Flags().StringVarP(&resolvePath, "path", "p", "",
		"the path to the yaml file")
	resolveMtaCmd.Flags().StringVarP(&workspaceDir, "workspace", "w", "",
		"the path to workspace-folder")
	resolveMtaCmd.Flags().StringVarP(&resolveModule, "module", "m", "",
		"module-name")

}

// createMtaCmd Create new MTA project
var resolveMtaCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Resolve variables and placeholders in an MTA file",
	Long: `MTA file typically contains variables in the form ~{var-name} and placeholders in the form ${placeholder}, 
resolve command print to stdout the MTA fil contents with as much as possible variables and placeholders replaced 
with concrete values, based on environment variables provided and .env files in the modules' folders`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logs.Logger.Info("Resolve MTA")
		err := resolve(resolvePath)
		if err != nil {
			logs.Logger.Error(err)
		}
		return err
	},
	Hidden:        false,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func resolve(mtaPath string) error {
	if len(resolveModule) == 0 {
		return errors.New("You must provide module name")
	}
	yamlData, err := ioutil.ReadFile(mtaPath)
	if err != nil {
		return err
	}
	mtaRaw, err := mta.Unmarshal(yamlData)
	if err != nil {
		return err
	}

	if len(workspaceDir) == 0 {
		workspaceDir = path.Dir(mtaPath)
	}
	m := NewMTAResolver(mtaRaw, workspaceDir)

	for _, module := range m.GetModules() {
		if module.Name == resolveModule {
			m.ResolveProperies(module)

			for key, val := range getPropertiesAsEnvVar(module) {
				fmt.Println(key + "=" + val)
			}
			return nil
		}
	}

	return errors.New("module not find")
}

func getPropertiesAsEnvVar(module *mta.Module) map[string]string {
	envVar := map[string]interface{}{}
	for key, val := range module.Properties {
		envVar[key] = val
	}

	for _, requires := range module.Requires {
		propMap := envVar
		if len(requires.Group) > 0 {
			propMap = map[string]interface{}{}
		}

		for key, val := range requires.Properties {
			propMap[key] = val
		}

		if len(requires.Group) > 0 {
			//append the array element to group
			group, ok := envVar[requires.Group]
			if ok {
				groupArray := group.([]map[string]interface{})
				envVar[requires.Group] = append(groupArray, propMap)
			} else {
				envVar[requires.Group] = []map[string]interface{}{propMap}
			}
		}
	}

	//serialize
	retEnvVar := map[string]string{}
	for key, val := range envVar {
		switch v := val.(type) {
		case string:
			retEnvVar[key] = v
		default:
			bytesVal, err := json.Marshal(val)
			if err != nil {
				//todo log
			}
			retEnvVar[key] = string(bytesVal)
		}
	}

	return retEnvVar
}
