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
func (mta *MTA) GetResourceByName(name string) *Resource {
	for _, r := range mta.Resources {
		if r.Name == name {
			return r
		}
	}
	return nil
}

// GetProvidesByName returns a specific provide by name
func (module *Module) GetProvidesByName(name string) *Provides {
	for i, p := range module.Provides {
		if p.Name == name {
			return &module.Provides[i]
		}
	}

	return nil
}

// GetRequiresByName returns a specific requires by name
func (module *Module) GetRequiresByName(name string) *Requires {
	for i, r := range module.Requires {
		if r.Name == name {
			return &module.Requires[i]
		}
	}

	return nil
}

// GetHookByName returns a specific hook by name
func (module *Module) GetHookByName(name string) *Hook {
	for i, h := range module.Hooks {
		if h.Name == name {
			return &module.Hooks[i]
		}
	}

	return nil
}

// GetRequiresByName returns a specific requires by name
func (resource *Resource) GetRequiresByName(name string) *Requires {
	for i, r := range resource.Requires {
		if r.Name == name {
			return &resource.Requires[i]
		}
	}

	return nil
}

// GetRequiresByName returns a specific requires by name
func (hook *Hook) GetRequiresByName(name string) *Requires {
	for i, r := range hook.Requires {
		if r.Name == name {
			return &hook.Requires[i]
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
func (meta *MetaData) UnmarshalYAML(node *yaml.Node) error {
	type metadata MetaData

	raw := metadata{} // Default values

	if err := node.Decode(&raw); err != nil {
		return err
	}

	*meta = MetaData(raw)
	return nil
}
