package validate

import (
	"fmt"
	"gopkg.in/yaml.v3"

	"github.com/SAP/cloud-mta/mta"
)

const (
	unknownNameInMetadataMsg             = `metadata cannot be defined for the "%s" undefined %s`
	emptyRequiredFieldMsg                = `the value for the required and non-overwritable "%s" %s cannot be empty`
	propertiesMetadataWithListOrGroupMsg = `the "properties-metadata" cannot be used in the context of list and group`

	// Key for mapTypes map
	mapTypeParameters = iota
	mapTypeProperties = iota
)

// Maps cannot be constants but this map is effectively used as a const
var mapTypes = map[int]mapTypeProps{
	mapTypeParameters: {
		entityKind:       parameterEntityKind,
		mapNodeName:      parametersYamlField,
		metadataNodeName: parametersMetadataField,
	},
	mapTypeProperties: {
		entityKind:       propertyEntityKind,
		mapNodeName:      propertiesYamlField,
		metadataNodeName: propertiesMetadataField,
	},
}

type mapTypeProps struct {
	entityKind       string
	mapNodeName      string
	metadataNodeName string
}

func checkParamsAndPropertiesMetadata(mta *mta.MTA, mtaNode *yaml.Node, source string, strict bool) (errors []YamlValidationIssue, warnings []YamlValidationIssue) {
	issues := validateMetadata(mta, mtaNode, source, checkMetadata)

	if strict {
		return issues, nil
	}
	return nil, issues
}

// Check:
// 1. Each property/parameter in the metadata is defined in the map
// 2. Non-optional and non-overwritable property/parameter is not nil
// 3. properties-metadata is not defined when list or group is defined
func checkMetadata(m map[string]interface{}, metadata map[string]mta.MetaData, parentNode *yaml.Node, mapType int) []YamlValidationIssue {
	var issues []YamlValidationIssue

	metadataNodeValue := getPropValueByName(parentNode, mapTypes[mapType].metadataNodeName)
	issues = checkMetadataKeyIsDefinedInMap(metadata, m, metadataNodeValue, issues, mapType)

	mapNode := getPropValueByName(parentNode, mapTypes[mapType].mapNodeName)
	issues = checkNoEmptyRequiredFields(metadata, m, mapNode, issues, mapType)

	metadataNodeName := getPropByName(parentNode, mapTypes[mapType].metadataNodeName)
	issues = checkPropertiesMetadataWithListOrGroup(mapType, metadataNodeName, parentNode, issues)

	return issues
}

func checkMetadataKeyIsDefinedInMap(metadata map[string]mta.MetaData, m map[string]interface{}, metadataNodeValue *yaml.Node, issues []YamlValidationIssue, mapType int) []YamlValidationIssue {
	for key := range metadata {
		_, ok := m[key]
		if !ok {
			keyNode := getPropByName(metadataNodeValue, key)
			issues = append(issues, YamlValidationIssue{Msg: fmt.Sprintf(unknownNameInMetadataMsg, key, mapTypes[mapType].entityKind), Line: keyNode.Line, Column: keyNode.Column})
		}
	}
	return issues
}

func checkNoEmptyRequiredFields(metadata map[string]mta.MetaData, m map[string]interface{}, mapNode *yaml.Node, issues []YamlValidationIssue, mapType int) []YamlValidationIssue {
	if metadata != nil {
		for key, value := range m {
			// If there's no metadata for the key we don't perform this check (since it's overwritable by default)
			if meta, ok := metadata[key]; ok {
				if !isPropertyOptional(meta.Optional) && !isPropertyOverWritable(meta.OverWritable) && value == nil {
					keyNode := getPropByName(mapNode, key)
					issues = append(issues, YamlValidationIssue{Msg: fmt.Sprintf(emptyRequiredFieldMsg, key, mapTypes[mapType].entityKind), Line: keyNode.Line, Column: keyNode.Column})
				}
			}
		}
	}
	return issues
}

func isPropertyOverWritable(value *bool) bool {
	if value == nil {
		return true
	}
	return *value
}

func isPropertyOptional(value *bool) bool {
	if value == nil {
		return false
	}
	return *value
}

func checkPropertiesMetadataWithListOrGroup(mapType int, metadataNodeName *yaml.Node, parentNode *yaml.Node, issues []YamlValidationIssue) []YamlValidationIssue {
	if mapType == mapTypeProperties && metadataNodeName != nil {
		if getPropByName(parentNode, listYamlField) != nil || getPropByName(parentNode, groupYamlField) != nil {
			issues = append(issues, YamlValidationIssue{Msg: propertiesMetadataWithListOrGroupMsg, Line: metadataNodeName.Line, Column: metadataNodeName.Column})
		}
	}
	return issues
}

// Helper definitions and functions for iterating over all parameters-metadata and properties-metadata fields in the MTA

type metadataValidator func(m map[string]interface{}, metadata map[string]mta.MetaData, parentNode *yaml.Node, mapType int) []YamlValidationIssue

func validateMetadata(mta *mta.MTA, mtaNode *yaml.Node, source string, checkMetadata metadataValidator) []YamlValidationIssue {
	var issues []YamlValidationIssue

	issues = append(issues, checkMetadata(mta.Parameters, mta.ParametersMetaData, mtaNode, mapTypeParameters)...)

	modulesNode := getPropContent(mtaNode, modulesYamlField)
	for i, module := range mta.Modules {
		issues = append(issues, checkMetadata(module.Parameters, module.ParametersMetaData, modulesNode[i], mapTypeParameters)...)
		issues = append(issues, checkMetadata(module.Properties, module.PropertiesMetaData, modulesNode[i], mapTypeProperties)...)

		providesNode := getPropContent(modulesNode[i], providesYamlField)
		for i, provides := range module.Provides {
			issues = append(issues, checkMetadata(provides.Properties, provides.PropertiesMetaData, providesNode[i], mapTypeProperties)...)
		}

		requiresNode := getPropContent(modulesNode[i], requiresYamlField)
		issues = append(issues, checkRequiresParamsAndPropertiesMetadata(requiresNode, module.Requires, checkMetadata)...)

		hooksNode := getPropContent(modulesNode[i], hooksYamlField)
		for i, hook := range module.Hooks {
			issues = append(issues, checkMetadata(hook.Parameters, hook.ParametersMetaData, hooksNode[i], mapTypeParameters)...)

			requiresNode = getPropContent(hooksNode[i], requiresYamlField)
			issues = append(issues, checkRequiresParamsAndPropertiesMetadata(requiresNode, hook.Requires, checkMetadata)...)
		}
	}

	resourcesNode := getPropContent(mtaNode, resourcesYamlField)
	for i, resource := range mta.Resources {
		issues = append(issues, checkMetadata(resource.Parameters, resource.ParametersMetaData, resourcesNode[i], mapTypeParameters)...)
		issues = append(issues, checkMetadata(resource.Properties, resource.PropertiesMetaData, resourcesNode[i], mapTypeProperties)...)

		requiresNode := getPropContent(resourcesNode[i], requiresYamlField)
		issues = append(issues, checkRequiresParamsAndPropertiesMetadata(requiresNode, resource.Requires, checkMetadata)...)
	}

	return issues
}

func checkRequiresParamsAndPropertiesMetadata(requiresNodes []*yaml.Node, requiresList []mta.Requires, validateMetadata metadataValidator) []YamlValidationIssue {
	var issues []YamlValidationIssue

	for i, requires := range requiresList {
		issues = append(issues, validateMetadata(requires.Parameters, requires.ParametersMetaData, requiresNodes[i], mapTypeParameters)...)
		issues = append(issues, validateMetadata(requires.Properties, requires.PropertiesMetaData, requiresNodes[i], mapTypeProperties)...)
	}
	return issues
}
