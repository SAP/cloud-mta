// Package mta provides a convenient way of exploring the structure of `mta.yaml` file objects
// such as retrieving a list of resources required by a specific module.
package mta

import (
	"bytes"
	"fmt"
	yamlv2 "gopkg.in/yaml.v2"
	"gopkg.in/yaml.v3"
)

// GetModules returns a list of MTA modules.
func (mta *MTA) GetModules() []*Module {
	return mta.Modules
}

// GetResources returns list of MTA resources.
func (mta *MTA) GetResources() []*Resource {
	return mta.Resources
}

// GetModuleByName returns a specific module by name.
func (mta *MTA) GetModuleByName(name string) (*Module, error) {
	for _, m := range mta.Modules {
		if m.Name == name {
			return m, nil
		}
	}
	return nil, fmt.Errorf(`the "%s" module is not defined`, name)
}

// GetResourceByName returns a specific resource by name.
func (mta *MTA) GetResourceByName(name string) (*Resource, error) {
	for _, r := range mta.Resources {
		if r.Name == name {
			return r, nil
		}
	}
	return nil, fmt.Errorf("the %s resource is not defined ", name)
}

// GetProvidesByName returns a specific provide by name
func (module *Module) GetProvidesByName(name string) *Provides {
	for _, p := range module.Provides {
		if p.Name == name {
			return &p
		}
	}

	return nil
}

// Unmarshal returns a reference to the MTA object from a byte array.
func Unmarshal(content []byte) (*MTA, error) {
	dec := yaml.NewDecoder(bytes.NewReader(content))
	dec.KnownFields(true)
	mtaStr := MTA{}
	err := dec.Decode(&mtaStr)
	//err := yaml.Unmarshal(content, &mtaStr)
	return &mtaStr, err
}

// Marshal marshals an MTA object
func Marshal(omta *MTA) ([]byte, error) {
	return yamlv2.Marshal(&omta)
}

// UnmarshalYAML unmarshals a MetaData object, setting default values for fields not in the source
func (meta *MetaData) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawMetadata MetaData
	raw := rawMetadata{OverWritable: true, Optional: false} // Default values
	if err := unmarshal(&raw); err != nil {
		return err
	}

	*meta = MetaData(raw)
	return nil
}
