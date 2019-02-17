package validate

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

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
			mtaFile, _ := ioutil.ReadFile("./testdata/mta.yaml")
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
				validateSchema, validateProject, true)
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
					true, true, true)
				Ω(warn).Should(Equal(""))
				Ω(err.Error()).Should(ContainSubstring("line 8: field abc not found in type mta.Module"))
				Ω(err.Error()).Should(ContainSubstring(`line 20: key "url" already set in map`))
				Ω(err.Error()).Should(ContainSubstring(`the "srv_api1" property set required by the "ui" module is not defined`))
				Ω(err.Error()).Should(ContainSubstring(`the "srv" path of the "srv" module does not exist`))
			})
			It("not strict", func() {
				warn, err := MtaYaml(getTestPath("mtahtml5"), "mtaNotStrict.yaml",
					true, true, false)
				Ω(warn).Should(ContainSubstring("line 8: field abc not found in type mta.Module"))
				Ω(warn).Should(ContainSubstring(`line 20: key "url" already set in map`))
				Ω(err.Error()).ShouldNot(ContainSubstring("line 8: field abc not found in type mta.Module"))
				Ω(err.Error()).ShouldNot(ContainSubstring(`line 20: key "url" already set in map`))
				Ω(err.Error()).Should(ContainSubstring(`the "srv_api1" property set required by the "ui" module is not defined`))
				Ω(err.Error()).Should(ContainSubstring(`the "srv" path of the "srv" module does not exist`))
			})

		})

		var _ = Describe("validate - unmarshalling fails", func() {
			It("Sanity", func() {
				err, warn := validate([]byte("bad Yaml"), getTestPath("mtahtml5"),
					true, false, false, func(mtaContent []byte, mtaStr interface{}) error {
						return errors.New("err")
					})
				Ω(warn).Should(BeNil())
				Ω(err).ShouldNot(BeNil())
				Ω(len(err)).Should(Equal(1))
				Ω(err[0].Msg).Should(ContainSubstring("err"))
			})

			It("invalid schema", func() {
				originalSchema := schemaDef
				schemaDef = []byte(`
desc: MTA DESCRIPTOR SCHEMA
# schema version must be extracted from here as there is no "version" element available to version schemas
  name: com.sap.mta.mta-schema_3.2.0 abc
`)
				_, err := MtaYaml(getTestPath("testproject"), "mta.yaml",
					true, false, true)
				Ω(err).Should(HaveOccurred())
				schemaDef = originalSchema
			})
		})
	})
})
