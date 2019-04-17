package validate

import (
	"fmt"
	"gopkg.in/yaml.v3"

	"github.com/SAP/cloud-mta/mta"
)

type nameInfo struct {
	object string
	Line   int
}

// isNameUnique - validates the global uniqueness of the names of modules, provided services and resources
func isNameUnique(mta *mta.MTA, mtaNode *yaml.Node, source string) []YamlValidationIssue {
	var issues []YamlValidationIssue
	// map: name -> object kind (module, provided services or resource)
	names := make(map[string]nameInfo)
	for i, module := range mta.Modules {
		line := getModuleLineByIndex(mtaNode, i)
		// validate module name
		issues = validateNameUniqueness(names, module.Name, moduleEntityKind, issues, line)
		for j, provide := range module.Provides {
			setLine := getProvidedSetByIndex(mtaNode, i, j)
			// validate name of provided service
			issues = validateNameUniqueness(names, provide.Name, providedPropEntityKind, issues, setLine)
		}
	}
	for i, resource := range mta.Resources {
		line := getResourceLineByIndex(mtaNode, i)
		// validate resource name
		issues = validateNameUniqueness(names, resource.Name, "resource", issues, line)
	}
	return issues
}

func getModuleLineByIndex(mtaNode *yaml.Node, index int) int {
	modulesNode := getPropValueByName(mtaNode, modulesYamlField)
	return getIndexedNodePropLine(modulesNode, index, nameYamlField)
}

func getResourceLineByIndex(mtaNode *yaml.Node, index int) int {
	resourcesNode := getPropValueByName(mtaNode, resourcesYamlField)
	return getIndexedNodePropLine(resourcesNode, index, nameYamlField)
}

func getProvidedSetByIndex(mtaNode *yaml.Node, moduleIndex, providedSetIndex int) int {
	modulesNode := getPropValueByName(mtaNode, modulesYamlField)
	moduleNode := modulesNode.Content[moduleIndex]
	provided := getPropValueByName(moduleNode, providesYamlField)
	return getIndexedNodePropLine(provided, providedSetIndex, nameYamlField)
}

func getIndexedNodePropLine(node *yaml.Node, index int, propName string) int {
	indexedNode := node.Content[index]
	nameNode := getPropValueByName(indexedNode, propName)
	if nameNode == nil {
		return 0
	}
	return nameNode.Line
}

// validateNameUniqueness - validate that name not defined already (not exists in the 'names' map)
func validateNameUniqueness(names map[string]nameInfo, name string,
	objectName string, issues []YamlValidationIssue, line int) []YamlValidationIssue {
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
				name, objectName, article, prevObject.object, prevObject.Line), line)
	} else {
		// name not found -> add it to the global map
		names[name] = nameInfo{object: objectName, Line: line}
	}
	return result
}
