package validate

import (
	"gopkg.in/yaml.v3"
	"strings"

	"github.com/SAP/cloud-mta/mta"
)

type checkSemantic func(mta *mta.MTA, root *yaml.Node, source string, strict bool) (errors []YamlValidationIssue, warnings []YamlValidationIssue)

const (
	configuration               = "configuration"
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
	noSourceYamlField           = "no-source"
	buildResultYamlField        = "build-result"
	supportedPlatformsYamlField = "supported-platforms"
	commandsYamlField           = "commands"
	beforeAllYamlField          = "before-all"
	afterAllYamlField           = "after-all"
	deployModeYamlField         = "deploy_mode"
	parametersMetadataField     = "parameters-metadata"
	propertiesMetadataField     = "properties-metadata"
	hooksYamlField              = "hooks"
	datatypeYamlField           = "datatype"
	publicYamlField             = "public"
	listYamlField               = "list"
	groupYamlField              = "group"

	npmOptsYamlField   = "npm-opts"
	gruntOptsYamlField = "grunt-opts"
	mavenOptsYamlField = "maven-opts"

	pathsValidation               = "paths"
	emptyPathValidation           = "emptyPath"
	namesValidation               = "names"
	requiredValidation            = "required"
	buildersValidation            = "builders"
	deprecatedOptsValidation      = "deprecatedOpts"
	deployerConstrValidation      = "deployerConstraints"
	metadataValidation            = "metadata"
	ifNoSourceParamBoolValidation = "checkNoSourceParam"

	propertiesMtaField      = "Properties"
	parametersMtaField      = "Parameters"
	buildParametersMtaField = "BuildParams"
	nameMtaField            = "Name"

	moduleEntityKind       = "module"
	resourceEntityKind     = "resource"
	propertyEntityKind     = "property"
	parameterEntityKind    = "parameter"
	buildParamEntityKind   = "build parameter"
	providedPropEntityKind = "provided property set"
	requiresPropEntityKind = "requires property set"
	hookPropEntityKind     = "hook"

	html5RepoDeployMode = "html5-repo"

	html5ModuleType = "html5"
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
	if !strings.Contains(exclude, emptyPathValidation) {
		validations = append(validations, ifModulePathEmpty)
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
	if !strings.Contains(exclude, metadataValidation) {
		validations = append(validations, checkParamsAndPropertiesMetadata)
	}
	if !strings.Contains(exclude, ifNoSourceParamBoolValidation) {
		validations = append(validations, ifNoSourceParamBool)
	}

	return validations
}

func getIndexedNodePropPosition(node *yaml.Node, index int, propName string) (line int, column int, propFound bool) {
	indexedNode := node.Content[index]
	nameNode := getPropValueByName(indexedNode, propName)
	if nameNode == nil {
		return indexedNode.Line, indexedNode.Column, false // First line of the indexed node (in case we can't find the property inside the node)
	}
	return nameNode.Line, nameNode.Column, true
}

func getNamedObjectNodeByIndex(parentNode *yaml.Node, fieldName string, index int) *yaml.Node {
	objectsNode := getPropValueByName(parentNode, fieldName)
	return objectsNode.Content[index]
}

func getNamedObjectPositionByIndex(parentNode *yaml.Node, fieldName string, index int) (line int, column int) {
	objectsNode := getPropValueByName(parentNode, fieldName)
	line, column, _ = getIndexedNodePropPosition(objectsNode, index, nameYamlField)
	return line, column
}
