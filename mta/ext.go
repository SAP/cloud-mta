package mta

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"path/filepath"
	"strings"

	"github.com/SAP/cloud-mta/internal/fs"
)

const (
	extUnmarshalErrorMsg      = `the "%s" file is not a valid MTA extension descriptor`
	mergeExtPathErrorMsg      = `could not merge the "%s" MTA extension`
	duplicateExtendsMsg       = `more than 1 extension descriptor file ("%s", "%s", ...) extends the same ID ("%s")`
	extensionIDSameAsMtaIDMsg = `the "%s" extension descriptor file has the same ID ("%s") as the "%s" file`
	duplicateExtensionIDMsg   = `more than 1 extension descriptor file ("%s", "%s", ...) has the same ID ("%s")`
	extendsMsg                = `the "%s" file extends "%s"`
	unknownExtendsMsg         = `some MTA extension descriptors extend unknown IDs: %s`

	versionMismatchMsg = `the "%s" schema version found in the "%s" MTA extension descriptor file does not match the "%s" schema version found in the MTA descriptor`

	mergeRootParametersErrorMsg = `could not merge parameters`

	mergeModulePropertiesErrorMsg             = `could not merge the properties of the "%s" module`
	mergeModuleParametersErrorMsg             = `could not merge the parameters of the "%s" module`
	mergeModuleBuildParametersErrorMsg        = `could not merge the build parameters of the "%s" module`
	mergeModuleIncludesErrorMsg               = `could not merge the 'includes' of the "%s" module`
	mergeModuleProvidesPropertiesErrorMsg     = `could not merge the properties for "%s" in the 'provides' section of the "%s" module`
	mergeModuleRequiresPropertiesErrorMsg     = `could not merge the properties for "%s" in the 'requires' section of the "%s" module`
	mergeModuleRequiresParametersErrorMsg     = `could not merge the parameters for "%s" in the 'requires' section of the "%s" module`
	mergeModuleHookParametersErrorMsg         = `could not merge the parameters of the "%s" hook in the "%s" module`
	mergeModuleHookRequiresPropertiesErrorMsg = `could not merge the properties for "%s" in the 'requires' section of the "%s" hook of the "%s" module`
	mergeModuleHookRequiresParametersErrorMsg = `could not merge the parameters for "%s" in the 'requires' section of the "%s" hook of the "%s" module`
	unknownModuleHookRequiresErrorMsg         = `"%s" in the 'requires' section of the "%s" hook of the "%s" module is defined in the MTA extension but not in the "mta.yaml" file`
	unknownModuleProvidesErrorMsg             = `"%s" in the 'provides' section of the "%s" module is defined in the MTA extension but not in the "mta.yaml" file`
	unknownModuleRequiresErrorMsg             = `"%s" in the 'requires' section of the "%s" module is defined in the MTA extension but not in the "mta.yaml" file`
	unknownModuleHookErrorMsg                 = `the "%s" hook in the "%s" module is defined in the MTA extension but not in the "mta.yaml" file`
	unknownModuleErrorMsg                     = `the "%s" module is defined in the MTA extension but not in the "mta.yaml" file`

	mergeResourceActiveErrorMsg             = `could not merge the 'active' property of the "%s" resource`
	mergeResourcePropertiesErrorMsg         = `could not merge the properties of the "%s" resource`
	mergeResourceParametersErrorMsg         = `could not merge the parameters of the "%s" resource`
	mergeResourceRequiresPropertiesErrorMsg = `could not merge the properties of "%s" in the 'requires' section of the "%s" resource`
	mergeResourceRequiresParametersErrorMsg = `could not merge the parameters of "%s" in the 'requires' section of the "%s" resource`
	unknownResourceRequiresErrorMsg         = `"%s" in the 'requires' section of the "%s" resource is defined in the MTA extension but not in the "mta.yaml" file`
	unknownResourceErrorMsg                 = `the "%s" resource is defined in the MTA extension but not in the "mta.yaml" file`

	overwriteStructuredWithScalarErrorMsg = `"%s": could not overwrite a structured value with a scalar value`
	overwriteScalarWithStructuredErrorMsg = `"%s": could not overwrite a scalar value with a structured value`
	overwriteNonOverwritableErrorMsg      = `the "%s" field cannot be overwritten`
)

// UnmarshalExt returns a reference to the EXT object from a byte array.
func UnmarshalExt(content []byte) (*EXT, error) {
	dec := yaml.NewDecoder(bytes.NewReader(content))
	dec.KnownFields(true)
	mtaExt := EXT{}
	err := dec.Decode(&mtaExt)
	return &mtaExt, err
}

func parseExtFile(extPath string) (*EXT, error) {
	mtaExtContent, err := fs.ReadFile(filepath.Join(extPath))
	if err != nil {
		return nil, err
	}
	mtaExt, err := UnmarshalExt(mtaExtContent)
	if err != nil {
		return nil, errors.Wrapf(err, extUnmarshalErrorMsg, extPath)
	}
	return mtaExt, nil
}

type extensionDetails struct {
	fileName string
	ext      *EXT
}

type ExtensionError struct {
	FileName     string
	err          error
	IsParseError bool
}

func (e ExtensionError) Error() string {
	return e.err.Error()
}

// mergeWithExtensionFiles merges the extensions in the order of the 'extends' chain.
// The extends chain, and the ID and schema version of each mtaext file is validated.
func mergeWithExtensionFiles(mta *MTA, extensions []string, mtaPath string) *ExtensionError {
	extensionsDetails, extErr := getSortedExtensions(extensions, mta.ID, mtaPath)
	if extErr != nil {
		return extErr
	}

	for _, extDetails := range extensionsDetails {
		err := checkSchemaVersionMatches(mta, extDetails)
		if err != nil {
			return &ExtensionError{extDetails.fileName, err, false}
		}
		err = Merge(mta, extDetails.ext, extDetails.fileName)
		if err != nil {
			return &ExtensionError{extDetails.fileName, err, false}
		}
	}
	return nil
}

func getSortedExtensions(extensionFileNames []string, mtaID string, mtaPath string) ([]extensionDetails, *ExtensionError) {
	// Parse all extension files and put them in a slice of extension details (the extension with the file name)
	extensions, err := parseExtensionsWithDetails(extensionFileNames)
	if err != nil {
		return nil, err
	}

	// Make sure each extension has its own ID
	err = checkExtensionIDsUniqueness(extensions, mtaID, mtaPath)
	if err != nil {
		return nil, err
	}

	// Make sure each extension extends a different ID and put them in a map of extends -> extension details
	extendsMap := make(map[string]extensionDetails, len(extensionFileNames))
	for _, details := range extensions {
		if value, ok := extendsMap[details.ext.Extends]; ok {
			return nil, &ExtensionError{details.fileName, errors.Errorf(duplicateExtendsMsg,
				value.fileName, details.fileName, details.ext.Extends), false}
		}
		extendsMap[details.ext.Extends] = details
	}

	// Verify chain of extensions and put the extensions in a slice by extends order
	return sortAndVerifyExtendsChain(extensionFileNames, mtaID, extendsMap)
}

func parseExtensionsWithDetails(extensionFileNames []string) ([]extensionDetails, *ExtensionError) {
	extensions := make([]extensionDetails, len(extensionFileNames))
	for i, extFileName := range extensionFileNames {
		extFile, err := parseExtFile(extFileName)
		if err != nil {
			return nil, &ExtensionError{extFileName, err, true}
		}
		extensions[i] = extensionDetails{extFileName, extFile}
	}
	return extensions, nil
}

func checkExtensionIDsUniqueness(extensions []extensionDetails, mtaID string, mtaPath string) *ExtensionError {
	extensionIDMap := make(map[string]extensionDetails, len(extensions))
	for _, details := range extensions {
		if details.ext.ID == mtaID {
			return &ExtensionError{details.fileName, errors.Errorf(extensionIDSameAsMtaIDMsg,
				details.fileName, mtaID, mtaPath), false}
		}
		if value, ok := extensionIDMap[details.ext.ID]; ok {
			return &ExtensionError{details.fileName, errors.Errorf(duplicateExtensionIDMsg,
				value.fileName, details.fileName, details.ext.ID), false}
		}
		extensionIDMap[details.ext.ID] = details
	}
	return nil
}

func sortAndVerifyExtendsChain(extensionFileNames []string, mtaID string, extendsMap map[string]extensionDetails) ([]extensionDetails, *ExtensionError) {
	sortedExtFiles := make([]extensionDetails, 0, len(extensionFileNames))
	currExtends := mtaID
	value, ok := extendsMap[currExtends]
	for ok {
		sortedExtFiles = append(sortedExtFiles, value)
		delete(extendsMap, currExtends)
		currExtends = value.ext.ID
		value, ok = extendsMap[currExtends]
	}
	// Check if there are extensions which extend unknown files
	if len(extendsMap) > 0 {
		// Build an error that looks like this:
		// `some MTA extension descriptors extend unknown IDs: file "myext.mtaext" extends "ext1"; file "aaa.mtaext" extends "ext2"`
		fileParts := make([]string, 0, len(extendsMap))
		fileName := ""
		for extends, details := range extendsMap {
			if fileName == "" {
				fileName = details.fileName
			}
			fileParts = append(fileParts, fmt.Sprintf(extendsMsg, details.fileName, extends))
		}
		// Return the error on the first encountered extension since we only support one error currently.
		// Note that it's not necessarily the first extension in the list of extension files (since the map iteration order is undefined).
		return nil, &ExtensionError{fileName, errors.Errorf(unknownExtendsMsg, strings.Join(fileParts, `; `)), false}
	}
	return sortedExtFiles, nil
}

func checkSchemaVersionMatches(mta *MTA, extDetails extensionDetails) error {
	mtaVersion := ""
	if mta.SchemaVersion != nil {
		mtaVersion = *mta.SchemaVersion
	}
	extVersion := ""
	if extDetails.ext.SchemaVersion != nil {
		extVersion = *extDetails.ext.SchemaVersion
	}

	// Check major version matches
	if strings.SplitN(mtaVersion, ".", 2)[0] != strings.SplitN(extVersion, ".", 2)[0] {
		return errors.Errorf(versionMismatchMsg, extVersion, extDetails.fileName, mtaVersion)
	}

	return nil
}

// Merge merges mta object with mta extension object extension properties complement and overwrite mta properties
func Merge(mta *MTA, mtaExt *EXT, extFilePath string) error {
	err := chain().
		extendMap(&mta.Parameters, mta.ParametersMetaData, mtaExt.Parameters, mergeRootParametersErrorMsg).
		err
	if err != nil {
		return wrapMergeError(err, extFilePath)
	}

	if err = mergeModules(*mta, mtaExt.Modules); err != nil {
		return wrapMergeError(err, extFilePath)
	}

	if err = mergeResources(*mta, mtaExt.Resources); err != nil {
		return wrapMergeError(err, extFilePath)
	}

	return nil
}

func wrapMergeError(err error, extFilePath string) error {
	return errors.Wrapf(err, mergeExtPathErrorMsg, extFilePath)
}

// mergeModules is responsible for handling the rules of merging modules
func mergeModules(mtaObj MTA, mtaExtModules []*ModuleExt) error {
	for _, extModule := range mtaExtModules {
		if module, _ := mtaObj.GetModuleByName(extModule.Name); module != nil {
			err := chain().
				extendMap(&module.Properties, module.PropertiesMetaData, extModule.Properties, mergeModulePropertiesErrorMsg, module.Name).
				extendMap(&module.Parameters, module.ParametersMetaData, extModule.Parameters, mergeModuleParametersErrorMsg, module.Name).
				extendMap(&module.BuildParams, nil, extModule.BuildParams, mergeModuleBuildParametersErrorMsg, module.Name).
				extendIncludes(&module.Includes, extModule.Includes, mergeModuleIncludesErrorMsg, module.Name).
				err
			if err != nil {
				return err
			}
			err = mergeModuleProvides(module, extModule)
			if err != nil {
				return err
			}
			if err = mergeRequires(extModule.Requires, module,
				msg{unknownModuleRequiresErrorMsg, []interface{}{extModule.Name}},
				msg{mergeModuleRequiresPropertiesErrorMsg, []interface{}{module.Name}},
				msg{mergeModuleRequiresParametersErrorMsg, []interface{}{module.Name}}); err != nil {
				return err
			}
			err = mergeModuleHooks(module, extModule)
			if err != nil {
				return err
			}
		} else {
			return errors.Errorf(unknownModuleErrorMsg, extModule.Name)
		}
	}

	return nil
}

func mergeModuleProvides(module *Module, extModule *ModuleExt) error {
	for _, extProvide := range extModule.Provides {
		if provide := module.GetProvidesByName(extProvide.Name); provide != nil {
			err := chain().
				extendMap(&provide.Properties, provide.PropertiesMetaData, extProvide.Properties, mergeModuleProvidesPropertiesErrorMsg, provide.Name, module.Name).
				err
			if err != nil {
				return err
			}
		} else {
			return errors.Errorf(unknownModuleProvidesErrorMsg, extProvide.Name, extModule.Name)
		}
	}
	return nil
}

func mergeModuleHooks(module *Module, extModule *ModuleExt) error {
	for _, extHook := range extModule.Hooks {
		if hook := module.GetHookByName(extHook.Name); hook != nil {
			err := chain().
				extendMap(&hook.Parameters, hook.ParametersMetaData, extHook.Parameters, mergeModuleHookParametersErrorMsg, hook.Name, module.Name).
				err
			if err != nil {
				return err
			}
			if err = mergeRequires(extHook.Requires, hook,
				msg{unknownModuleHookRequiresErrorMsg, []interface{}{extHook.Name, extModule.Name}},
				msg{mergeModuleHookRequiresPropertiesErrorMsg, []interface{}{hook.Name, module.Name}},
				msg{mergeModuleHookRequiresParametersErrorMsg, []interface{}{hook.Name, module.Name}}); err != nil {
				return err
			}
		} else {
			return errors.Errorf(unknownModuleHookErrorMsg, extHook.Name, extModule.Name)
		}
	}
	return nil
}

// mergeResources is responsible for handling the rules of merging resources
func mergeResources(mtaObj MTA, mtaExtResources []*ResourceExt) error {
	for _, extResource := range mtaExtResources {
		if resource := mtaObj.GetResourceByName(extResource.Name); resource != nil {
			err := chain().
				extendBoolPtr(&resource.Active, &extResource.Active, mergeResourceActiveErrorMsg, resource.Name).
				extendMap(&resource.Properties, resource.PropertiesMetaData, extResource.Properties, mergeResourcePropertiesErrorMsg, resource.Name).
				extendMap(&resource.Parameters, resource.ParametersMetaData, extResource.Parameters, mergeResourceParametersErrorMsg, resource.Name).
				err
			if err != nil {
				return err
			}
			if err = mergeRequires(extResource.Requires, resource,
				msg{unknownResourceRequiresErrorMsg, []interface{}{extResource.Name}},
				msg{mergeResourceRequiresPropertiesErrorMsg, []interface{}{resource.Name}},
				msg{mergeResourceRequiresParametersErrorMsg, []interface{}{resource.Name}}); err != nil {
				return err
			}
		} else {
			return errors.Errorf(unknownResourceErrorMsg, extResource.Name)
		}
	}

	return nil
}

type requiresProvider interface {
	GetRequiresByName(name string) *Requires
}
type msg struct {
	msg  string
	args []interface{}
}

func (msg msg) getArgs(prependArgs ...interface{}) []interface{} {
	return append(prependArgs, msg.args...)
}

// mergeRequires is responsible for merging the requires part of modules, resources etc
func mergeRequires(requires []Requires, extRequiresProvider requiresProvider, unknownRequiresMsg msg, mergePropertiesMsg msg, mergeParametersMsg msg) error {
	for _, extRequires := range requires {
		if requires := extRequiresProvider.GetRequiresByName(extRequires.Name); requires != nil {
			err := chain().
				extendMap(&requires.Properties, requires.PropertiesMetaData, extRequires.Properties, mergePropertiesMsg.msg, mergePropertiesMsg.getArgs(requires.Name)...).
				extendMap(&requires.Parameters, requires.ParametersMetaData, extRequires.Parameters, mergeParametersMsg.msg, mergeParametersMsg.getArgs(requires.Name)...).
				err
			if err != nil {
				return err
			}
		} else {
			return errors.Errorf(unknownRequiresMsg.msg, unknownRequiresMsg.getArgs(extRequires.Name)...)
		}
	}
	return nil
}

// isFieldOverWritable test is the current field allowed to be overwritten in mta file
func isFieldOverWritable(field string, meta map[string]MetaData, m map[string]interface{}) bool {
	if meta != nil && m[field] != nil {
		if metaData, exists := meta[field]; exists && metaData.OverWritable != nil {
			return *metaData.OverWritable
		}
	}
	return true
}

// extendMap extends map with elements of mta extension map
func extendMap(m *map[string]interface{}, meta map[string]MetaData, ext map[string]interface{}) error {
	if ext != nil {
		if *m == nil {
			*m = make(map[string]interface{})
		}
		for key, value := range ext {
			if isFieldOverWritable(key, meta, *m) {
				err := mergeMapKey(m, key, value)
				if err != nil {
					return err
				}
			} else {
				return errors.Errorf(overwriteNonOverwritableErrorMsg, key)
			}
		}
	}
	return nil
}

func mergeMapKey(m *map[string]interface{}, key string, value interface{}) error {
	extMapValue, extIsMap, _ := getMapValue(value)
	mMapValue, mIsMap, converted := getMapValue((*m)[key])

	if (*m)[key] == nil || value == nil || (!extIsMap && !mIsMap) {
		(*m)[key] = value
	} else if mIsMap && extIsMap {
		err := extendMap(&mMapValue, nil, extMapValue)
		if converted {
			(*m)[key] = mMapValue
		}
		if err != nil {
			return errors.Wrapf(err, `"%s"`, key)
		}
	} else {
		// Both values aren't nil. One of them is a map and the other is not.
		if mIsMap {
			return errors.Errorf(overwriteStructuredWithScalarErrorMsg, key)
		}
		return errors.Errorf(overwriteScalarWithStructuredErrorMsg, key)
	}
	return nil
}

// extendIncludes extends Slice Includes with elements of mta extension Slice Includes
func extendIncludes(m *[]Includes, ext []Includes) {
	if ext != nil {
		if *m == nil {
			*m = []Includes{}
		}
		*m = append(*m, ext...)
	}
}

// extendBoolPtr extends *bool with element of mta extension *bool
func extendBoolPtr(m **bool, ext **bool) {
	if ext != nil && *ext != nil {
		*m = *ext
	}
}

func chain() *chainError {
	return &chainError{}
}

// chainError is a struct for chaining error from the different function, allowing them to be called in one line.
type chainError struct {
	err error
}

func (v *chainError) extendMap(m *map[string]interface{}, meta map[string]MetaData, ext map[string]interface{}, msg string, args ...interface{}) *chainError {
	if v.err != nil {
		return v
	}
	err := extendMap(m, meta, ext)
	if err != nil {
		v.err = errors.Wrapf(err, msg, args...)
	}
	return v
}

func (v *chainError) extendIncludes(m *[]Includes, ext []Includes, msg string, args ...interface{}) *chainError {
	if v.err != nil {
		return v
	}
	extendIncludes(m, ext)
	return v
}

func (v *chainError) extendBoolPtr(m **bool, ext **bool, msg string, args ...interface{}) *chainError {
	if v.err != nil {
		return v
	}
	extendBoolPtr(m, ext)
	return v
}

func getMapValue(value interface{}) (ret map[string]interface{}, isMap bool, converted bool) {
	converted = false
	mapValue, isMap := value.(map[string]interface{})
	if !isMap {
		interfaceMap, isInterfaceMap := value.(map[interface{}]interface{})
		if isInterfaceMap {
			mapValue, isMap = convertMap(interfaceMap)
			converted = isMap
		}
	}
	return mapValue, isMap, converted
}

// convertMap converts type map[interface{}]interface{} to map[string]interface{}
func convertMap(m map[interface{}]interface{}) (map[string]interface{}, bool) {
	res := make(map[string]interface{})
	for key, value := range m {
		strKey, ok := key.(string)
		if !ok {
			return nil, false
		}
		res[strKey] = value
	}

	return res, true
}
