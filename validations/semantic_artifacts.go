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
		// no path check for modules with build parameter "no-source" set to true
		noSource, issue := ifNoSource(module, modulesNode, index)
		if issue != nil {
			issues = append(issues, *issue)
		} else if noSource {
			continue
		}

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
			line, propFound := getIndexedNodePropLine(modulesNode, index, pathYamlField)
			if !propFound {
				line, _ = getIndexedNodePropLine(modulesNode, index, nameYamlField)
			}
			// path not exists -> add an issue
			issues = appendIssue(issues, fmt.Sprintf(`the "%s" path of the "%s" module does not exist`,
				modulePath, module.Name), line)
		}
	}

	return issues, nil
}

func ifNoSource(module *mta.Module, modulesNode *yaml.Node, index int) (bool, *YamlValidationIssue) {
	if module.BuildParams != nil && module.BuildParams[noSourceYamlField] != nil {
		moduleNode := modulesNode.Content[index]
		buildParametersNode := getPropValueByName(moduleNode, buildParametersYamlField)
		noSourceNode := getPropValueByName(buildParametersNode, noSourceYamlField)
		noSource, ok := module.BuildParams[noSourceYamlField].(bool)
		if ok {
			return noSource, nil
		}
		return false, &YamlValidationIssue{
			Msg:  `the "no-source" build parameter must be a boolean`,
			Line: noSourceNode.Line,
		}
	}
	return false, nil
}
