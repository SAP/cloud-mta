package validate

import (
	"gopkg.in/yaml.v3"
	"strings"

	"github.com/SAP/cloud-mta/mta"
)

type checkExtSemantic func(mta *mta.EXT, root *yaml.Node, source string, strict bool) (errors []YamlValidationIssue, warnings []YamlValidationIssue)

// runSemanticValidations - runs semantic validations
func runExtSemanticValidations(mtaExt *mta.EXT, root *yaml.Node, source string, exclude string, strict bool) ([]YamlValidationIssue, []YamlValidationIssue) {
	var errors []YamlValidationIssue
	var warnings []YamlValidationIssue

	validations := getExtSemanticValidations(exclude)
	for _, validation := range validations {
		validationErrors, validationWarnings := validation(mtaExt, root, source, strict)
		errors = append(errors, validationErrors...)
		warnings = append(warnings, validationWarnings...)
	}
	return errors, warnings
}

// getSemanticValidations - gets list of all semantic validations minus excludes validations
func getExtSemanticValidations(exclude string) []checkExtSemantic {
	var validations []checkExtSemantic
	if !strings.Contains(exclude, namesValidation) {
		validations = append(validations, checkSingleExtendNames)
	}
	if !strings.Contains(exclude, deprecatedOptsValidation) {
		validations = append(validations, checkExtDeprecatedOpts)
	}
	return validations
}
