package mta

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
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
	schemaVersion := "3.2"
	mta := &MTA{
		SchemaVersion: &schemaVersion,
		ID:            "com.acme.scheduling",
		Version:       "1.132.1-edfsd+ewfe",
		Parameters:    map[string]interface{}{"deployer-version": ">=1.2.0",},
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
					"filter": map[interface{}]interface{}{
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
			Ω(*mta).Should(BeEquivalentTo(*m))
			Ω(len(m.Modules)).Should(Equal(2))
		})
		It("Invalid content", func() {
			_, err := Unmarshal([]byte("wrong mta"))
			Ω(err).Should(HaveOccurred())
		})
	})

	var _ = Describe("UnmarshalExt", func() {
		It("Sanity", func() {
			wd, err := os.Getwd()
			Ω(err).Should(Succeed())
			content, err := ioutil.ReadFile(filepath.Join(wd, "testdata", "mta.yaml"))
			Ω(err).Should(Succeed())
			m, err := UnmarshalExt(content)
			Ω(err).Should(Succeed())
			Ω(len(m.Modules)).Should(Equal(2))
		})
		It("Invalid content", func() {
			_, err := UnmarshalExt([]byte("wrong mtaExt"))
			Ω(err).Should(HaveOccurred())
		})
	})

	var _ = Describe("extendMap", func() {
		var m1 map[string]interface{}
		var m2 map[string]interface{}
		var m3 map[string]interface{}
		var m4 map[string]interface{}

		BeforeEach(func() {
			m1 = make(map[string]interface{})
			m2 = make(map[string]interface{})
			m3 = make(map[string]interface{})
			m4 = nil
			m1["a"] = "aa"
			m1["b"] = "xx"
			m2["b"] = "bb"
			m3["c"] = "cc"
		})

		var _ = DescribeTable("Sanity", func(m *map[string]interface{}, e *map[string]interface{}, ln int, key string, value interface{}) {
			extendMap(m, e)
			Ω(len(*m)).Should(Equal(ln))

			if value != nil {
				Ω((*m)[key]).Should(Equal(value))
			} else {
				Ω((*m)[key]).Should(BeNil())
			}
		},
			Entry("overwrite", &m1, &m2, 2, "b", "bb"),
			Entry("add", &m1, &m3, 3, "c", "cc"),
			Entry("res equals ext", &m4, &m1, 2, "b", "xx"),
			Entry("nothing to add", &m1, &m4, 2, "b", "xx"),
			Entry("both nil", &m4, &m4, 0, "b", nil),
		)
	})

	var _ = Describe("MergeMtaAndExt", func() {
		It("Sanity", func() {
			moduleA := Module{
				Name: "modA",
				Properties: map[string]interface{}{
					"a": "aa",
					"b": "xx",
				},
			}
			moduleB := Module{
				Name: "modB",
				Properties: map[string]interface{}{
					"b": "yy",
				},
			}
			moduleAExt := ModuleExt{
				Name: "modA",
				Properties: map[string]interface{}{
					"a": "aa",
					"b": "bb",
				},
			}
			mta := MTA{
				Modules: []*Module{&moduleA, &moduleB},
			}
			mtaExt := EXT{
				Modules: []*ModuleExt{&moduleAExt},
			}
			Merge(&mta, &mtaExt)
			m, err := mta.GetModuleByName("modA")
			Ω(err).Should(Succeed())
			Ω(m.Properties["b"]).Should(Equal("bb"))
		})
	})

})
