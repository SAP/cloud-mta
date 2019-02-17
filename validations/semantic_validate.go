package validate

import (
	"github.com/SAP/cloud-mta/mta"
)

type checkSemantic func(mta *mta.MTA, source string) []YamlValidationIssue

// runSemanticValidations - runs semantic validations
func runSemanticValidations(mtaStr *mta.MTA, source string) []YamlValidationIssue {
	var issues []YamlValidationIssue

	validations := []checkSemantic{ifModulePathExists, isNameUnique, ifRequiredDefined}
	for _, validation := range validations {
		validationIssues := validation(mtaStr, source)
		issues = append(issues, validationIssues...)

	}
	return issues
}
