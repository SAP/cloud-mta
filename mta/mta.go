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
func (meta *MetaData) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// TODO this should ignore unknown fields according to the spec.
	// However, trying to do so doesn't work properly using the Go yaml library (v2 or v3).
	//
	// The unmarshal function sent here returns errors for unknown fields and doesn't expose a way to set the KnownFields parameter in the internal decoder.
	//
	// Trying to "unmarshal" to a yaml.Node object (so we can then decode it without getting errors about unknown fields)
	// doesn't work because unmarshal can only work on pointers (otherwise it panics), but sending a pointer to a yaml.Node
	// tries to unmarshal the value as a yaml.Node object (with the fields of the yaml.Node), even though
	// it seems there is logic there to handle this type (this looks like a bug).
	//
	// Another option is to use the new Unmarshaler interface that receives a node:
	// func (meta *MetaData) UnmarshalYAML(node *yaml.Node) error) error {...}
	// A yaml.Node can be decoded without getting errors about unknown fields, and this works when there are no other
	// errors in the unmarshal. But if UnmarshalYaml returns an error (due to an error returned from node.Decode), several issues occur:
	// 1. If we return an error from UnmarshalYaml, the error handling catches it too late and the MTA object is returned
	//    wrong - e.g. if a parameters-metadata entry inside a module cannot be unmarshaled, the whole module is returned nil.
	// 2. After the error is returned, no further unmarhsaling is done on the Modules sequence. They're all returned as nil and only the errors
	//    from the first UnmarshalYAML call are returned.
	// I opened a bug on these 2 issues here: https://github.com/go-yaml/yaml/issues/499
	//
	// I also tried to use the unmarshal function (in the current method) to unmarshal the metadata to map[interface{}]interface{}
	// and then marshal that to a []byte, and use a decoder with KnownFields=false to decode the []byte to a MetaData structure.
	// This works but the line numbers returned for errors are wrong because we lose the original yaml.Node in the process.
	//
	// To check if this works un-ignore the test in mta_validate_test.go that mentions this method ("doesn't give errors on unknown fields in properties-metadata and parameters-metadata").

	type rawMetadata MetaData
	raw := rawMetadata{OverWritable: true, Optional: false} // Default values

	if err := unmarshal(&raw); err != nil {
		return err
	}

	*meta = MetaData(raw)
	return nil
}
