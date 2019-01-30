package validate

import (
	"gopkg.in/yaml.v2"

	"github.com/SAP/cloud-mta/mta"
)

type checkSemantic func(mta *mta.MTA, source string) []YamlValidationIssue

// runSemanticValidations - runs semantic validations
func runSemanticValidations(yamlContent []byte, source string) []YamlValidationIssue {
	var issues []YamlValidationIssue
	mtaStr := mta.MTA{}
	err := yaml.Unmarshal(yamlContent, &mtaStr)
	if err != nil {
		issues = appendIssue(issues, "validation failed when unmarshalling the MTA file because: "+err.Error())
		return issues
	}
	validations := []checkSemantic{validateModulesPaths, validateNamesUniqueness}
	for _, validation := range validations {
		validationIssues := validation(&mtaStr, source)
		issues = append(issues, validationIssues...)

	}
	return issues
}
