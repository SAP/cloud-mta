package validate

import (
	"gopkg.in/yaml.v3"
	"strings"

	"github.com/SAP/cloud-mta/mta"
)

type checkSemantic func(mta *mta.MTA, root *yaml.Node, source string, strict bool) (errors []YamlValidationIssue, warnings []YamlValidationIssue)

const (
	pathYamlField               = "path"
	nameYamlField               = "name"
	modulesYamlField            = "modules"
	providesYamlField           = "provides"
	resourcesYamlField          = "resources"
	requiresYamlField           = "requires"
	propertiesYamlField         = "properties"
	parametersYamlField         = "parameters"
	buildParametersYamlField    = "build-parameters"
	builderYamlField            = "builder"
	buildResultYamlField        = "build-result"
	supportedPlatformsYamlField = "supported-platforms"
	commandsYamlField           = "commands"
	beforeAllYamlField          = "before-all"
	afterAllYamlField           = "after-all"
	deployModeYamlField         = "deploy-mode"

	npmOptsYamlField   = "npm-opts"
	gruntOptsYamlField = "grunt-opts"
	mavenOptsYamlField = "maven-opts"

	pathsValidation          = "paths"
	namesValidation          = "names"
	requiredValidation       = "required"
	buildersValidation       = "builders"
	deprecatedOptsValidation = "deprecatedOpts"
	deployerConstrValidation = "deployerConstraints"

	propertiesMtaField      = "Properties"
	parametersMtaField      = "Parameters"
	buildParametersMtaField = "BuildParams"
	nameMtaField            = "Name"

	moduleEntityKind       = "module"
	propertyEntityKind     = "property"
	parameterEntityKind    = "parameter"
	buildParamEntityKind   = "build parameter"
	providedPropEntityKind = "provided property set"

	html5RepoDeployMode = "html5-repo"
)

// runSemanticValidations - runs semantic validations
func runSemanticValidations(mtaStr *mta.MTA, root *yaml.Node, source string, exclude string, strict bool) ([]YamlValidationIssue, []YamlValidationIssue) {
	var errors []YamlValidationIssue
	var warnings []YamlValidationIssue

	validations := getSemanticValidations(exclude)
	for _, validation := range validations {
		validationErrors, validationWarnings := validation(mtaStr, root, source, strict)
		errors = append(errors, validationErrors...)
		warnings = append(warnings, validationWarnings...)

	}
	return errors, warnings
}

// getSemanticValidations - gets list of all semantic validations minus excludes validations
func getSemanticValidations(exclude string) []checkSemantic {
	var validations []checkSemantic
	if !strings.Contains(exclude, pathsValidation) {
		validations = append(validations, ifModulePathExists)
	}
	if !strings.Contains(exclude, namesValidation) {
		validations = append(validations, isNameUnique)
	}
	if !strings.Contains(exclude, requiredValidation) {
		validations = append(validations, ifRequiredDefined)
	}
	if !strings.Contains(exclude, buildersValidation) {
		validations = append(validations, checkBuildersSemantic)
	}
	if !strings.Contains(exclude, deprecatedOptsValidation) {
		validations = append(validations, checkDeprecatedOpts)
	}
	if !strings.Contains(exclude, deployerConstrValidation) {
		validations = append(validations, checkDeployerConstraints)
	}

	return validations
}
