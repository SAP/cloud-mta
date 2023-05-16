package resolver

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"

	"github.com/SAP/cloud-mta/internal/logs"
	"github.com/SAP/cloud-mta/mta"
)

const (
	emptyModuleNameMsg = "provide a name for the module"
	moduleNotFoundMsg  = `could not find the "%s" module`
	marshalFailsMag    = `could not marshal the "%s" environment variable`
	missingPrefixMsg   = `could not resolve the value for the "~{%s}" variable; missing required prefix`

	defaultEnvFileName = ".env"
)

var envGetter = os.Environ

// ResolveResult is the result of the Resolve function. This is serialized to json when requested.
type ResolveResult struct {
	Properties map[string]string `json:"properties"`
	Messages   []string          `json:"messages"`
}

// Resolve - resolve module's parameters
func Resolve(workspaceDir, moduleName, path string, extensions []string, envFile string) (result ResolveResult, messages []string, err error) {
	if len(moduleName) == 0 {
		return result, nil, errors.New(emptyModuleNameMsg)
	}
	mtaRaw, messages, err := mta.GetMtaFromFile(path, extensions, false)
	if err != nil {
		return result, messages, err
	}
	if len(workspaceDir) == 0 {
		workspaceDir = filepath.Dir(path)
	}

	// If environment file name is not provided - set the default file name to .env
	envFilePath := defaultEnvFileName
	if len(envFile) > 0 {
		envFilePath = envFile
	}

	m := NewMTAResolver(mtaRaw, workspaceDir)

	for _, module := range m.GetModules() {
		if module.Name == moduleName {
			m.ResolveProperties(module, envFilePath)

			propVarMap, err := getPropertiesAsEnvVar(module)
			if err != nil {
				return result, messages, err
			}
			result.Properties = propVarMap
			result.Messages = m.messages
			return result, messages, nil
		}
	}

	return result, messages, errors.Errorf(moduleNotFoundMsg, moduleName)
}

func getPropertiesAsEnvVar(module *mta.Module) (map[string]string, error) {
	envVar := map[string]interface{}{}
	for key, val := range module.Properties {
		envVar[key] = val
	}

	for _, requires := range module.Requires {
		propMap := envVar
		if len(requires.Group) > 0 {
			propMap = map[string]interface{}{}
		}

		for key, val := range requires.Properties {
			propMap[key] = val
		}

		if len(requires.Group) > 0 {
			//append the array element to group
			group, ok := envVar[requires.Group]
			if ok {
				groupArray := group.([]map[string]interface{})
				envVar[requires.Group] = append(groupArray, propMap)
			} else {
				envVar[requires.Group] = []map[string]interface{}{propMap}
			}
		}
	}

	//serialize
	return serializePropertiesAsEnvVars(envVar)
}

func serializePropertiesAsEnvVars(envVar map[string]interface{}) (map[string]string, error) {
	retEnvVar := map[string]string{}
	for key, val := range envVar {
		switch v := val.(type) {
		case string:
			retEnvVar[key] = v
		default:
			bytesVal, err := json.Marshal(val)
			if err != nil {
				return nil, errors.Errorf(marshalFailsMag, key)
			}
			retEnvVar[key] = string(bytesVal)
		}
	}

	return retEnvVar, nil
}

// MTAResolver is used to resolve MTA properties' variables
type MTAResolver struct {
	mta.MTA
	WorkingDir string
	context    *ResolveContext
	messages   []string
}

const resourceType = 1
const moduleType = 2

const variablePrefix = "~"
const placeholderPrefix = "$"

type mtaSource struct {
	Name       string
	Parameters map[string]interface{} `yaml:"parameters,omitempty"`
	Properties map[string]interface{} `yaml:"properties,omitempty"`
	Type       int
	Module     *mta.Module
	Resource   *mta.Resource
}

// NewMTAResolver is a factory function for MTAResolver
func NewMTAResolver(m *mta.MTA, workspaceDir string) *MTAResolver {
	resolver := &MTAResolver{*m, workspaceDir, &ResolveContext{
		global:    map[string]string{},
		modules:   map[string]map[string]string{},
		resources: map[string]map[string]string{},
	}, []string{}}

	for _, module := range m.Modules {
		resolver.context.modules[module.Name] = map[string]string{}
	}

	for _, resource := range m.Resources {
		resolver.context.resources[resource.Name] = map[string]string{}
	}
	return resolver
}

func resolvePath(path string, parts ...string) string {
	absolutePath := path
	if !filepath.IsAbs(path) {
		absolutePath = filepath.Join(append(parts, absolutePath)...)
	}
	return absolutePath
}

// ResolveProperties is the main function to trigger the resolution
func (m *MTAResolver) ResolveProperties(module *mta.Module, envFilePath string) {

	if m.Parameters == nil {
		m.Parameters = map[string]interface{}{}
	}

	//add env variables
	for _, val := range envGetter() {
		pos := strings.Index(val, "=")
		if pos > 0 {
			key := strings.Trim(val[:pos], " ")
			value := strings.Trim(val[pos+1:], " ")
			m.addValueToContext(key, value)
		}
	}

	//add .env file in module's path to the module context
	if len(module.Path) > 0 {
		envFile := resolvePath(envFilePath, m.WorkingDir, module.Path)
		envMap, err := godotenv.Read(envFile)
		if err == nil {
			for key, value := range envMap {
				m.addValueToContext(key, value)
			}
		}
	}
	m.addServiceNames(module)

	//top level properties
	for key, value := range module.Properties {
		//no expected variables
		propValue := m.resolve(module, nil, value)
		module.Properties[key] = m.resolvePlaceholders(module, nil, nil, propValue)
	}

	//required properties:
	for _, req := range module.Requires {
		requiredSource := m.findProvider(req.Name)
		for propName, PropValue := range req.Properties {
			resolvedValue := m.resolve(module, &req, PropValue)
			//replace value with resolved value
			req.Properties[propName] = m.resolvePlaceholders(module, requiredSource, &req, resolvedValue)
		}
	}
}

func (m *MTAResolver) addValueToContext(key, value string) {
	//if the key has format of "module/key", or "resource/key" writes the value to the module's context
	slashPos := strings.Index(key, "/")
	if slashPos > 0 {
		modName := key[:slashPos]
		key = key[slashPos+1:]
		modulesContext, ok := m.context.modules[modName]
		if !ok {
			modulesContext, ok = m.context.resources[modName]
		}
		if ok {
			modulesContext[key] = value
		}
	} else {
		m.context.global[key] = value
	}

}

func (m *MTAResolver) resolve(sourceModule *mta.Module, requires *mta.Requires, valueObj interface{}) interface{} {
	switch valueObj := valueObj.(type) {
	case map[interface{}]interface{}:
		v := convertToJSONSafe(valueObj)
		return m.resolve(sourceModule, requires, v)
	case map[string]interface{}:
		for k, v := range valueObj {
			valueObj[k] = m.resolve(sourceModule, requires, v)
		}
		return valueObj
	case []interface{}:
		for i, v := range valueObj {
			valueObj[i] = m.resolve(sourceModule, requires, v)
		}
		return valueObj
	case string:
		return m.resolveString(sourceModule, requires, valueObj)
	default:
		//if the value is not a string but a leaf, just return it
		return valueObj
	}

}

func (m *MTAResolver) resolveString(sourceModule *mta.Module, requires *mta.Requires, value string) interface{} {
	pos := 0

	pos, variableName, wholeValue := parseNextVariable(pos, value, variablePrefix)
	if pos < 0 {
		//no variables
		return value
	}
	varValue := m.getVariableValue(sourceModule, requires, variableName)

	if wholeValue {
		return varValue
	}
	for pos >= 0 {
		varValueStr, _ := convertToString(varValue)
		value = value[:pos] + varValueStr + value[pos+len(variableName)+3:]

		pos, variableName, _ = parseNextVariable(pos+len(varValueStr), value, variablePrefix)
		if pos >= 0 {
			varValue = m.getVariableValue(sourceModule, requires, variableName)
		}
	}

	return value
}

func convertToString(valueObj interface{}) (string, bool) {
	switch v := valueObj.(type) {
	case string:
		return v, false
	}
	valueBytes, err := json.Marshal(convertToJSONSafe(valueObj))
	if err != nil {
		logs.Logger.Error(err)
		return "", false
	}
	return string(valueBytes), true
}

// return start position, name of variable and if it is a whole value
func parseNextVariable(pos int, value string, prefix string) (int, string, bool) {

	endSign := "}"
	posStart := strings.Index(value[pos:], prefix+"{")
	if posStart < 0 {
		return -1, "", false
	}
	posStart += pos

	if string(value[posStart+2]) == "{" {
		endSign = "}}"
	}

	posEnd := strings.Index(value[posStart+2:], endSign)
	if posEnd < 0 {
		//bad value
		return -1, "", false
	}
	posEnd += posStart + 1 + len(endSign)
	wholeValue := posStart == 0 && posEnd == len(value)-1

	return posStart, value[posStart+2 : posEnd], wholeValue
}

func (m *MTAResolver) getVariableValue(sourceModule *mta.Module, requires *mta.Requires, variableName string) interface{} {
	var providerName string
	if requires == nil {
		slashPos := strings.Index(variableName, "/")
		if slashPos > 0 {
			providerName = variableName[:slashPos]
			variableName = variableName[slashPos+1:]
		} else {
			m.addMessage(fmt.Sprintf(missingPrefixMsg, variableName))
			return "~{" + variableName + "}"
		}

	} else {
		providerName = requires.Name
	}

	source := m.findProvider(providerName)
	if source != nil {
		for propName, propValue := range source.Properties {
			if propName == variableName {

				//Do not pass module and requires, because it is a wrong scope
				//it is either global->module->requires
				//or           global->resource
				propValue = m.resolvePlaceholders(nil, source, nil, propValue)
				return convertToJSONSafe(propValue)
			}
		}
	}

	if source != nil && source.Type == resourceType && source.Resource.Type == "configuration" {
		provID, ok := getStringFromMap(source.Resource.Parameters, "provider-id")
		if ok {
			m.addMessage(fmt.Sprint("Missing configuration ", provID, "/", variableName))
		}
	}

	return "~{" + variableName + "}"
}

func (m *MTAResolver) resolvePlaceholders(sourceModule *mta.Module, source *mtaSource, requires *mta.Requires, valueObj interface{}) interface{} {
	switch valueObj := valueObj.(type) {
	case map[interface{}]interface{}:
		v := convertToJSONSafe(valueObj)
		return m.resolvePlaceholders(sourceModule, source, requires, v)
	case map[string]interface{}:
		for k, v := range valueObj {
			valueObj[k] = m.resolvePlaceholders(sourceModule, source, requires, v)
		}
		return valueObj
	case []interface{}:
		for k, v := range valueObj {
			valueObj[k] = m.resolvePlaceholders(sourceModule, source, requires, v)
		}
		return valueObj
	case string:
		return m.resolvePlaceholdersString(sourceModule, source, requires, valueObj)
	default:
		//if the value is not a string but a leaf, just return it
		return valueObj
	}
}

func (m *MTAResolver) resolvePlaceholdersString(sourceModule *mta.Module, source *mtaSource, requires *mta.Requires, value string) interface{} {
	pos := 0
	pos, placeholderName, wholeValue := parseNextVariable(pos, value, placeholderPrefix)

	if pos < 0 {
		return value
	}
	placeholderValue := m.getParameter(sourceModule, source, requires, placeholderName)

	if wholeValue {
		return placeholderValue
	}
	for pos >= 0 {
		phValueStr, _ := convertToString(placeholderValue)
		value = value[:pos] + phValueStr + value[pos+len(placeholderName)+3:]
		pos, placeholderName, _ = parseNextVariable(pos+len(phValueStr), value, placeholderPrefix)
		if pos >= 0 {
			placeholderValue = m.getParameter(sourceModule, source, requires, placeholderName)
		}
	}

	return value
}

func (m *MTAResolver) getParameterFromSource(source *mtaSource, paramName string) string {
	if source != nil {
		// See if the value was configured externally first (in VCAP_SERVICES, env var etc)
		// The source can be a module or a resource
		module, found := m.context.modules[source.Name]
		if found {
			paramValStr, ok := module[paramName]
			if ok {
				return paramValStr
			}
		}

		resource, found := m.context.resources[source.Name]
		if found {
			paramValStr, ok := resource[paramName]
			if ok {
				return paramValStr
			}
		}

		// If it was not defined externally, try to get it from the source parameters
		paramVal, found := getStringFromMap(source.Parameters, paramName)
		if found {
			return paramVal
		}

	}
	return ""
}

func (m *MTAResolver) getParameter(sourceModule *mta.Module, source *mtaSource, requires *mta.Requires, paramName string) string {
	//first on source parameters scope
	paramValStr := m.getParameterFromSource(source, paramName)

	//first on source parameters scope
	if paramValStr != "" {
		return paramValStr
	}

	//then try on requires level
	if requires != nil {
		paramVal, ok := getStringFromMap(requires.Parameters, paramName)
		if ok {
			return paramVal
		}
	}

	if sourceModule != nil {
		paramVal, ok := getStringFromMap(sourceModule.Parameters, paramName)
		if ok {
			return paramVal
		}
		//defaults to context's module params:
		paramValStr, ok = m.context.modules[sourceModule.Name][paramName]
		if ok {
			return paramValStr
		}
	}

	//then on MTA root scope
	paramVal, ok := getStringFromMap(m.Parameters, paramName)
	if ok {
		return paramVal
	}

	//then global scope
	paramValStr, ok = m.context.global[paramName]
	if ok {
		return paramValStr
	}

	if source == nil {
		m.addMessage(fmt.Sprint("Missing ", paramName))
	} else {
		m.addMessage(fmt.Sprint("Missing ", source.Name+"/"+paramName))
	}

	return "${" + paramName + "}"
}

func (m *MTAResolver) findProvider(name string) *mtaSource {
	for _, module := range m.Modules {
		for _, provides := range module.Provides {
			if provides.Name == name {
				source := mtaSource{Name: module.Name, Properties: provides.Properties, Parameters: nil, Type: moduleType, Module: module}
				return &source
			}
		}
	}

	//in case of resource, its name is the matching to the requires name
	for _, resource := range m.Resources {
		if resource.Name == name {
			source := mtaSource{Name: resource.Name, Properties: resource.Properties, Parameters: resource.Parameters, Type: resourceType, Resource: resource}
			return &source
		}

	}
	return nil
}

func (m *MTAResolver) addMessage(message string) {
	// This check is necessary so the same message won't be written twice.
	// This happens when a placeholder references a parameter that is not defined,
	// because we try to resolve the parameter while resolving the placeholder and then
	// we try to resolve the parameter again as a parameter.
	if !containsString(m.messages, message) {
		m.messages = append(m.messages, message)
	}
}

func containsString(slice []string, value string) bool {
	for _, curr := range slice {
		if curr == value {
			return true
		}
	}
	return false
}

func convertToJSONSafe(val interface{}) interface{} {
	switch v := val.(type) {
	case map[interface{}]interface{}:
		res := map[string]interface{}{}
		for k, v := range v {
			res[fmt.Sprint(k)] = convertToJSONSafe(v)
		}
		return res
	case []interface{}:
		for k, v2 := range v {
			v[k] = convertToJSONSafe(v2)
		}
		return v
	}
	return val
}

func getStringFromMap(params map[string]interface{}, key string) (string, bool) {
	// Only return the parameter value if it's a string, to prevent a panic.
	// Note: this is used mainly for parameter values during resolve.
	// The deployer DOES support non-string parameters, both as the whole value
	// (it keeps the same type) and inside a string.
	// It stringifies the value in side a string but it's not the usual json stringify.
	// For example, if we have this string:
	//   prop_from_resource: "this is the prop: ~{some_prop}"
	// And some_prop is defined like this:
	//   stuct_field: abc
	// We will get a resolved value like this from the Deployer:
	// "this is the prop: {stuct_field=abc}"
	// We do not support this use case currently.
	value, ok := params[key]
	if ok && value != nil {
		str, isString := value.(string)
		if isString {
			return str, true
		}
	}
	return "", false
}
