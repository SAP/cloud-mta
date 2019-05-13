package validate

import (
	"gopkg.in/yaml.v3"

	"github.com/SAP/cloud-mta/mta"
)

func checkBuilderSchema(mta *mta.MTA, mtaNode *yaml.Node, source string) []YamlValidationIssue {
	var issues []YamlValidationIssue

	modulesNode := getPropContent(mtaNode, modulesYamlField)

	for i, module := range mta.Modules {
		if module.BuildParams != nil && module.BuildParams[builderYamlField] != nil {
			_, ok := module.BuildParams[builderYamlField].(string)
			if !ok {
				// the "builder" field is not a string
				buildParamsNode := getPropValueByName(modulesNode[i], buildParametersYamlField)
				builderNode := getPropValueByName(buildParamsNode, builderYamlField)
				issues = appendIssue(issues, `the "builder" property is defined incorrectly; the property must be a string`, builderNode.Line)
			} else {
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
						issues = appendIssue(issues, `the "commands" property is defined incorrectly; the property must be a sequence of strings`, commandsParamsNode.Line)
					}
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
