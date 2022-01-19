package mta

// MTA mta schema, the schema will contain the latest mta schema version
// and all the previous version will be as subset of the latest
// Todo - Add the missing properties to support the latest 3.2 version
type MTA struct {
	// indicates MTA schema version, using semver.
	SchemaVersion *string `yaml:"_schema-version" json:"_schema-version"`
	// A globally unique ID of this MTA. Unlimited string of unicode characters.
	ID string `yaml:"ID" json:"ID"`
	// A non-translatable description of this MTA. This is not a text for application users
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Application version, using semantic versioning standard
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
	// The provider or vendor of this software
	Provider string `yaml:"provider,omitempty" json:"provider,omitempty"`
	// A copyright statement from the provider
	Copyright string `yaml:"copyright,omitempty" json:"copyright,omitempty"`
	// list of modules
	Modules []*Module `yaml:"modules,omitempty" json:"modules,omitempty"`
	// Module type declarations
	ModuleTypes []*ModuleTypes `yaml:"module-types,omitempty" json:"module-types,omitempty"`
	// Resource declarations. Resources can be anything required to run the application which is not provided by the application itself
	Resources []*Resource `yaml:"resources,omitempty" json:"resources,omitempty"`
	// Resource type declarations
	ResourceTypes []*ResourceTypes `yaml:"resource-types,omitempty" json:"resource-types,omitempty"`
	// Parameters can be used to steer the behavior of tools which interpret this descriptor
	Parameters         map[string]interface{} `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	ParametersMetaData map[string]MetaData    `yaml:"parameters-metadata,omitempty" json:"parameters-metadata,omitempty"`
	// Experimental - use for pre/post hook
	BuildParams *ProjectBuild `yaml:"build-parameters,omitempty" json:"build-parameters,omitempty"`
}

// Module - modules section.
type Module struct {
	// An MTA internal module name. Names need to be unique within the MTA scope
	Name string `yaml:"name" json:"name"`
	// a globally unique type ID. Deployment tools will interpret this type ID
	Type string `yaml:"type" json:"type"`
	// a non-translatable description of this module. This is not a text for application users
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// A file path which identifies the location of module artifacts.
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
	// Provided property values can be accessed by "~{<name-of-provides-section>/<provided-property-name>}". Such expressions can be part of an arbitrary string
	Properties         map[string]interface{} `yaml:"properties,omitempty" json:"properties,omitempty"`
	PropertiesMetaData map[string]MetaData    `yaml:"properties-metadata,omitempty" json:"properties-metadata,omitempty"`
	// THE 'includes' ELEMENT IS ONLY RELEVANT FOR DEVELOPMENT DESCRIPTORS (PRIO TO BUILD), NOT FOR DEPLOYMENT DESCRIPTORS!
	Includes []Includes `yaml:"includes,omitempty" json:"includes,omitempty"`
	// list of names either matching a resource name or a name provided by another module within the same MTA
	Requires []Requires `yaml:"requires,omitempty" json:"requires,omitempty"`
	// List of provided names (MTA internal)to which properties (= configuration data) can be attached
	Provides []Provides `yaml:"provides,omitempty" json:"provides,omitempty"`
	// Parameters can be used to steer the behavior of tools which interpret this descriptor. Parameters are not made available to the module at runtime
	Parameters         map[string]interface{} `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	ParametersMetaData map[string]MetaData    `yaml:"parameters-metadata,omitempty" json:"parameters-metadata,omitempty"`
	// Build-parameters are specifically steering the behavior of build tools.
	BuildParams map[string]interface{} `yaml:"build-parameters,omitempty" json:"build-parameters,omitempty"`
	// A list containing the names of the modules that must be deployed prior to this one.
	DeployedAfter []string `yaml:"deployed-after,omitempty" json:"deployed-after,omitempty"`
	// Defined and executed at specific phases of module deployment.
	Hooks []Hook `yaml:"hooks,omitempty" json:"hooks,omitempty"`
}

// ModuleTypes module types declarations
type ModuleTypes struct {
	// An MTA internal name of the module type. Can be specified in the 'type' element of modules
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// The name of the extended type. Can be another resource type defined in this descriptor or one of the default types supported by the deployer
	Extends string `yaml:"extends,omitempty" json:"extends,omitempty"`
	// Properties inherited by all resources of this type
	Properties         map[string]interface{} `yaml:"properties,omitempty" json:"properties,omitempty"`
	PropertiesMetaData map[string]MetaData    `yaml:"properties-metadata,omitempty" json:"properties-metadata,omitempty"`
	// Parameters inherited by all resources of this type
	Parameters         map[string]interface{} `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	ParametersMetaData map[string]MetaData    `yaml:"parameters-metadata,omitempty" json:"parameters-metadata,omitempty"`
}

// Provides List of provided names to which properties (configs data) can be attached.
type Provides struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Indicates, that the provided properties shall be made publicly available by the deployer
	Public bool `yaml:"public,omitempty" json:"public,omitempty"`
	// property names and values make up the configuration data which is to be provided to requiring modules at runtime
	Properties         map[string]interface{} `yaml:"properties,omitempty" json:"properties,omitempty"`
	PropertiesMetaData map[string]MetaData    `yaml:"properties-metadata,omitempty" json:"properties-metadata,omitempty"`
}

// Requires list of names either matching a resource name or a name provided by another module within the same MTA.
type Requires struct {
	// an MTA internal name which must match either a provided name, a resource name, or a module name within the same MTA
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// A group name which shall be use by a deployer to group properties for lookup by a module runtime.
	Group string `yaml:"group,omitempty" json:"group,omitempty"`
	// All required and found configuration data sets will be assembled into a JSON array and provided to the module by the lookup name as specified by the value of 'list'
	List string `yaml:"list,omitempty" json:"list,omitempty"`
	// Provided property values can be accessed by "~{<provided-property-name>}". Such expressions can be part of an arbitrary string
	Properties         map[string]interface{} `yaml:"properties,omitempty" json:"properties,omitempty"`
	PropertiesMetaData map[string]MetaData    `yaml:"properties-metadata,omitempty" json:"properties-metadata,omitempty"`
	// Parameters can be used to influence the behavior of tools which interpret this descriptor. Parameters are not made available to requiring modules at runtime
	Parameters         map[string]interface{} `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	ParametersMetaData map[string]MetaData    `yaml:"parameters-metadata,omitempty" json:"parameters-metadata,omitempty"`
	// THE 'includes' ELEMENT IS ONLY RELEVANT FOR DEVELOPMENT DESCRIPTORS (PRIO TO BUILD), NOT FOR DEPLOYMENT DESCRIPTORS!
	Includes []Includes `yaml:"includes,omitempty" json:"includes,omitempty"`
}

// Resource can be anything required to run the application which is not provided by the application itself.
type Resource struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// A type of a resource. This type is interpreted by and must be known to the deployer. Resources can be untyped
	Type string `yaml:"type,omitempty" json:"type,omitempty"`
	// A non-translatable description of this resource. This is not a text for application users
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Parameters can be used to influence the behavior of tools which interpret this descriptor. Parameters are not made available to requiring modules at runtime
	Parameters         map[string]interface{} `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	ParametersMetaData map[string]MetaData    `yaml:"parameters-metadata,omitempty" json:"parameters-metadata,omitempty"`
	// property names and values make up the configuration data which is to be provided to requiring modules at runtime
	Properties         map[string]interface{} `yaml:"properties,omitempty" json:"properties,omitempty"`
	PropertiesMetaData map[string]MetaData    `yaml:"properties-metadata,omitempty" json:"properties-metadata,omitempty"`
	// THE 'includes' ELEMENT IS ONLY RELEVANT FOR DEVELOPMENT DESCRIPTORS (PRIO TO BUILD), NOT FOR DEPLOYMENT DESCRIPTORS!
	Includes []Includes `yaml:"includes,omitempty" json:"includes,omitempty"`
	// A resource can be declared to be optional, if the MTA can compensate for its non-existence
	Optional bool `yaml:"optional,omitempty" json:"optional,omitempty"`
	// If a resource is declared to be active, it is allocated and bound according to declared requirements
	Active *bool `yaml:"active,omitempty" json:"active,omitempty"`
	// A list containing the names of the resources that must be processed prior to this one.
	ProcessedAfter []string `yaml:"processed-after,omitempty" json:"processed-after,omitempty"`
	// list of names either matching a resource name or a name provided by another module within the same MTA
	Requires []Requires `yaml:"requires,omitempty" json:"requires,omitempty"`
}

// ResourceTypes resources type declarations
type ResourceTypes struct {
	// An MTA internal name of the module type. Can be specified in the 'type' element of modules
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// The name of the extended type. Can be another resource type defined in this descriptor or one of the default types supported by the deployer
	Extends string `yaml:"extends,omitempty" json:"extends,omitempty"`
	// Properties inherited by all resources of this type
	Properties         map[string]interface{} `yaml:"properties,omitempty" json:"properties,omitempty"`
	PropertiesMetaData map[string]MetaData    `yaml:"properties-metadata,omitempty" json:"properties-metadata,omitempty"`
	// Parameters inherited by all resources of this type
	Parameters         map[string]interface{} `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	ParametersMetaData map[string]MetaData    `yaml:"parameters-metadata,omitempty" json:"parameters-metadata,omitempty"`
}

// Includes The 'includes' element only relevant for development descriptor, not for deployment descriptor
type Includes struct {
	// A name of an include section. This name will be used by a builder to generate a parameter section in the deployment descriptor
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// A path pointing to a file which contains a map of parameters, either in JSON or in YAML format.
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
}

// ProjectBuild - experimental use for pre/post build hook
type ProjectBuild struct {
	BeforeAll []ProjectBuilder `yaml:"before-all,omitempty" json:"before-all,omitempty"`
	AfterAll  []ProjectBuilder `yaml:"after-all,omitempty" json:"after-all,omitempty"`
}

// ProjectBuilder - project builder descriptor
type ProjectBuilder struct {
	Builder  string   `yaml:"builder,omitempty" json:"builder,omitempty"`
	Timeout  string   `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Commands []string `yaml:"commands,omitempty" json:"commands,omitempty"`
}

// Hook - defined and executed at specific phases of module deployment.
type Hook struct {
	// An MTA internal name which can be used for documentation purposes and shown by the deployer.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Defines the type of action that should be executed by the deployer.
	Type string `yaml:"type,omitempty" json:"type,omitempty"`
	// A list of strings that define the points at which the hook must be executed.
	Phases             []string               `yaml:"phases,omitempty" json:"phases,omitempty"`
	Parameters         map[string]interface{} `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	ParametersMetaData map[string]MetaData    `yaml:"parameters-metadata,omitempty" json:"parameters-metadata,omitempty"`
	Requires           []Requires             `yaml:"requires,omitempty" json:"requires,omitempty"`
}

// MetaData - The properties-metadata and the parameters-metadata structure
type MetaData struct {
	// If set to true, the value can be overwritten by an extension descriptor.
	OverWritable *bool `yaml:"overwritable,omitempty" json:"overwritable,omitempty"`
	// If set to false, a value must be present in the final deployment configuration.
	Optional *bool `yaml:"optional,omitempty" json:"optional,omitempty"`
	// An interface with which a UI-tool can query for possible parameter names together with the expected datatypes and default values.
	Datatype interface{} `yaml:"datatype,omitempty" json:"datatype,omitempty"`
	// Indicate sensitive information to a UI-tool which it can use, e.g., for masking a value
	Sensitive bool `yaml:"sensitive,omitempty" json:"sensitive,omitempty"`
}
