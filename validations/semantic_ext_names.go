package validate

import (
	"fmt"
	"gopkg.in/yaml.v3"

	"github.com/SAP/cloud-mta/mta"
)

const (
	nameAlreadyExtendedMsg = `the "%s" %s has already been extended in this file; %s %s with the same name was found on line %d`
)

// checkSingleExtendNames validates that each object (module, resource, hook inside module, provides inside module, requires inside parent wherever used) is extended only once
func checkSingleExtendNames(mta *mta.EXT, root *yaml.Node, source string, strict bool) ([]YamlValidationIssue, []YamlValidationIssue) {
	var issues []YamlValidationIssue
	// map: name -> object kind (module, provided services or resource) and line
	moduleNames := make(map[string]nameInfo)
	for i, module := range mta.Modules {
		moduleNode := getNamedObjectNodeByIndex(root, modulesYamlField, i)
		line, column := getNamedObjectPositionByIndex(root, modulesYamlField, i)
		// validate module name
		issues = validateNameIsExtendedOnce(moduleNames, module.Name, moduleEntityKind, issues, line, column)

		providesNames := make(map[string]nameInfo)
		for j, provide := range module.Provides {
			providesLine, providesColumn := getNamedObjectPositionByIndex(moduleNode, providesYamlField, j)
			// validate name of provided service
			issues = validateNameIsExtendedOnce(providesNames, provide.Name, providedPropEntityKind, issues, providesLine, providesColumn)
		}

		// validate requires
		issues = validateRequiresIsExtendedOnce(module.Requires, moduleNode, issues)

		hookNames := make(map[string]nameInfo)
		for j, hook := range module.Hooks {
			hookNode := getNamedObjectNodeByIndex(moduleNode, hooksYamlField, j)
			hookLine, hookColumn := getNamedObjectPositionByIndex(moduleNode, hooksYamlField, j)
			// validate hook name
			issues = validateNameIsExtendedOnce(hookNames, hook.Name, hookPropEntityKind, issues, hookLine, hookColumn)
			// validate requires
			issues = validateRequiresIsExtendedOnce(hook.Requires, hookNode, issues)
		}
	}

	resourceNames := make(map[string]nameInfo)
	for i, resource := range mta.Resources {
		resourceNode := getNamedObjectNodeByIndex(root, resourcesYamlField, i)
		line, column := getNamedObjectPositionByIndex(root, resourcesYamlField, i)
		// validate resource name
		issues = validateNameIsExtendedOnce(resourceNames, resource.Name, resourceEntityKind, issues, line, column)
		// validate requires
		issues = validateRequiresIsExtendedOnce(resource.Requires, resourceNode, issues)
	}
	return issues, nil
}

func validateRequiresIsExtendedOnce(requiresList []mta.Requires, parentNode *yaml.Node, issues []YamlValidationIssue) []YamlValidationIssue {
	requiresNames := make(map[string]nameInfo)
	for i, requires := range requiresList {
		requiresLine, requiresColumn := getNamedObjectPositionByIndex(parentNode, requiresYamlField, i)
		issues = validateNameIsExtendedOnce(requiresNames, requires.Name, requiresPropEntityKind, issues, requiresLine, requiresColumn)
	}
	return issues
}

// validateNameIsExtendedOnce - validate that name not defined already (not exists in the 'names' map)
func validateNameIsExtendedOnce(names map[string]nameInfo, name string,
	objectName string, issues []YamlValidationIssue, line int, column int) []YamlValidationIssue {
	result := issues
	// try to find name in the global map
	prevObject, ok := names[name]
	// name found -> add issue
	if ok {
		article := "a"
		if objectName == prevObject.object {
			article = "another"
		}
		result = appendIssue(result,
			fmt.Sprintf(nameAlreadyExtendedMsg, name, objectName, article, prevObject.object, prevObject.Line), line, column)
	} else {
		// name not found -> add it to the global map
		names[name] = nameInfo{object: objectName, Line: line, Column: column}
	}
	return result
}
