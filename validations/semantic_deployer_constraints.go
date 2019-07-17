package validate

import (
	"fmt"
	"gopkg.in/yaml.v3"

	"github.com/SAP/cloud-mta/mta"
)

const (
	moreInfo          = `for more information, see %q`
	moreInfoAddr      = "https://sap.github.io/cloud-mta-build-tool/migration/#features-that-are-handled-differently-in-the-cloud-mta-build-tool"
	missingConfigsMsg = `the mandatory build configurations "supported-platforms" and "build-result" are missing for the %s module\n` + moreInfo
	missingConfigMsg  = `the mandatory build configuration "%s" is missing for the %s module\n` + moreInfo
)

func checkDeployerConstraints(mta *mta.MTA, mtaNode *yaml.Node, source string, strict bool) ([]YamlValidationIssue, []YamlValidationIssue) {
	var issues []YamlValidationIssue

	if mta.Parameters != nil && mta.Parameters[deployModeYamlField] != nil {
		deployMode, ok := mta.Parameters[deployModeYamlField].(string)
		// case !ok is handled by schema validations
		if ok && deployMode == html5RepoDeployMode {
			issues = append(issues, checkHtmlModulesParams(mta, mtaNode)...)
		}
	}

	return issues, nil
}

func checkHtmlModulesParams(mta *mta.MTA, mtaNode *yaml.Node) []YamlValidationIssue {
	var issues []YamlValidationIssue

	modulesNode := getPropContent(mtaNode, modulesYamlField)

	for i, module := range mta.Modules {
		if module.Type == "html5" {
			issues = append(issues, checkHtmlModuleParams(module, modulesNode[i])...)
		}
	}

	return issues
}

func checkHtmlModuleParams(module *mta.Module, moduleNode *yaml.Node) []YamlValidationIssue {

	supportedPlatformsDefined := false
	buildResultDefined := false
	if module.BuildParams != nil {
		supportedPlatformsDefined = module.BuildParams[supportedPlatformsYamlField] != nil
		buildResultDefined = module.BuildParams[buildResultYamlField] != nil
	}

	if !supportedPlatformsDefined && !buildResultDefined {
		return []YamlValidationIssue{
			{
				Msg:  fmt.Sprintf(missingConfigsMsg, module.Name, moreInfoAddr),
				Line: moduleNode.Line,
			},
		}
	} else if !supportedPlatformsDefined {
		return []YamlValidationIssue{
			{
				Msg:  fmt.Sprintf(missingConfigMsg, supportedPlatformsYamlField, module.Name, moreInfoAddr),
				Line: moduleNode.Line,
			},
		}
	} else if !buildResultDefined {
		return []YamlValidationIssue{
			{
				Msg:  fmt.Sprintf(missingConfigMsg, buildResultYamlField, module.Name, moreInfoAddr),
				Line: moduleNode.Line,
			},
		}
	}
	return nil
}
