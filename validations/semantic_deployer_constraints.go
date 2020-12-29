package validate

import (
	"fmt"
	"gopkg.in/yaml.v3"

	"github.com/SAP/cloud-mta/mta"
)

const (
	missingConfigDocLink = "https://sap.github.io/cloud-mta-build-tool/migration/#features-that-are-handled-differently-in-the-cloud-mta-build-tool"
	missingConfigsMsg    = `the "%s" module does not contain the mandatory "supported-platforms" and "build-result" build configurations; see %q`
	missingConfigMsg     = `the "%s" module does not contain the mandatory "%s" build configuration; see %q`
)

func checkDeployerConstraints(mta *mta.MTA, mtaNode *yaml.Node, source string, strict bool) ([]YamlValidationIssue, []YamlValidationIssue) {
	var issues []YamlValidationIssue

	if mta.Parameters != nil && mta.Parameters[deployModeYamlField] != nil {
		deployMode, ok := mta.Parameters[deployModeYamlField].(string)
		// case !ok is handled by schema validations
		if ok && deployMode == html5RepoDeployMode {
			issues = append(issues, checkHTML5ModulesParams(mta, mtaNode)...)
		}
	}

	return issues, nil
}

func checkHTML5ModulesParams(mta *mta.MTA, mtaNode *yaml.Node) []YamlValidationIssue {
	var issues []YamlValidationIssue

	modulesNode := getPropContent(mtaNode, modulesYamlField)

	for i, module := range mta.Modules {
		if module.Type == html5ModuleType {
			issues = append(issues, checkHTML5ModuleParams(module, modulesNode[i])...)
		}
	}

	return issues
}

func checkHTML5ModuleParams(module *mta.Module, moduleNode *yaml.Node) []YamlValidationIssue {

	supportedPlatformsDefined := false
	buildResultDefined := false
	if module.BuildParams != nil {
		supportedPlatformsDefined = module.BuildParams[supportedPlatformsYamlField] != nil
		buildResultDefined = module.BuildParams[buildResultYamlField] != nil
	}

	if !supportedPlatformsDefined && !buildResultDefined {
		return []YamlValidationIssue{
			{
				Msg:    fmt.Sprintf(missingConfigsMsg, module.Name, missingConfigDocLink),
				Line:   moduleNode.Line,
				Column: moduleNode.Column,
			},
		}
	} else if !supportedPlatformsDefined {
		return []YamlValidationIssue{
			{
				Msg:    fmt.Sprintf(missingConfigMsg, module.Name, supportedPlatformsYamlField, missingConfigDocLink),
				Line:   moduleNode.Line,
				Column: moduleNode.Column,
			},
		}
	} else if !buildResultDefined {
		return []YamlValidationIssue{
			{
				Msg:    fmt.Sprintf(missingConfigMsg, module.Name, buildResultYamlField, missingConfigDocLink),
				Line:   moduleNode.Line,
				Column: moduleNode.Column,
			},
		}
	}
	return nil
}
