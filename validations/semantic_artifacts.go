package validate

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"

	"github.com/SAP/cloud-mta/mta"
)

// ifModulePathExists - validates the existence of modules paths used in the MTA descriptor
func ifModulePathExists(mta *mta.MTA, mtaNode *yaml.Node, source string, strict bool) ([]YamlValidationIssue, []YamlValidationIssue) {
	var issues []YamlValidationIssue

	modulesNode := getPropValueByName(mtaNode, modulesYamlField)
	for index, module := range mta.Modules {
		modulePath := module.Path
		// "path" property not defined -> use module name as a path
		if modulePath == "" {
			modulePath = module.Name
		}
		// build full path
		fullPath := filepath.Join(source, modulePath)
		// check existence of file/folder
		_, err := os.Stat(fullPath)
		if err != nil {
			line := getIndexedNodePropLine(modulesNode, index, pathYamlField)
			if line == 0 {
				line = getIndexedNodePropLine(modulesNode, index, nameYamlField)
			}
			// path not exists -> add an issue
			issues = appendIssue(issues, fmt.Sprintf(`the "%s" path of the "%s" module does not exist`,
				modulePath, module.Name), line)
		}
	}

	return issues, nil
}
