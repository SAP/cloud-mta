package validate

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"reflect"
	"regexp"
	"strings"

	"github.com/SAP/cloud-mta/mta"
)

// ifRequiredDefined - validates that required property sets are defined in modules, provided sections or resources
func ifRequiredDefined(mta *mta.MTA, mtaNode *yaml.Node, source string, strict bool) ([]YamlValidationIssue, []YamlValidationIssue) {
	var issues []YamlValidationIssue

	// init set of all provided property sets
	provided := make(map[string]map[string]interface{})

	for _, module := range mta.Modules {
		// add module to provided property sets
		provided[module.Name] = module.Properties
		// add all property sets provided by module
		for _, prov := range module.Provides {
			provided[prov.Name] = prov.Properties
		}
	}

	// init set of all provided property sets with configuration type
	configurationProvided := make(map[string]bool)

	// add resources to provided property sets
	for _, resource := range mta.Resources {
		if resource.Type == configuration {
			configurationProvided[resource.Name] = true
		} else {
			provided[resource.Name] = resource.Properties
		}
	}

	modulesNode := getPropContent(mtaNode, modulesYamlField)
	for i, module := range mta.Modules {
		issues = append(issues, checkComponent(provided, configurationProvided, module, modulesNode[i], "module")...)
		for j, moduleProvides := range module.Provides {
			providesNode := getPropValueByName(modulesNode[i], providesYamlField)
			issues = append(issues, checkComponent(provided, configurationProvided, &moduleProvides, providesNode.Content[j], "provided property set of the "+module.Name+" module")...)
		}
	}

	resourcesNode := getPropContent(mtaNode, resourcesYamlField)
	for i, resource := range mta.Resources {
		issues = append(issues, checkComponent(provided, configurationProvided, resource, resourcesNode[i], "resource")...)
	}
	return issues, nil
}

func structFieldToMap(str interface{}, field string) map[string]interface{} {
	v := reflect.ValueOf(str).Elem().FieldByName(field)
	if v.IsValid() {
		mapValue, ok := v.Addr().Interface().(*map[string]interface{})
		if ok {
			return *mapValue
		}
	}
	return nil
}

func structFieldToRequires(str interface{}) []mta.Requires {
	v := reflect.ValueOf(str).Elem().FieldByName("Requires")
	if v.IsValid() {
		mapValue, ok := v.Addr().Interface().(*[]mta.Requires)
		if ok {
			return *mapValue
		}
	}
	return []mta.Requires{}
}

func structFieldToString(str interface{}) string {
	v := reflect.ValueOf(str).Elem().FieldByName(nameMtaField)
	strValue := v.Addr().Interface().(*string)
	return *strValue
}

func checkComponent(provided map[string]map[string]interface{}, configurationProvided map[string]bool, component interface{}, compNode *yaml.Node, compDesc string) []YamlValidationIssue {
	var issues []YamlValidationIssue

	compName := structFieldToString(component)
	propsNode := getPropValueByName(compNode, propertiesYamlField)
	issues = append(issues,
		checkRequiredProperties(provided, configurationProvided, "", structFieldToMap(component, propertiesMtaField),
			fmt.Sprintf(`"%s" %s`, compName, compDesc), propsNode, propertyEntityKind)...)
	paramsNode := getPropValueByName(compNode, parametersYamlField)
	issues = append(issues,
		checkRequiredProperties(provided, configurationProvided, "", structFieldToMap(component, parametersMtaField),
			fmt.Sprintf(`"%s" %s`, compName, compDesc), paramsNode, parameterEntityKind)...)
	buildParamsNode := getPropValueByName(compNode, buildParametersYamlField)
	issues = append(issues,
		checkRequiredProperties(provided, configurationProvided, "", structFieldToMap(component, buildParametersMtaField),
			fmt.Sprintf(`"%s" %s`, compName, compDesc), buildParamsNode, buildParamEntityKind)...)
	// check that each required by resource property set was provided in mta.yaml
	requiresNode := getPropValueByName(compNode, requiresYamlField)
	for i, requires := range structFieldToRequires(component) {
		_, contains := provided[requires.Name]
		_, containsConfiguration := configurationProvided[requires.Name]
		if !contains && !containsConfiguration {
			reqNameNode := getPropValueByName(requiresNode.Content[i], nameYamlField)
			issues = appendIssue(issues,
				fmt.Sprintf(`the "%s" property set required by the "%s" %s is not defined`,
					requires.Name, compName, compDesc), reqNameNode.Line, reqNameNode.Column)
		}
		// check that each property of resource is resolved
		reqPropsNode := getPropValueByName(requiresNode.Content[i], propertiesYamlField)
		issues = append(issues,
			checkRequiredProperties(provided, configurationProvided, requires.Name, requires.Properties,
				fmt.Sprintf(`"%s" %s`, compName, compDesc), reqPropsNode, propertyEntityKind)...)
		// check that each parameter of resource is resolved
		reqParamsNode := getPropValueByName(requiresNode.Content[i], parametersYamlField)
		issues = append(issues,
			checkRequiredProperties(provided, configurationProvided, requires.Name, requires.Parameters,
				fmt.Sprintf(`"%s" %s`, compName, compDesc), reqParamsNode, parameterEntityKind)...)
	}
	return issues
}

func checkRequiredProperties(providedProps map[string]map[string]interface{}, configurationProvided map[string]bool, requiredPropSet string,
	requiredEntities map[string]interface{}, requiringObject string, node *yaml.Node, entityKind string) []YamlValidationIssue {

	var issues []YamlValidationIssue
	if requiredEntities == nil {
		return nil
	}
	for entityName, entityValue := range requiredEntities {
		entityNode := getPropValueByName(node, entityName)
		issues = append(issues, checkValue(providedProps, configurationProvided, entityName, entityKind, requiredPropSet, requiringObject, entityValue, entityNode)...)
	}
	return issues
}

func checkValue(providedProps map[string]map[string]interface{}, configurationProvided map[string]bool,
	entityName, entityKind, propSet, requiringObject string, entityValue interface{}, node *yaml.Node) []YamlValidationIssue {
	var issues []YamlValidationIssue
	propValueStr, ok := entityValue.(string)
	if ok {
		// property is simple - check if it can be resolved
		issues = checkStringEntityValue(providedProps, configurationProvided, entityName, propValueStr, entityKind, propSet, requiringObject, node)
	} else {
		propValueMap, ok := entityValue.(map[string]interface{})
		if ok {
			// property is a map
			for key, value := range propValueMap {
				childNode := getPropValueByName(node, key)
				// check every sub property
				issues = append(issues, checkValue(providedProps, configurationProvided, entityName+"."+key, entityKind, propSet, requiringObject, value, childNode)...)
			}
		}
	}
	return issues
}

func checkStringEntityValue(providedProps map[string]map[string]interface{}, configurationProvided map[string]bool,
	entityName, entityValue, entityKind, propSet, requiringObject string, entityNode *yaml.Node) []YamlValidationIssue {
	var issues []YamlValidationIssue
	r := regexp.MustCompile(`~{[^{}]+}`)
	// find all placeholders
	matches := r.FindAllString(entityValue, -1)
	for _, match := range matches {
		// get placeholder pure name, removing tilda and brackets
		requiredProp := strings.TrimPrefix(strings.TrimSuffix(match, "}"), "~{")
		// if property set was not provided it has to be presented in placeholder
		if propSet == "" {
			// split placeholder to property set and property name
			requiredPropArr := strings.SplitN(requiredProp, "/", 2)
			if len(requiredPropArr) != 2 {
				// no property set provided
				issues = appendIssue(issues,
					fmt.Sprintf(`the "%s" %s of the %s is unresolved; the "%s" property is not provided`,
						entityName, entityKind, requiringObject, requiredProp), entityNode.Line, entityNode.Column)
			} else {
				// check existence of property if property set
				issues = appendIssue(issues,
					checkRequiredProperty(providedProps, configurationProvided, entityName, entityKind, requiredPropArr[0], requiredPropArr[1], requiringObject), entityNode.Line, entityNode.Column)
			}
		} else {
			// check existence of property if property set
			issues = appendIssue(issues,
				checkRequiredProperty(providedProps, configurationProvided, entityName, entityKind, propSet, requiredProp, requiringObject), entityNode.Line, entityNode.Column)
		}

	}
	return issues
}

func checkRequiredProperty(providedProps map[string]map[string]interface{}, configurationProvided map[string]bool, entityName, entityKind,
	requiredSet, requiredProp, requiringObject string) string {
	providedSet, ok := providedProps[requiredSet]
	if ok {
		_, ok = providedSet[requiredProp]
	}
	if !ok {
		_, ok = configurationProvided[requiredSet]
	}
	if !ok {
		return fmt.Sprintf(`the "%s" %s of the %s is unresolved; the "%s/%s" property is not provided`,
			entityName, entityKind, requiringObject, requiredSet, requiredProp)
	}
	return ""
}
