package validate

import (
	"fmt"
	"gopkg.in/yaml.v3"

	"github.com/SAP/cloud-mta/mta"
)

const (
	deprecatedOptMsg = `the "%s" build configuration parameter is not supported by Cloud MTA Build tool; use "custom" builder instead\n` +
		`for more details on the "custom" builder, see %q`

	customBuilderDocLink = "https://sap.github.io/cloud-mta-build-tool/configuration/#configuring-the-custom-builder"
)

func checkDeprecatedOpt(module *mta.Module, buildParamsNode *yaml.Node, optFieldName string) []YamlValidationIssue {

	if module.BuildParams[optFieldName] != nil {
		optsNode := getPropByName(buildParamsNode, optFieldName)
		return []YamlValidationIssue{{Msg: fmt.Sprintf(deprecatedOptMsg, optFieldName, customBuilderDocLink), Line: optsNode.Line}}
	}
	return nil
}

func checkDeprecatedOpts(mta *mta.MTA, mtaNode *yaml.Node, source string, strict bool) ([]YamlValidationIssue, []YamlValidationIssue) {
	var issues []YamlValidationIssue

	modulesNode := getPropContent(mtaNode, modulesYamlField)

	opts := []string{npmOptsYamlField, gruntOptsYamlField, mavenOptsYamlField}

	for i, module := range mta.Modules {
		if module.BuildParams != nil {
			buildParamsNode := getPropValueByName(modulesNode[i], buildParametersYamlField)
			for _, opt := range opts {
				issues = append(issues, checkDeprecatedOpt(module, buildParamsNode, opt)...)
			}
		}
	}

	return issues, nil
}
