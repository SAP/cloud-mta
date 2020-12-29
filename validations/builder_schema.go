package validate

import (
	"fmt"
	"gopkg.in/yaml.v3"

	"github.com/SAP/cloud-mta/mta"
)

func checkStringProperty(props map[string]interface{}, propsNode *yaml.Node, propName string) []YamlValidationIssue {

	_, ok := props[propName].(string)
	if props[propName] != nil && !ok {
		propNode := getPropValueByName(propsNode, propName)
		return []YamlValidationIssue{
			{
				Msg:    fmt.Sprintf(`the "%s" property is defined incorrectly; the property must be a string`, propName),
				Line:   propNode.Line,
				Column: propNode.Column,
			},
		}
	}
	return nil
}

func checkBuilderSchema(mta *mta.MTA, mtaNode *yaml.Node, source string) []YamlValidationIssue {
	var issues []YamlValidationIssue

	issues = append(issues, checkStringProperty(mta.Parameters, getPropValueByName(mtaNode, parametersYamlField), deployModeYamlField)...)

	modulesNode := getPropContent(mtaNode, modulesYamlField)

	for i, module := range mta.Modules {
		if module.BuildParams != nil {
			issues = append(issues, checkStringProperty(module.BuildParams, getPropValueByName(modulesNode[i], buildParametersYamlField), builderYamlField)...)
			if module.BuildParams[commandsYamlField] != nil {
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
					issues = appendIssue(issues, `the "commands" property is defined incorrectly; the property must be a sequence of strings`, commandsParamsNode.Line, commandsParamsNode.Column)
				}
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
