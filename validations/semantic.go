package validate

import (
	"github.com/SAP/cloud-mta/mta"
)

type checkSemantic func(mta *mta.MTA, source string) []YamlValidationIssue

// runSemanticValidations - runs semantic validations
func runSemanticValidations(mta *mta.MTA, source string) []YamlValidationIssue {
	validations := []checkSemantic{validateModulesPaths, validateNamesUniqueness}
	var issues []YamlValidationIssue
	for _, validation := range validations {
		validationIssues := validation(mta, source)
		issues = append(issues, validationIssues...)

	}
	return issues
}
