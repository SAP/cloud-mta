package validate

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"

	"github.com/SAP/cloud-mta/mta"
)

// ifNoSourceParamBool - validates that "no-source" build parameter is boolean if defined
func ifNoSourceParamBool(mta *mta.MTA, mtaNode *yaml.Node, source string, strict bool) ([]YamlValidationIssue, []YamlValidationIssue) {
	var issues []YamlValidationIssue

	modulesNode := getPropValueByName(mtaNode, modulesYamlField)
	for index, module := range mta.Modules {
		_, issue := ifNoSource(module, modulesNode, index)
		if issue != nil {
			issues = append(issues, *issue)
		}
	}
	return issues, nil
}

// ifModulePathExists - validates the existence of modules paths used in the MTA descriptor
func ifModulePathExists(mta *mta.MTA, mtaNode *yaml.Node, source string, strict bool) ([]YamlValidationIssue, []YamlValidationIssue) {
	var issues []YamlValidationIssue

	modulesNode := getPropValueByName(mtaNode, modulesYamlField)
	for index, module := range mta.Modules {
		// no path check for modules with build parameter "no-source" set to true
		noSource, _ := ifNoSource(module, modulesNode, index)
		if !noSource && module.Path != "" {
			// build full path
			fullPath := filepath.Join(source, module.Path)
			// check existence of file/folder
			_, err := os.Stat(fullPath)
			if err != nil {
				line, column, _ := getIndexedNodePropPosition(modulesNode, index, pathYamlField)
				// path not exists -> add an issue
				issues = appendIssue(issues, fmt.Sprintf(`the "%s" path of the "%s" module does not exist`,
					module.Path, module.Name), line, column)
			}
		}
	}

	return issues, nil
}

// ifModulePathEmpty - validates that path is defined in module if no-source configuration is not used
func ifModulePathEmpty(mta *mta.MTA, mtaNode *yaml.Node, source string, strict bool) ([]YamlValidationIssue, []YamlValidationIssue) {
	var issues []YamlValidationIssue

	modulesNode := getPropValueByName(mtaNode, modulesYamlField)
	for index, module := range mta.Modules {
		// no path check for modules with build parameter "no-source" set to true
		noSource, _ := ifNoSource(module, modulesNode, index)
		if !noSource && module.Path == "" {
			moduleNode := modulesNode.Content[index]
			pathNode := getPropValueByName(moduleNode, pathYamlField)
			if pathNode == nil {
				issues = appendIssue(issues, fmt.Sprintf(`the path of the "%s" module is not defined`,
					module.Name), moduleNode.Line, moduleNode.Column)
			} else {
				issues = appendIssue(issues, fmt.Sprintf(`the path of the "%s" module is empty`,
					module.Name), pathNode.Line, pathNode.Column)
			}
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
			Msg:    `the "no-source" build parameter must be a boolean`,
			Line:   noSourceNode.Line,
			Column: noSourceNode.Column,
		}
	}
	return false, nil
}
