// 1. Position information for schema issues (blocked).
// 2. Path information for schema issues.	// 2. Path information for schema issues.
// 4. regExp patterns also implicitly require a NotMapSequence validation	// 4. regExp patterns also implicitly require a NotMapSequence validation.

package validate

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/smallfish/simpleyaml"
)

// buildValidationsFromSchemaText is an entry point that accepts a .yaml file schema as plain text and produces the YAML validation functions
// and schema issues that are detected.
func buildValidationsFromSchemaText(yaml []byte) ([]YamlCheck, []YamlValidationIssue) {
	var validations []YamlCheck
	var schemaIssues []YamlValidationIssue

	y, parseError := simpleyaml.NewYaml(yaml)
	if parseError != nil {
		schemaIssues = appendIssue(schemaIssues, "validation failed when parsing the MTA schema file: "+parseError.Error(), 0, 0)
		return validations, schemaIssues
	}

	return buildValidationsFromSchema(y)
}

// Internal YAML validation builder
// Will be called recursively and traverse the schema structure.
func buildValidationsFromSchema(schema *simpleyaml.Yaml) ([]YamlCheck, []YamlValidationIssue) {
	var validations []YamlCheck
	var schemaIssues []YamlValidationIssue

	typeNode := schema.Get("type")
	typeNodeValue, _ := typeNode.String()
	if typeNode.IsFound() && (typeNodeValue == "map" || typeNodeValue == "seq") {
		switch typeNodeValue {
		// type: map
		// mapping:
		//   ...
		case "map":
			mappingNode := schema.Get("mapping")
			if !mappingNode.IsMap() {
				schemaIssues = appendIssue(schemaIssues, "invalid .yaml file schema: the mapping node must be a map", 0, 0)
				return validations, schemaIssues
			}
			mappingValidations, mappingSchemaIssues := buildValidationsForMapping(mappingNode)
			schemaIssues = append(schemaIssues, mappingSchemaIssues...)
			// Only perform validations on the mapping value when it isn't null, unless specified as required
			mappingValidations, schemaIssues = buildOptionalOrRequiredValidation(schema, mappingValidations, schemaIssues)
			validations = append(validations, mappingValidations...)

			// type: seq
			// sequence:
			//  - type: map
			//    mapping:
			//      name: {required: true}
			//    ...
		case "seq":
			sequenceNode := schema.Get("sequence")
			if !sequenceNode.IsArray() {
				schemaIssues = appendIssue(schemaIssues, "invalid .yaml file schema: the sequence node must be an array", 0, 0)
				return validations, schemaIssues
			}

			seqSize, _ := sequenceNode.GetArraySize()
			if seqSize != 1 {
				schemaIssues = appendIssue(schemaIssues, "invalid .yaml file schema: the sequence node must have exactly one item", 0, 0)
				return validations, schemaIssues
			}

			// A sequence schema node has exactly 1 element
			sequenceItemNode := sequenceNode.GetIndex(0)
			sequenceValidations, newSchemaIssues := buildValidationsFromSequence(sequenceItemNode)
			schemaIssues = append(schemaIssues, newSchemaIssues...)
			// Only perform validations on the sequence values when they aren't null, unless specified as required
			sequenceValidations, schemaIssues = buildOptionalOrRequiredValidation(schema, sequenceValidations, schemaIssues)
			validations = append(validations, sequenceValidations...)
		}
		// {required: true, pattern: /^[a-z]+$/}
	} else {
		return buildLeafValidations(schema)
	}

	return validations, schemaIssues
}

func buildValidationsForMapping(mappingNode *simpleyaml.Yaml) ([]YamlCheck, []YamlValidationIssue) {
	var validations []YamlCheck
	var schemaIssues []YamlValidationIssue

	// Default rule for unknown keys - normally unmarshalled to Go map
	if keys, _ := mappingNode.GetMapKeys(); len(keys) == 1 && keys[0] == "=" {
		// type: map
		// mapping:
		//   =:
		//     type: ...
		value := mappingNode.Get(keys[0])
		propInnerValidations, newSchemaIssues := buildValidationsFromSchema(value)
		propWrapperValidation := forEachProperty(propInnerValidations...)
		newValidations := append(validations, propWrapperValidation)

		schemaIssues = append(schemaIssues, newSchemaIssues...)
		validations = append(validations, newValidations...)
	} else {
		// type: map
		// mapping:
		//   firstName:  {required: true}
		//   ...
		newValidations, newSchemaIssues := buildValidationsFromMap(mappingNode)
		schemaIssues = append(schemaIssues, newSchemaIssues...)
		validations = append(validations, newValidations...)
	}
	return validations, schemaIssues
}

// Create Validations for a mapping
// each key's inner validations will be wrapping in a "property" validation
func buildValidationsFromMap(y *simpleyaml.Yaml) ([]YamlCheck, []YamlValidationIssue) {
	var validations []YamlCheck
	var schemaIssues []YamlValidationIssue

	validations = append(validations, typeIsMap())

	keys, _ := y.GetMapKeys()
	for _, key := range keys {
		value := y.Get(key)
		propInnerValidations, newSchemaIssues := buildValidationsFromSchema(value)
		schemaIssues = append(schemaIssues, newSchemaIssues...)
		propWrapperValidation := property(key, propInnerValidations...)
		validations = append(validations, propWrapperValidation)
	}
	return validations, schemaIssues
}

// Creates validations for a sequence
// Will wrap the nested checks with a "typeIsArray" check and iterate over the elements
// using "forEach"
func buildValidationsFromSequence(y *simpleyaml.Yaml) ([]YamlCheck, []YamlValidationIssue) {
	var validations []YamlCheck

	sequenceInnerValidations, newIssues := buildValidationsFromSchema(y)
	seqChecksWrapper := sequenceFailFast(typeIsArray(), forEach(sequenceInnerValidations...))
	validations = append(validations, seqChecksWrapper)

	return validations, newIssues
}

// Will create the "edge" nodes validations, these are specific checks
// for a specific path at the end of the YAML Schema
// e.g: {required: true, pattern: /^[a-zA-Z]$/}
func buildLeafValidations(y *simpleyaml.Yaml) ([]YamlCheck, []YamlValidationIssue) {
	var validations []YamlCheck
	var schemaIssues []YamlValidationIssue

	validations, schemaIssues = invokeLeafValidation(y, validations, schemaIssues, buildTypeValidation)

	validations, schemaIssues = invokeLeafValidation(y, validations, schemaIssues, buildPatternValidation)

	// Special handling is needed for "optional" and "required" Validations, must be invoked last
	// and receive all previous built validations as extra wrapping will be done here.
	return buildOptionalOrRequiredValidation(y, validations, schemaIssues)

}

func buildOptionalOrRequiredValidation(y *simpleyaml.Yaml, validations []YamlCheck, schemaIssues []YamlValidationIssue) ([]YamlCheck, []YamlValidationIssue) {
	requiredNode := y.Get("required")
	if requiredNode.IsFound() {
		requiredValue := getLiteralStringValue(requiredNode)
		if requiredValue != "true" && requiredValue != "false" {
			schemaIssues = appendIssue(schemaIssues,
				fmt.Sprint("invalid .yaml file schema: the required node must be a boolean "), 0, 0)
			return validations, schemaIssues
		}

		// When a "required" check is needed there is no need to perform additional validations
		// if the property is missing, thus we use "sequenceFailFast"
		if requiredValue == "true" {
			// The required check must be performed first in the sequence.
			validationsWithRequiredFirst := append([]YamlCheck{required()}, validations...)
			validations = []YamlCheck{sequenceFailFast(validationsWithRequiredFirst...)}
			return validations, schemaIssues
		}
	}
	// When "required" is false or doesn't exist, we must only perform additional validations
	// if the property actually exists, thus we use "optional"
	// An optional wraps all our other validations.
	validations = []YamlCheck{optional(validations...)}
	return validations, schemaIssues
}

func buildTypeValidation(y *simpleyaml.Yaml) ([]YamlCheck, []YamlValidationIssue) {
	var validations []YamlCheck
	var schemaIssues []YamlValidationIssue

	typeNode := y.Get("type")
	if typeNode.IsFound() {
		typeValue, stringErr := typeNode.String()
		if stringErr != nil {
			schemaIssues = appendIssue(schemaIssues, "invalid .yaml file schema: the type node must be a string", 0, 0)
			return validations, schemaIssues
		}
		if typeValue == "bool" {
			validations = append(validations, typeIsBoolean())
		}
	}

	enumNode := y.Get("enum")
	if enumNode.IsFound() {
		enumValidations, enumSchemaIssues := buildEnumValidation(enumNode)
		validations = append(validations, enumValidations...)
		schemaIssues = append(schemaIssues, enumSchemaIssues...)
	}
	return validations, schemaIssues
}

func buildEnumValidation(enumNode *simpleyaml.Yaml) ([]YamlCheck, []YamlValidationIssue) {
	if !enumNode.IsArray() {
		return []YamlCheck{}, []YamlValidationIssue{{"invalid .yaml file schema: enums values must be listed as an array", 0, 0}}
	}

	enumsNumber, _ := enumNode.GetArraySize()

	var enumValues []string
	for i := 0; i < enumsNumber; i++ {
		enumValueNode := enumNode.GetIndex(i)
		if enumValueNode.IsArray() || enumValueNode.IsMap() {
			return []YamlCheck{}, []YamlValidationIssue{{"invalid .yaml file schema: enum values must be simple", 0, 0}}
		}
		enumValue := getLiteralStringValue(enumValueNode)
		enumValues = append(enumValues, enumValue)
	}

	return []YamlCheck{matchesEnumValues(enumValues)}, []YamlValidationIssue{}
}

func buildPatternValidation(y *simpleyaml.Yaml) ([]YamlCheck, []YamlValidationIssue) {
	var validations []YamlCheck
	var schemaIssues []YamlValidationIssue

	patternNode := y.Get("pattern")
	if patternNode.IsFound() {
		patternValue, err := patternNode.String()
		if err != nil {
			schemaIssues = appendIssue(schemaIssues, "invalid .yaml file schema: the pattern node must be a string", 0, 0)
			return validations, schemaIssues
		}
		// TODO: we must validate: NOT MAP/SEQ
		patternWithoutSlashes := strings.TrimSuffix(strings.TrimPrefix(patternValue, "/"), "/")
		_, err = regexp.Compile(patternWithoutSlashes)
		if err != nil {
			schemaIssues = append(schemaIssues,
				YamlValidationIssue{
					Msg:    "invalid .yaml file schema: the pattern node is invalid because: " + err.Error(),
					Line:   0,
					Column: 0,
				})
		} else {
			validations = append(validations, matchesRegExp(patternWithoutSlashes))
		}
	}
	return validations, schemaIssues
}

// Utility to reduce verbosity
func invokeLeafValidation(y *simpleyaml.Yaml, validations []YamlCheck, schemaIsssues []YamlValidationIssue,
	leafBuilder func(y *simpleyaml.Yaml) ([]YamlCheck, []YamlValidationIssue)) ([]YamlCheck, []YamlValidationIssue) {
	newValidations, newSchemaIssues := leafBuilder(y)
	validations = append(validations, newValidations...)
	schemaIsssues = append(schemaIsssues, newSchemaIssues...)

	return validations, schemaIsssues
}

func appendIssue(issues []YamlValidationIssue, issue string, line int, column int) []YamlValidationIssue {
	if issue == "" {
		return issues
	}
	return append(issues, []YamlValidationIssue{{Msg: issue, Line: line, Column: column}}...)
}
