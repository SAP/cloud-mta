package validate

import (
	"fmt"
	"gopkg.in/yaml.v3"

	"github.com/SAP/cloud-mta/mta"
)

const (
	deprecatedOptMsg     = `the "%s" build configuration parameter is not supported by Cloud MTA Build tool; use the "custom" builder instead; see %q`
	customBuilderDocLink = "https://sap.github.io/cloud-mta-build-tool/configuration/#configuring-the-custom-builder"
)

func checkDeprecatedOpt(buildParams map[string]interface{}, buildParamsNode *yaml.Node, optFieldName string) []YamlValidationIssue {

	if buildParams[optFieldName] != nil {
		optsNode := getPropByName(buildParamsNode, optFieldName)
		return []YamlValidationIssue{{Msg: fmt.Sprintf(deprecatedOptMsg, optFieldName, customBuilderDocLink), Line: optsNode.Line, Column: optsNode.Column}}
	}
	return nil
}

func checkDeprecatedOpts(mta *mta.MTA, mtaNode *yaml.Node, source string, strict bool) ([]YamlValidationIssue, []YamlValidationIssue) {
	var issues []YamlValidationIssue

	modulesNode := getPropContent(mtaNode, modulesYamlField)

	opts := []string{npmOptsYamlField, gruntOptsYamlField, mavenOptsYamlField}

	for i, module := range mta.Modules {
		buildParams := module.BuildParams
		if buildParams != nil {
			buildParamsNode := getPropValueByName(modulesNode[i], buildParametersYamlField)
			for _, opt := range opts {
				issues = append(issues, checkDeprecatedOpt(buildParams, buildParamsNode, opt)...)
			}
		}
	}

	return issues, nil
}

func checkExtDeprecatedOpts(mta *mta.EXT, mtaNode *yaml.Node, source string, strict bool) ([]YamlValidationIssue, []YamlValidationIssue) {
	var issues []YamlValidationIssue

	modulesNode := getPropContent(mtaNode, modulesYamlField)

	opts := []string{npmOptsYamlField, gruntOptsYamlField, mavenOptsYamlField}

	for i, module := range mta.Modules {
		buildParams := module.BuildParams
		if buildParams != nil {
			buildParamsNode := getPropValueByName(modulesNode[i], buildParametersYamlField)
			for _, opt := range opts {
				issues = append(issues, checkDeprecatedOpt(buildParams, buildParamsNode, opt)...)
			}
		}
	}

	return issues, nil
}
