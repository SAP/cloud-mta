package validate

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/SAP/cloud-mta/mta"
)

// ifRequiredDefined - validates that required property sets are defined in modules, provided sections or resources
func ifRequiredDefined(mta *mta.MTA, source string) []YamlValidationIssue {
	var issues []YamlValidationIssue

	// init set of all provided property sets
	providedSet := make(map[string]map[string]interface{})

	for _, module := range mta.Modules {
		// add module to provided property sets
		providedSet[module.Name] = collectLeafProperties(module.Properties)
		// add all property sets provided by module
		for _, provided := range module.Provides {
			providedSet[provided.Name] = collectLeafProperties(provided.Properties)
		}
	}

	// add resources to provided property sets
	for _, resource := range mta.Resources {
		providedSet[resource.Name] = collectLeafProperties(resource.Properties)
	}

	for _, module := range mta.Modules {
		issues = append(issues,
			checkRequiredProperties(providedSet, "", module.Properties,
				fmt.Sprintf(`"%s" module`, module.Name))...)
		// check that each required by module property set was provided in mta.yaml
		for _, requires := range module.Requires {
			if _, contains := providedSet[requires.Name]; !contains {
				issues = appendIssue(issues,
					fmt.Sprintf(`the "%s" property set required by the "%s" module is not defined`,
						requires.Name, module.Name))
			}
			// check that each property of module is resolved
			issues = append(issues,
				checkRequiredProperties(providedSet, requires.Name, requires.Properties,
					fmt.Sprintf(`"%s" module`, module.Name))...)
		}
	}

	for _, resource := range mta.Resources {
		issues = append(issues,
			checkRequiredProperties(providedSet, "", resource.Properties,
				fmt.Sprintf(`"%s" resource`, resource.Name))...)
		// check that each required by resource property set was provided in mta.yaml
		for _, requires := range resource.Requires {
			if _, contains := providedSet[requires.Name]; !contains {
				issues = appendIssue(issues,
					fmt.Sprintf(`the "%s" property set required by the "%s" resource is not defined`,
						requires.Name, resource.Name))
			}
			// check that each property of resource is resolved
			issues = append(issues,
				checkRequiredProperties(providedSet, requires.Name, requires.Properties,
					fmt.Sprintf(`"%s" resource`, resource.Name))...)
		}
	}
	return issues
}

// collectLeafProperties - go through all properties, including hierarchical and collect their leafs
func collectLeafProperties(props map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for name, value := range props {
		valueMap, ok := value.(map[interface{}]interface{})
		if ok {
			// property is map
			subProps := collectLeafProperties(convertMap(valueMap))
			// combine name of the property with upper level property name
			for subProp := range subProps {
				result[name+"."+subProp] = nil
			}
		} else {
			// simple property
			result[name] = nil
		}
	}
	return result
}

func convertMap(m map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range m {
		result[key.(string)] = value
	}
	return result
}

func checkRequiredProperties(providedProps map[string]map[string]interface{}, requiredPropSet string,
	requiredProps map[string]interface{}, requiringObject string) []YamlValidationIssue {

	var issues []YamlValidationIssue
	for propName, propValue := range requiredProps {
		issues = append(issues, checkValue(providedProps, propName, requiredPropSet, requiringObject, propValue)...)
	}
	return issues
}

func checkValue(providedProps map[string]map[string]interface{},
	propName, propSet, requiringObject string, propValue interface{}) []YamlValidationIssue {
	var issues []YamlValidationIssue
	propValueStr, ok := propValue.(string)
	if ok {
		// property is simple - check if it can be resolved
		issues = checkStringPropertyValue(providedProps, propName, propValueStr, propSet, requiringObject)
	} else {
		propValueMap, ok := propValue.(map[interface{}]interface{})
		if ok {
			// property is a map
			for key, value := range propValueMap {
				// check every sub property
				issues = append(issues, checkValue(providedProps, propName+"."+key.(string), propSet, requiringObject, value)...)
			}
		}
	}
	return issues
}

func checkStringPropertyValue(providedProps map[string]map[string]interface{},
	propName, propValue, propSet, requiringObject string) []YamlValidationIssue {
	var issues []YamlValidationIssue
	r := regexp.MustCompile(`~{[^{}]+}`)
	// find all placeholders
	matches := r.FindAllString(propValue, -1)
	for _, match := range matches {
		// get placeholder pure name, removing tilda and brackets
		requiredProp := strings.TrimPrefix(strings.TrimSuffix(match, "}"), "~{")
		// if property set was not provided it has to be presented in placeholder
		if propSet == "" {
			// split placeholder to property set and property name
			requiredPropArr := strings.SplitN(requiredProp, ".", 2)
			if len(requiredPropArr) != 2 {
				// no property set provided
				issues = appendIssue(issues,
					fmt.Sprintf(`the "%s" property of the %s is unresolved; the "%s" property is not provided`,
						propName, requiringObject, requiredProp))
			} else {
				// check existence of property if property set
				issues = appendIssue(issues,
					checkRequiredProperty(providedProps, propName, requiredPropArr[0], requiredPropArr[1], requiringObject))
			}
		} else {
			// check existence of property if property set
			issues = appendIssue(issues,
				checkRequiredProperty(providedProps, propName, propSet, requiredProp, requiringObject))
		}

	}
	return issues
}

func checkRequiredProperty(providedProps map[string]map[string]interface{}, property,
requiredSet, requiredProp, requiringObject string) string {
	providedSet, ok := providedProps[requiredSet]
	if ok {
		_, ok = providedSet[requiredProp]
	}
	if !ok {
		return fmt.Sprintf(`the "%s" property of the %s is unresolved; the "%s.%s" property is not provided`,
			property, requiringObject, requiredSet, requiredProp)
	}
	return ""
}
