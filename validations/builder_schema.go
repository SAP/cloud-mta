package validate

import (
	"fmt"
	"gopkg.in/yaml.v3"

	"github.com/SAP/cloud-mta/mta"
)

func checkStringParam(params map[string]interface{}, parentNode *yaml.Node, paramsListName string, paramName string) []YamlValidationIssue {
	var issues []YamlValidationIssue

	if params != nil && params[paramName] != nil {
		paramsNode := getPropValueByName(parentNode, paramsListName)
		_, ok := params[paramName].(string)
		if !ok {
			paramNode := getPropValueByName(paramsNode, paramName)
			issues = appendIssue(issues,
				fmt.Sprintf(`the "%s" property is defined incorrectly; the property must be a string`, paramName),
				paramNode.Line)
		}
	}
	return issues
}

func checkBuilderSchema(mta *mta.MTA, mtaNode *yaml.Node, source string) []YamlValidationIssue {
	var issues []YamlValidationIssue

	issues = append(issues, checkStringParam(mta.Parameters, mtaNode, parametersYamlField, deployModeYamlField)...)

	modulesNode := getPropContent(mtaNode, modulesYamlField)

	for i, module := range mta.Modules {
		issues = append(issues, checkStringParam(module.BuildParams, modulesNode[i], buildParametersYamlField, builderYamlField)...)
		if module.BuildParams != nil && module.BuildParams[commandsYamlField] != nil {
			// check that "commands" fields is a sequence of strings
			_, ok := module.BuildParams[commandsYamlField].([]string)
			if !ok {
				// sequence of interfaces has to be convertible to the sequence of strings
				commands, okI := module.BuildParams[commandsYamlField].([]interface{})
				if okI {
					ok = ifCommandsStrings(commands)
				}
			}
			if !ok {
				buildParamsNode := getPropValueByName(modulesNode[i], buildParametersYamlField)
				commandsParamsNode := getPropValueByName(buildParamsNode, commandsYamlField)
				issues = appendIssue(issues, `the "commands" property is defined incorrectly; the property must be a sequence of strings`, commandsParamsNode.Line)
			}
		}
	}

	return issues
}

func ifCommandsStrings(commands []interface{}) bool {
	for _, commandI := range commands {
		_, ok := commandI.(string)
		if !ok {
			return false
		}
	}
	return true
}
