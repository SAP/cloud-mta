package mta

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"strings"
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
			PropertiesMetaData: map[string]MetaData{
				"backend_type": {
					Optional:     false,
					OverWritable: true,
					Datatype:     "str",
				},
			},
			Parameters: map[string]interface{}{
				"domain":   nil,
				"password": "asfhuwehkew efgehk",
			},
			ParametersMetaData: map[string]MetaData{
				"domain": {
					Optional:     false,
					OverWritable: true,
				},
			},
			Includes: []Includes{
				{
					Name: "configs",
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
					PropertiesMetaData: map[string]MetaData{
						"url": {
							Optional:     true,
							OverWritable: true,
						},
					},
				},
				{
					Name:   "finished_backend_tasks",
					Public: true,
					Properties: map[string]interface{}{
						"url": "${default-url}/finishedTasks",
					},
					PropertiesMetaData: map[string]MetaData{
						"url": {
							Optional:     true,
							OverWritable: false,
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
					PropertiesMetaData: map[string]MetaData{
						"scheduler_url": {
							Optional:     false,
							OverWritable: true,
						},
					},
					Includes: []Includes{
						{
							Name: "configs",
							Path: "cfg/parameters.json",
						},
					},
				},
			},
			DeployedAfter: []string{"scheduler"},
			Hooks: []Hook{
				{
					Name:   "hook",
					Type:   "task",
					Phases: []string{"application.before-stop.live", "application.before-stop.idle"},
					Parameters: map[string]interface{}{
						"name":    "foo-task",
						"command": "sleep 5m",
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

	buildersBefore := []ProjectBuilder{
		{
			Builder: "mybuilder",
		},
	}
	buildersAfter := []ProjectBuilder{
		{
			Builder: "otherbuilder",
		},
	}

	buildParams := &ProjectBuild{
		BeforeAll: buildersBefore,
		AfterAll:  buildersAfter,
	}

	schemaVersion := "3.3"
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
						Name: "configs",
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
				ParametersMetaData: map[string]MetaData{
					"filter": {
						Optional:     false,
						OverWritable: false,
					},
				},
				Properties: map[string]interface{}{
					"plugin_name": "${name}",
					"plugin_url":  "${url}/sources",
				},
				PropertiesMetaData: map[string]MetaData{
					"plugin_name": {
						Optional:     true,
						OverWritable: true,
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
				ParametersMetaData: map[string]MetaData{
					"buildpack": {
						Optional:     false,
						OverWritable: true,
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
				ParametersMetaData: map[string]MetaData{
					"service-plan": {
						Optional:     false,
						OverWritable: true,
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

		Describe("Get methods on Module", func() {
			It("GetProvidesByName - Sanity", func() {
				module := mta.Modules[0]
				Ω(*(module.GetProvidesByName("backend_task"))).Should(Equal(module.Provides[0]))
				Ω(*(module.GetProvidesByName("finished_backend_tasks"))).Should(Equal(module.Provides[1]))
				Ω(module.GetProvidesByName("finished")).Should(BeNil())
			})
		})
	})

	var _ = Describe("Unmarshal", func() {
		It("Sanity", func() {
			wd, err := os.Getwd()
			Ω(err).Should(Succeed())
			content, err := readFile(filepath.Join(wd, "testdata", "mta.yaml"))
			Ω(err).Should(Succeed())
			m, err := Unmarshal(content)
			Ω(err).Should(Succeed())

			Ω(mta).Should(BeEquivalentTo(m))
		})

		It("Wrong deployed-after value", func() {
			content, err := readFile(getTestPath("mtaWrongDeployedAfter.yaml"))
			Ω(err).Should(Succeed())
			_, err = Unmarshal(content)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(ContainSubstring("line 54: cannot unmarshal !!int `1` into []string"))
		})

		It("Wrong properties-metadata value", func() {
			content, err := readFile(getTestPath("mtaWrongMetaData.yaml"))
			Ω(err).Should(Succeed())
			_, err = Unmarshal(content)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(ContainSubstring("line 23: cannot unmarshal !!bool `true` into mta.metadata"))
		})
	})
})

func readFile(file string) ([]byte, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, `failed to read the "%v" file`, file)
	}
	s := string(content)
	s = strings.Replace(s, "\r\n", "\r", -1)
	content = []byte(s)
	return content, nil
}
