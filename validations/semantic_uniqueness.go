package validate

import (
	"fmt"
	"gopkg.in/yaml.v3"

	"github.com/SAP/cloud-mta/mta"
)

type nameInfo struct {
	object string
	Line   int
	Column int
}

// isNameUnique - validates the global uniqueness of the names of modules, provided services and resources
func isNameUnique(mta *mta.MTA, mtaNode *yaml.Node, source string, strict bool) ([]YamlValidationIssue, []YamlValidationIssue) {
	var issues []YamlValidationIssue
	// map: name -> object kind (module, provided services or resource)
	names := make(map[string]nameInfo)
	for i, module := range mta.Modules {
		line, column := getModulePositionByIndex(mtaNode, i)
		// validate module name
		issues = validateNameUniqueness(names, module.Name, moduleEntityKind, issues, line, column)
		for j, provide := range module.Provides {
			setLine, setColumn := getProvidedSetPositionByIndex(mtaNode, i, j)
			// validate name of provided service
			issues = validateNameUniqueness(names, provide.Name, providedPropEntityKind, issues, setLine, setColumn)
		}
	}
	for i, resource := range mta.Resources {
		line, column := getResourcePositionByIndex(mtaNode, i)
		// validate resource name
		issues = validateNameUniqueness(names, resource.Name, "resource", issues, line, column)
	}
	return issues, nil
}

func getModulePositionByIndex(mtaNode *yaml.Node, index int) (line int, column int) {
	return getNamedObjectPositionByIndex(mtaNode, modulesYamlField, index)
}

func getResourcePositionByIndex(mtaNode *yaml.Node, index int) (line int, column int) {
	return getNamedObjectPositionByIndex(mtaNode, resourcesYamlField, index)
}

func getProvidedSetPositionByIndex(mtaNode *yaml.Node, moduleIndex, providedSetIndex int) (line int, column int) {
	moduleNode := getNamedObjectNodeByIndex(mtaNode, modulesYamlField, moduleIndex)
	provided := getPropValueByName(moduleNode, providesYamlField)
	line, column, _ = getIndexedNodePropPosition(provided, providedSetIndex, nameYamlField)
	return line, column
}

// validateNameUniqueness - validate that name not defined already (not exists in the 'names' map)
func validateNameUniqueness(names map[string]nameInfo, name string,
	objectName string, issues []YamlValidationIssue, line int, column int) []YamlValidationIssue {
	result := issues
	// try to find name in the global map
	prevObject, ok := names[name]
	// name found -> add issue
	if ok {
		var article string
		if objectName == prevObject.object {
			article = "another"
		} else {
			article = "a"
		}
		result = appendIssue(result,
			fmt.Sprintf(`the "%s" %s name is already in use; %s %s was found with the same name on line %d`,
				name, objectName, article, prevObject.object, prevObject.Line), line, column)
	} else {
		// name not found -> add it to the global map
		names[name] = nameInfo{object: objectName, Line: line, Column: column}
	}
	return result
}
