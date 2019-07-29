package mta

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	unmarshalExtErrorMsg   = "could not unmarshal the MTA extension file"
	mergeModulesErrorMsg   = `could not merge modules from MTA extension with ID "%s"`
	mergeResourcesErrorMsg = `could not merge resources from MTA extension with ID "%s"`

	mergeModuleErrorMsg           = `could not merge the "%s" module`
	mergeModuleProvidesErrorMsg   = `could not merge the "%s" provides in the "%s" module`
	unknownModuleProvidesErrorMsg = `the "%s" provides in the "%s" module is defined in the MTA extension but not in the "mta.yaml" file`
	unknownModuleErrorMsg         = `the "%s" module is defined in the MTA extension but not in the "mta.yaml" file`

	mergeResourceErrorMsg   = `could not merge the "%s" resource`
	unknownResourceErrorMsg = `the "%s" resource is defined in the MTA extension but not in the "mta.yaml" file`

	overwriteStructuredWithScalarErrorMsg = `"%s": cannot overwrite a structured value with a scalar value`
	overwriteScalarWithStructuredErrorMsg = `"%s": cannot overwrite a scalar value with a structured value`
	overwriteNonOverwritableErrorMsg      = `the "%s" field is not overwritable`
)

// UnmarshalExt returns a reference to the EXT object from a byte array.
func UnmarshalExt(content []byte) (*EXT, error) {
	var m EXT
	// Unmarshal MTA file
	err := yaml.Unmarshal(content, &m)
	if err != nil {
		err = errors.Wrap(err, unmarshalExtErrorMsg)
	}
	return &m, err
}

// Merge merges mta object with mta extension object extension properties complement and overwrite mta properties
func Merge(mta *MTA, mtaExt *EXT) error {
	if err := mergeModules(*mta, mtaExt.Modules); err != nil {
		return errors.Wrapf(err, mergeModulesErrorMsg, mtaExt.ID)
	}

	if err := mergeResources(*mta, mtaExt.Resources); err != nil {
		return errors.Wrapf(err, mergeResourcesErrorMsg, mtaExt.ID)
	}

	return nil
}

// mergeModules is responsible for handling the rules of merging modules
func mergeModules(mtaObj MTA, mtaExtModules []*ModuleExt) error {
	for _, extModule := range mtaExtModules {
		if module, err := mtaObj.GetModuleByName(extModule.Name); err == nil {
			err = chain().
				extendMap(&module.Properties, module.PropertiesMetaData, extModule.Properties).
				extendMap(&module.Parameters, module.ParametersMetaData, extModule.Parameters).
				extendMap(&module.BuildParams, nil, extModule.BuildParams).
				extendIncludes(&module.Includes, extModule.Includes).
				err
			if err != nil {
				return errors.Wrapf(err, mergeModuleErrorMsg, module.Name)
			}
			for _, extProvide := range extModule.Provides {
				if provide := module.GetProvidesByName(extProvide.Name); provide != nil {
					if err = extendMap(&provide.Properties, provide.PropertiesMetaData, extProvide.Properties); extModule != nil {
						return errors.Wrapf(err, mergeModuleProvidesErrorMsg, provide.Name, module.Name)
					}
				} else {
					return errors.Errorf(unknownModuleProvidesErrorMsg, extProvide.Name, extModule.Name)
				}
			}
		} else {
			return errors.Wrapf(err, unknownModuleErrorMsg, extModule.Name)
		}
	}

	return nil
}

// mergeResources is responsible for handling the rules of merging modules
func mergeResources(mtaObj MTA, mtaExtResources []*ResourceExt) error {
	for _, extResource := range mtaExtResources {
		if resource, err := mtaObj.GetResourceByName(extResource.Name); err == nil {
			err = chain().
				extendBool(&resource.Active, &extResource.Active).
				err
			if err != nil {
				return errors.Wrapf(err, mergeResourceErrorMsg, resource.Name)
			}
		} else {
			return errors.Wrapf(err, unknownResourceErrorMsg, extResource.Name)
		}
	}

	return nil
}

// isFieldOverWritable test is the current field allowed to be overwritten in mta file
func isFieldOverWritable(field string, meta map[string]MetaData, m map[string]interface{}) bool {
	if meta != nil && m[field] != nil {
		if metaData, exists := meta[field]; exists {
			return metaData.OverWritable
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
	extMapValue, extIsMap := value.(map[string]interface{})
	mMapValue, mIsMap := (*m)[key].(map[string]interface{})
	if (*m)[key] == nil || value == nil || (!extIsMap && !mIsMap) {
		(*m)[key] = value
	} else if mIsMap && extIsMap {
		if err := extendMap(&mMapValue, nil, extMapValue); err != nil {
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
func extendIncludes(m *[]Includes, ext []Includes) error {
	if ext != nil {
		if *m == nil {
			*m = []Includes{}
		}
		*m = append(*m, ext...)
	}
	return nil
}

// extendBool extends bool with element of mta extension bool
func extendBool(m *bool, ext *bool) error {
	if ext != nil {
		*m = *ext
	}

	return nil
}

func chain() *chainError {
	return &chainError{}
}

// chainError is a struct for chaining error from the different function, allowing them to be called in one line.
type chainError struct {
	err error
}

func (v *chainError) extendMap(m *map[string]interface{}, meta map[string]MetaData, ext map[string]interface{}) *chainError {
	if v.err != nil {
		return v
	}
	v.err = extendMap(m, meta, ext)
	return v
}

func (v *chainError) extendIncludes(m *[]Includes, ext []Includes) *chainError {
	if v.err != nil {
		return v
	}
	v.err = extendIncludes(m, ext)
	return v
}

func (v *chainError) extendBool(m *bool, ext *bool) *chainError {
	if v.err != nil {
		return v
	}
	v.err = extendBool(m, ext)
	return v
}
