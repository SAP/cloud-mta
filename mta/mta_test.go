package mta

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mta", func() {

	modules := []*Module{
		{
			Name: "backend",
			Type: "java.tomcat",
			Path: "java",
			BuildParams: map[string]interface{}{
				"builder": "maven",
			},
			Properties: map[string]interface{}{
				"backend_type": nil,
			},
			PropertiesMetaData: map[string]interface{}{
				"backend_type": map[interface{}]interface{}{
					"optional":     false,
					"overwritable": true,
					"datatype":     "str",
				},
			},
			Parameters: map[string]interface{}{
				"domain":   nil,
				"password": "asfhuwehkew efgehk",
			},
			ParametersMetaData: map[string]interface{}{
				"domain": map[interface{}]interface{}{
					"optional":     false,
					"overwritable": true,
				},
			},
			Includes: []Includes{
				{
					Name: "config",
					Path: "cfg/parameters.json",
				},
			},
			Provides: []Provides{
				{
					Name:   "backend_task",
					Public: true,
					Properties: map[string]interface{}{
						"url": "${default-url}/tasks",
					},
					PropertiesMetaData: map[string]interface{}{
						"url": map[interface{}]interface{}{
							"optional":     true,
							"overwritable": true,
						},
					},
				},
			},
			Requires: []Requires{
				{
					Name: "database",
				},
				{
					Name: "scheduler_api",
					List: "mylist",
					Properties: map[string]interface{}{
						"scheduler_url": "~{url}",
					},
					PropertiesMetaData: map[string]interface{}{
						"scheduler_url": map[interface{}]interface{}{
							"optional": false,
						},
					},
					Includes: []Includes{
						{
							Name: "config",
							Path: "cfg/parameters.json",
						},
					},
				},
			},
			DeployedAfter: []interface{}{"scheduler"},
		},
		{
			Name: "scheduler",
			Type: "javascript.nodejs",
			Provides: []Provides{
				{
					Name: "scheduler_api",
					Properties: map[string]interface{}{
						"url": "${default-url}/api/v2",
					},
				},
			},
			Requires: []Requires{
				{
					Name: "backend_task",
					Properties: map[string]interface{}{
						"task_url": "~{url}",
					},
				},
			},
		},
	}

	buildersBefore := Builders{
		{
			Builder: "mybuilder",
		},
	}
	buildersAfter := Builders{
		{
			Builder: "otherbuilder",
		},
	}

	buildParams := &ProjectBuild{
		BeforeAll: struct {
			Builders Builders `yaml:"builders,omitempty"`
		}{
			Builders: buildersBefore,
		},
		AfterAll: struct {
			Builders Builders `yaml:"builders,omitempty"`
		}{
			Builders: buildersAfter,
		},
	}

	schemaVersion := "3.2"
	mta := &MTA{
		SchemaVersion: &schemaVersion,
		ID:            "com.acme.scheduling",
		Version:       "1.132.1-edfsd+ewfe",
		Parameters:    map[string]interface{}{"deployer-version": ">=1.2.0"},
		BuildParams:   buildParams,
		Modules:       modules,
		Resources: []*Resource{
			{
				Name: "database",
				Type: "postgresql",
			},
			{
				Name:     "plugins",
				Type:     "configuration",
				Optional: true,
				Active:   false,
				Requires: []Requires{
					{
						Name: "scheduler_api",
						Parameters: map[string]interface{}{
							"par1": "value",
						},
						Properties: map[string]interface{}{
							"prop1": "${value}-~{url}",
						},
					},
				},
				Includes: []Includes{
					{
						Name: "config",
						Path: "cfg/security.json",
					},
					{
						Name: "creation",
						Path: "djdk.yaml",
					},
				},
				Parameters: map[string]interface{}{
					"filter": map[string]interface{}{
						"type": "com.acme.plugin",
					},
				},
				ParametersMetaData: map[string]interface{}{
					"filter": map[interface{}]interface{}{
						"optional":     false,
						"overwritable": false,
					},
				},
				Properties: map[string]interface{}{
					"plugin_name": "${name}",
					"plugin_url":  "${url}/sources",
				},
				PropertiesMetaData: map[string]interface{}{
					"plugin_name": map[interface{}]interface{}{
						"optional": true,
					},
				},
			},
		},
		ModuleTypes: []*ModuleTypes{
			{
				Name:    "java.tomcat",
				Extends: "java",
				Parameters: map[string]interface{}{
					"buildpack": nil,
					"memory":    "256M",
				},
				ParametersMetaData: map[string]interface{}{
					"buildpack": map[interface{}]interface{}{
						"optional": false,
					},
				},
				Properties: map[string]interface{}{
					"TARGET_RUNTIME": "tomcat",
				},
			},
		},
		ResourceTypes: []*ResourceTypes{
			{
				Name:    "postgresql",
				Extends: "managed-service",
				Parameters: map[string]interface{}{
					"service":      "postgresql",
					"service-plan": nil,
				},
				ParametersMetaData: map[string]interface{}{
					"service-plan": map[interface{}]interface{}{
						"optional": false,
					},
				},
			},
		},
	}
	var _ = Describe("MTA tests", func() {

		var _ = Describe("Parsing", func() {
			It("Modules parsing - sanity", func() {
				mtaFile, _ := ioutil.ReadFile("./testdata/mta.yaml")
				// Unmarshal file
				oMta := &MTA{}
				Ω(yaml.Unmarshal(mtaFile, oMta)).Should(Succeed())
				Ω(oMta.Modules).Should(HaveLen(2))
				Ω(oMta.GetModules()).Should(Equal(modules))

			})

		})

		var _ = Describe("Get methods on MTA", func() {
			It("GetModules", func() {
				Ω(mta.GetModules()).Should(Equal(modules))
			})
			It("GetResourceByName - Sanity", func() {
				Ω(mta.GetResourceByName("database")).Should(Equal(mta.Resources[0]))
				Ω(mta.GetResourceByName("plugins")).Should(Equal(mta.Resources[1]))
			})
			It("GetResourceByName - Negative", func() {
				_, err := mta.GetResourceByName("")
				Ω(err).Should(HaveOccurred())
			})
			It("GetResources - Sanity ", func() {
				Ω(mta.GetResources()).Should(Equal(mta.Resources))
			})
			It("GetModuleByName - Sanity ", func() {
				Ω(mta.GetModuleByName("backend")).Should(Equal(modules[0]))
				Ω(mta.GetModuleByName("scheduler")).Should(Equal(modules[1]))
			})
			It("GetModuleByName - Negative ", func() {
				_, err := mta.GetModuleByName("foo")
				Ω(err).Should(HaveOccurred())
			})
		})
	})

	var _ = Describe("Unmarshal", func() {
		It("Sanity", func() {
			wd, err := os.Getwd()
			Ω(err).Should(Succeed())
			content, err := ioutil.ReadFile(filepath.Join(wd, "testdata", "mta.yaml"))
			Ω(err).Should(Succeed())
			m, err := Unmarshal(content)
			Ω(err).Should(Succeed())

			Ω(mta).Should(BeEquivalentTo(m))
		})


	})

})
