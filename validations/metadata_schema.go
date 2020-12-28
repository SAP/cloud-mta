package validate

import (
	"fmt"
	"gopkg.in/yaml.v3"

	"github.com/SAP/cloud-mta/mta"
)

func checkMetadataSchema(mta *mta.MTA, mtaNode *yaml.Node, source string) []YamlValidationIssue {
	return validateMetadata(mta, mtaNode, source, checkNoDatatypeInParametersMetadata)
}

// Check the "datatype" field is not defined in parameters-metadata
func checkNoDatatypeInParametersMetadata(m map[string]interface{}, metadata map[string]mta.MetaData, parentNode *yaml.Node, mapType int) []YamlValidationIssue {
	if mapType != mapTypeParameters {
		return nil
	}

	var issues []YamlValidationIssue
	metadataNode := getPropValueByName(parentNode, parametersMetadataField)

	for key := range metadata {
		valueNode := getPropValueByName(metadataNode, key)
		datatypeKeyNode := getPropByName(valueNode, datatypeYamlField)
		if datatypeKeyNode != nil {
			issues = append(issues, YamlValidationIssue{Msg: fmt.Sprintf(propertyExistsErrorMsg, datatypeYamlField, parametersMetadataField), Line: datatypeKeyNode.Line, Column: datatypeKeyNode.Column})
		}
	}

	return issues
}
