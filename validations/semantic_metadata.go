package validate

import (
	"fmt"
	"gopkg.in/yaml.v3"

	"github.com/SAP/cloud-mta/mta"
)

const (
	unknownNameInMetadataMsg = `metadata cannot be defined for the "%s" undefined %s`
	emptyRequiredFieldMsg    = `the value for the required and non-overwritable "%s" %s cannot be empty`

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
		issues = append(issues, checkRequiresParamsAndPropertiesMetadata(requiresNode, module.Requires)...)

		hooksNode := getPropContent(modulesNode[i], hooksYamlField)
		for i, hook := range module.Hooks {
			issues = append(issues, checkMetadata(hook.Parameters, hook.ParametersMetaData, hooksNode[i], mapTypeParameters)...)

			requiresNode = getPropContent(hooksNode[i], requiresYamlField)
			issues = append(issues, checkRequiresParamsAndPropertiesMetadata(requiresNode, hook.Requires)...)
		}
	}

	resourcesNode := getPropContent(mtaNode, resourcesYamlField)
	for i, resource := range mta.Resources {
		issues = append(issues, checkMetadata(resource.Parameters, resource.ParametersMetaData, resourcesNode[i], mapTypeParameters)...)
		issues = append(issues, checkMetadata(resource.Properties, resource.PropertiesMetaData, resourcesNode[i], mapTypeProperties)...)

		requiresNode := getPropContent(resourcesNode[i], requiresYamlField)
		issues = append(issues, checkRequiresParamsAndPropertiesMetadata(requiresNode, resource.Requires)...)
	}

	if strict {
		return issues, nil
	}
	return nil, issues
}

func checkRequiresParamsAndPropertiesMetadata(requiresNodes []*yaml.Node, requiresList []mta.Requires) []YamlValidationIssue {
	var issues []YamlValidationIssue

	for i, requires := range requiresList {
		issues = append(issues, checkMetadata(requires.Parameters, requires.ParametersMetaData, requiresNodes[i], mapTypeParameters)...)
		issues = append(issues, checkMetadata(requires.Properties, requires.PropertiesMetaData, requiresNodes[i], mapTypeProperties)...)
	}
	return issues
}

// Check each property/parameter in the metadata is defined in the map, and that non-optional and non-overwritable property/parameter is not nil
func checkMetadata(m map[string]interface{}, metadata map[string]mta.MetaData, parentNode *yaml.Node, mapType int) []YamlValidationIssue {
	metadataNode := getPropValueByName(parentNode, mapTypes[mapType].metadataNodeName)
	mapNode := getPropValueByName(parentNode, mapTypes[mapType].mapNodeName)

	var issues []YamlValidationIssue
	for key := range metadata {
		_, ok := m[key]
		if !ok {
			keyNode := getPropByName(metadataNode, key)
			issues = append(issues, YamlValidationIssue{Msg: fmt.Sprintf(unknownNameInMetadataMsg, key, mapTypes[mapType].entityKind), Line: keyNode.Line})
		}
	}
	if metadata != nil {
		for key, value := range m {
			// If there's no metadata for the key we don't perform this check (since it's overwritable by default)
			if meta, ok := metadata[key]; ok {
				if !meta.Optional && !meta.OverWritable && value == nil {
					keyNode := getPropByName(mapNode, key)
					issues = append(issues, YamlValidationIssue{Msg: fmt.Sprintf(emptyRequiredFieldMsg, key, mapTypes[mapType].entityKind), Line: keyNode.Line})
				}
			}
		}
	}
	return issues
}
