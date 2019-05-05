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
				buildParamsNode := getPropValueByName(modulesNode[i], buildParametersYamlField)
				builderNode := getPropValueByName(buildParamsNode, builderYamlField)
				issues = appendIssue(issues, `the "builder" property is defined incorrectly; the property must be a string`, builderNode.Line)
			} else {
				if module.BuildParams[commandsYamlField] != nil {
					_, ok := module.BuildParams[commandsYamlField].([]string)
					commands, okI := module.BuildParams[commandsYamlField].([]interface{})
					if okI {
						for _, commandI := range commands {
							_, okI = commandI.(string)
							if !okI {
								break
							}
						}
					}
					if !ok && !okI {
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
