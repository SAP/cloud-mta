package validate

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	"github.com/SAP/cloud-mta/internal/fs"
	"github.com/SAP/cloud-mta/mta"
)

func getTestPath(relPath ...string) string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "testdata", filepath.Join(relPath...))
}

var _ = Describe("MTA tests", func() {

	var _ = Describe("Parsing", func() {
		It("Modules parsing - sanity", func() {
			var moduleSrv = mta.Module{
				Name: "srv",
				Type: "java",
				Path: "srv",
				Requires: []mta.Requires{
					{
						Name: "db",
						Properties: map[string]interface{}{
							"JBP_CONFIG_RESOURCE_CONFIGURATION": `[tomcat/webapps/ROOT/META-INF/context.xml: {"service_name_for_DefaultDB" : "~{hdi-container-name}"}]`,
						},
					},
				},
				Provides: []mta.Provides{
					{
						Name:       "srv_api",
						Properties: map[string]interface{}{"url": "${default-url}"},
					},
				},
				Parameters: map[string]interface{}{"memory": "512M"},
				Properties: map[string]interface{}{
					"VSCODE_JAVA_DEBUG_LOG_LEVEL": "ALL",
					"APPC_LOG_LEVEL":              "info",
				},
			}
			var moduleUI = mta.Module{
				Name: "ui",
				Type: "html5",
				Path: "ui",
				Requires: []mta.Requires{
					{
						Name:  "srv_api",
						Group: "destinations",
						Properties: map[string]interface{}{
							"forwardAuthToken": true,
							"strictSSL":        false,
							"name":             "srv_api",
							"url":              "~{url}",
						},
					},
				},
				BuildParams: map[string]interface{}{"builder": "grunt"},
				Parameters:  map[string]interface{}{"disk-quota": "256M", "memory": "256M"},
			}
			var modules = []*mta.Module{&moduleSrv, &moduleUI}
			mtaFile, _ := fs.ReadFile("./testdata/mta.yaml")
			// Unmarshal file
			oMta := &mta.MTA{}
			Ω(yaml.Unmarshal(mtaFile, oMta)).Should(Succeed())
			Ω(oMta.Modules).Should(HaveLen(2))
			Ω(oMta.GetModules()).Should(Equal(modules))

		})

	})

	var _ = Describe("Validation", func() {
		var _ = DescribeTable("getValidationMode", func(flag string, expectedValidateSchema, expectedValidateProject, expectedSuccess bool) {
			res1, res2, err := GetValidationMode(flag)
			Ω(res1).Should(Equal(expectedValidateSchema))
			Ω(res2).Should(Equal(expectedValidateProject))
			Ω(err == nil).Should(Equal(expectedSuccess))
		},
			Entry("default", "", true, true, true),
			Entry("schema", "schema", true, false, true),
			Entry("semantic", "semantic", true, true, true),
			Entry("invalid", "value", false, false, false),
		)

		var _ = DescribeTable("validateMtaYaml", func(projectRelPath string,
			validateSchema, validateProject, expectedSuccess bool) {
			_, err := MtaYaml(getTestPath(projectRelPath), "mta.yaml",
				validateSchema, validateProject, true, "")
			Ω(err == nil).Should(Equal(expectedSuccess))
		},
			Entry("invalid path to yaml - all", "ui5app1", true, true, false),
			Entry("invalid path to yaml - schema", "ui5app1", true, false, false),
			Entry("invalid path to yaml - project", "ui5app1", false, true, false),
			Entry("invalid path to yaml - nothing to validate", "ui5app1", false, false, true),
			Entry("valid schema", "mtahtml5", true, false, true),
			Entry("invalid project - no ui5app2 path", "mtahtml5", false, true, false),
		)

		var _ = Describe("validateMtaYaml - strict flag checks", func() {
			It("strict", func() {
				warn, err := MtaYaml(getTestPath("mtahtml5"), "mtaNotStrict.yaml",
					true, true, true, "")
				Ω(warn).Should(Equal(""))
				fmt.Println(err.Error())
				Ω(err.Error()).Should(ContainSubstring("line 8: field abc not found in type mta.Module"))
				Ω(err.Error()).Should(ContainSubstring(`line 20: mapping key "url" already defined at line 19`))
				Ω(err.Error()).Should(ContainSubstring(`the "srv_api1" property set required by the "ui" module is not defined`))
				Ω(err.Error()).Should(ContainSubstring(`the "srv" path of the "srv" module does not exist`))
			})
			It("not strict", func() {
				warn, err := MtaYaml(getTestPath("mtahtml5"), "mtaNotStrict.yaml",
					true, true, false, "")
				Ω(warn).Should(ContainSubstring("line 8: field abc not found in type mta.Module"))
				Ω(warn).Should(ContainSubstring(`line 20: mapping key "url" already defined at line 19`))
				Ω(err.Error()).ShouldNot(ContainSubstring("line 8: field abc not found in type mta.Module"))
				Ω(err.Error()).ShouldNot(ContainSubstring(`line 20: key "url" already set in map`))
				Ω(err.Error()).Should(ContainSubstring(`the "srv_api1" property set required by the "ui" module is not defined`))
				Ω(err.Error()).Should(ContainSubstring(`the "srv" path of the "srv" module does not exist`))
			})

		})

		var _ = Describe("validate - unmarshalling fails", func() {
			It("Sanity", func() {
				err, warn := validate([]byte("bad Yaml"), getTestPath("mtahtml5"),
					true, false, true, "")
				Ω(warn).Should(BeNil())
				Ω(err).ShouldNot(BeNil())
				Ω(len(err)).Should(Equal(5))
				Ω(err[0].Msg).Should(ContainSubstring("cannot unmarshal"))
			})

			It("Empty mta content", func() {
				err, warn := validate([]byte(""), getTestPath("mtahtml5"),
					true, false, true, "")
				Ω(warn).Should(BeNil())
				Ω(err).ShouldNot(BeNil())
				Ω(err[0].Msg).Should(Equal("EOF"))
			})

			It("invalid schema", func() {
				originalSchema := schemaDef
				schemaDef = []byte(`
desc: MTA DESCRIPTOR SCHEMA
# schema version must be extracted from here as there is no "version" element available to version schemas
  name: com.sap.mta.mta-schema_3.2.0 abc
`)
				_, err := MtaYaml(getTestPath("testproject"), "mta.yaml",
					true, false, true, "")
				Ω(err).Should(HaveOccurred())
				schemaDef = originalSchema
			})
		})

		Context("metadata validations", func() {
			It("validates the metadata schema", func() {
				err, warn := validate([]byte(`
ID: mymta
version: 1.0.0
_schema-version: '3.1'

parameters:
  param1: 1
parameters-metadata:
  param1:
    overwritable: abc
modules:
- name: module1
  type: html5
  path: testapp
  parameters:
    memory: 1024M
  parameters-metadata:
    memory:
      optional: 12
  properties:
    a: 1
    b: true
  properties-metadata:
    a:
      sensitive: "is it?"
      datatype: some_type
    b:
      datatype: float
`), getTestPath("mtahtml5"),
					true, false, true, "")
				Ω(warn).Should(BeNil())
				Ω(err).Should(ConsistOf(
					YamlValidationIssue{"cannot unmarshal !!str `abc` into bool", 10},
					YamlValidationIssue{`the "parameters-metadata.param1.overwritable" property must be a boolean`, 10},
					YamlValidationIssue{"cannot unmarshal !!int `12` into bool", 19},
					YamlValidationIssue{`the "modules[0].parameters-metadata.memory.optional" property must be a boolean`, 19},
					YamlValidationIssue{"cannot unmarshal !!str `is it?` into bool", 25},
					YamlValidationIssue{`the "some_type" value of the "modules[0].properties-metadata.a.datatype" enum property is invalid; expected one of the following: str,int,float,bool`, 26},
				))
			})

			It("doesn't give errors on unknown fields in properties-metadata and parameters-metadata", func() {
				err, warn := validate([]byte(`
ID: mymta
version: 1.0.0
_schema-version: '3.1'

parameters:
  param1: 1
parameters-metadata:
  param1:
    overwritable1: abc
modules:
- name: module1
  type: html5
  path: testapp
  properties:
    a: 1
    b: true
  properties-metadata:
    a:
      unknown_metadata_key: "not handled by the MBT"
      datatype: int
    b:
      datatype: bool
`), getTestPath("mtahtml5"),
					true, false, true, "")
				Ω(warn).Should(BeNil())
				Ω(err).Should(BeNil())
			})

			It("gives an error for datatype field in parameters-metadata", func() {
				err, warn := validate([]byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

parameters:
  a: 1
parameters-metadata:
  a:
    datatype: str
    optional: false

modules:
- name: ui5app1
  type: html5
  parameters:
    memory:
    env: abc
  parameters-metadata:
    memory:
      overwritable: false
      datatype: str
    env:
      optional: true
resources:
- name: res1
  type: custom
  parameters:
    m:
  parameters-metadata:
    m:
      overwritable: false
      datatype: float
`), getTestPath("mtahtml5"),
					true, false, true, "")
				Ω(warn).Should(BeNil())

				datatypeNotAllowedForParametersMetadata := fmt.Sprintf(propertyExistsErrorMsg, datatypeYamlField, parametersMetadataField)
				Ω(err).Should(ConsistOf(
					YamlValidationIssue{datatypeNotAllowedForParametersMetadata, 10},
					YamlValidationIssue{datatypeNotAllowedForParametersMetadata, 22},
					YamlValidationIssue{datatypeNotAllowedForParametersMetadata, 33},
				))
			})

		})
	})

	It("convertError", func() {
		Ω(convertError(fmt.Errorf("line 999999999999999999999999999: aaa"))).Should(BeEquivalentTo([]YamlValidationIssue{{Msg: "aaa", Line: 1}}))
	})
})
