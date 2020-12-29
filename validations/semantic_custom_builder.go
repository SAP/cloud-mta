package validate

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/SAP/cloud-mta/mta"
)

const (
	customBuilder = "custom"
)

func checkProjectBuilders(builders []mta.ProjectBuilder, mtaNode *yaml.Node, fieldName string) []YamlValidationIssue {
	var issues []YamlValidationIssue

	if builders == nil {
		return nil
	}

	buildParamsNode := getPropValueByName(mtaNode, buildParametersYamlField)
	buildersNodes := getPropContent(buildParamsNode, fieldName)
	for i, builderStr := range builders {
		builder := builderStr.Builder
		commandsDefined := builderStr.Commands != nil
		commandsNode := getPropValueByName(buildersNodes[i], commandsYamlField)
		issues = append(issues, checkCustomBuilder(builder, commandsDefined, buildersNodes[i], commandsNode)...)
	}
	return issues
}

func checkCustomBuilder(builder string, commandsDefined bool, builderNode *yaml.Node, commandsNode *yaml.Node) []YamlValidationIssue {
	if builder == customBuilder && !commandsDefined {
		return []YamlValidationIssue{{Msg: `the "commands" property is missing in the "custom" builder`, Line: builderNode.Line, Column: builderNode.Column}}
	} else if builder != customBuilder && commandsDefined {
		return []YamlValidationIssue{{Msg: fmt.Sprintf(`the "commands" property is not supported by the "%s" builder`, builder), Line: commandsNode.Line, Column: commandsNode.Column}}
	}
	return nil
}

func checkBuildersSemantic(mta *mta.MTA, mtaNode *yaml.Node, source string, strict bool) ([]YamlValidationIssue, []YamlValidationIssue) {
	var issues []YamlValidationIssue

	if mta.BuildParams != nil {
		issues = append(issues, checkProjectBuilders(mta.BuildParams.BeforeAll, mtaNode, beforeAllYamlField)...)
		issues = append(issues, checkProjectBuilders(mta.BuildParams.AfterAll, mtaNode, afterAllYamlField)...)
	}

	modulesNode := getPropContent(mtaNode, modulesYamlField)

	for i, module := range mta.Modules {
		if module.BuildParams != nil && module.BuildParams[builderYamlField] != nil {
			builder, ok := module.BuildParams[builderYamlField].(string)
			if !ok {
				// not an issue of semantics, handled by schema validation
				continue
			}
			buildParamsNode := getPropValueByName(modulesNode[i], buildParametersYamlField)
			builderNode := getPropValueByName(buildParamsNode, builderYamlField)
			commandsDefined := module.BuildParams[commandsYamlField] != nil
			commandsNode := getPropValueByName(buildParamsNode, commandsYamlField)
			issues = append(issues, checkCustomBuilder(builder, commandsDefined, builderNode, commandsNode)...)
		}
	}

	if strict {
		return issues, nil
	}
	return nil, issues
}
