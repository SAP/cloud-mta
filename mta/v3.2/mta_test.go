package v3_2

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var _ = Describe("Mta", func() {
	var _ = Describe("Parsing", func() {
		It("Modules parsing - sanity", func() {
			schemaVersion := "3.2"
			var module = MTA_3_2{
				SchemaVersion: &schemaVersion,
				ID:            "com.acme.scheduling",
				Version:       "1.132.1-edfsd+ewfe",
				Parameters:map[string]interface{}{"deployer-version":">=1.2.0",},
				Modules: []*Module{
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
						Parameters: map[string]interface{}{
							"domain": nil,
							"password": "asfhuwehkew efgehk",
						},
						Includes: []Includes{
							{
								Name: "config",
								Path: "cfg/parameters.json",
							},
						},
						Provides: []Provides{
							{
								Name: "backend_task",
								Properties: map[string]interface{}{
									"url": "${default-url}/tasks",
								},
							},
						},
						Requires: []Requires{
							{
								Name: "database",
							},
							{
								Name: "scheduler_api",
								Properties: map[string]interface{}{
									"scheduler_url": "~{url}",
								},
								Includes:[]Includes {
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
				},
				Resources: []*Resource{
					{
						Name: "database",
						Type: "postgresql",
					},
					{
						Name:     "plugins",
						Type:     "configuration",
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
						Properties: map[string]interface{}{
							"plugin_name": "${name}",
							"plugin_url":  "${url}/sources",
						},
					},
				},
			}
			mtaFile, _ := ioutil.ReadFile("test.yaml")
			// Parse file
			oMta := &MTA_3_2{}
			Ω(yaml.Unmarshal(mtaFile, oMta)).Should(Succeed())
			Ω(module).Should(BeEquivalentTo(*oMta))
			Ω(oMta.Modules).Should(HaveLen(2))
		})
	})
})