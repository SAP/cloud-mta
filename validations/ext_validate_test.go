package validate

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("MTAEXT validation tests", func() {
	var _ = DescribeTable("validate Mtaext", func(projectRelPath string, fileName string,
		validateSchema, validateSemantic, expectedSuccess bool) {
		_, err := Mtaext(getTestPath(projectRelPath), getTestPath(projectRelPath, fileName), validateSchema, validateSemantic, true, "")
		if expectedSuccess {
			Ω(err).Should(Succeed())
		} else {
			Ω(err).Should(HaveOccurred())
		}
	},
		Entry("validate schema and semantic when path to extension is invalid should fail", "ui5app1", "my.mtaext", true, true, false),
		Entry("validate schema when path to extension is invalid should fail", "ui5app1", "my.mtaext", true, false, false),
		Entry("validate semantic when path to extension is invalid should fail", "ui5app1", "my.mtaext", false, true, false),
		Entry("nothing to validate when path to extension is invalid should not fail", "ui5app1", "my.mtaext", false, false, true),
		Entry("validate schema should succeed when extension schema is valid", "mtahtml5", "my.mtaext", true, false, true),
		Entry("validate semantic when module is extended twice should fail", "mtahtml5", "my.mtaext", false, true, false),
	)

	var _ = Describe("validate Mtaext - strict flag checks", func() {
		It("strict", func() {
			warn, err := Mtaext(getTestPath("mtahtml5"), getTestPath("mtahtml5", "myNotStrict.mtaext"),
				true, true, true, "")
			Ω(warn).Should(Equal(""))
			Ω(err).Should(HaveOccurred())
			message := err.Error()
			Ω(message).Should(ContainSubstring("line 11: field parameters-metadata not found in type mta.ModuleExt"))
			Ω(message).Should(ContainSubstring("line 18: field properties-metadata not found in type mta.ModuleExt"))
			Ω(message).Should(ContainSubstring(`line 26: mapping key "A" already defined at line 25`))
			Ω(message).Should(ContainSubstring(`line 30: mapping key "B" already defined at line 29`))
			Ω(message).Should(ContainSubstring("line 32: field badprop not found in type mta.ResourceExt"))
			Ω(message).Should(ContainSubstring("line 15: %s", fmt.Sprintf(nameAlreadyExtendedMsg, "ui5app", "module", "another", "module", 7)))
		})
		It("not strict", func() {
			warn, err := Mtaext(getTestPath("mtahtml5"), getTestPath("mtahtml5", "myNotStrict.mtaext"),
				true, true, false, "")
			message := warn
			Ω(message).Should(ContainSubstring("line 11: field parameters-metadata not found in type mta.ModuleExt"))
			Ω(message).Should(ContainSubstring("line 18: field properties-metadata not found in type mta.ModuleExt"))
			Ω(message).Should(ContainSubstring(`line 26: mapping key "A" already defined at line 25`))
			Ω(message).Should(ContainSubstring(`line 30: mapping key "B" already defined at line 29`))
			Ω(message).Should(ContainSubstring("line 32: field badprop not found in type mta.ResourceExt"))
			Ω(message).ShouldNot(ContainSubstring("line 15:"))

			Ω(err).Should(HaveOccurred())
			message = err.Error()
			Ω(message).Should(ContainSubstring("line 15: %s", fmt.Sprintf(nameAlreadyExtendedMsg, "ui5app", "module", "another", "module", 7)))
			Ω(message).ShouldNot(ContainSubstring("line 11:"))
			Ω(message).ShouldNot(ContainSubstring("line 18:"))
			Ω(message).ShouldNot(ContainSubstring("line 26:"))
			Ω(message).ShouldNot(ContainSubstring("line 30:"))
			Ω(message).ShouldNot(ContainSubstring("line 32:"))
		})
	})

	var _ = Describe("validateExt - unmarshalling fails", func() {
		It("Sanity", func() {
			err, warn := validateExt([]byte("bad Yaml"), getTestPath("mtahtml5"), "my.mtaext",
				true, false, true, "")
			Ω(warn).Should(BeNil())
			Ω(err).ShouldNot(BeNil())
			Ω(len(err)).Should(Equal(5))
			Ω(err[0].Msg).Should(ContainSubstring("cannot unmarshal"))
		})

		It("Only returns unmarshaling error once", func() {
			err, warn := validateExt([]byte(`- bad Yaml
a b`), getTestPath("mtahtml5"), "my.mtaext",
				true, false, true, "")
			Ω(warn).Should(BeNil())
			Ω(err).ShouldNot(BeNil())

			// Check we don't get the same error twice
			errorsMap := make(map[string]interface{})
			for _, value := range err {
				if _, ok := errorsMap[value.Msg]; ok {
					Ω(ok).Should(BeFalse(), fmt.Sprintf("Got the same message twice: %s", value.Msg))
				}
				errorsMap[value.Msg] = nil
			}
		})

		It("Empty mta content", func() {
			err, warn := validateExt([]byte(""), "a.mtaext", getTestPath("mtahtml5"),
				true, false, true, "")
			Ω(warn).Should(BeNil())
			Ω(err).ShouldNot(BeNil())
			Ω(err[0].Msg).Should(Equal("EOF"))
		})

		When("schema definition is different", func() {
			var originalSchema []byte
			BeforeEach(func() {
				originalSchema = extSchemaDef
			})

			It("invalid schema definition", func() {
				extSchemaDef = []byte(`
desc: MTA DESCRIPTOR SCHEMA
# schema version must be extracted from here as there is no "version" element available to version schemas
  name: com.sap.mta.mta-schema_3.2.0 abc
`)
				_, err := Mtaext(getTestPath("mtahtml5"), getTestPath("mtahtml5", "my.mtaext"),
					true, false, true, "")
				Ω(err).Should(HaveOccurred())
				Ω(err.Error()).Should(ContainSubstring("validation failed when parsing the MTA schema file"))
			})

			AfterEach(func() {
				extSchemaDef = originalSchema
			})
		})

	})

	It("file extension which is not mtaext does not cause an error", func() {
		err, warn := validateExt([]byte(`
ID: mymtaext
extends: somemta
_schema-version: '3.1'
`), getTestPath("mtahtml5"), "ext.yaml",
			false, true, true, "")
		Ω(warn).Should(BeNil())
		Ω(err).Should(BeNil())
	})

	It("Unallowed fields in extension file return errors", func() {
		err, warn := validateExt([]byte(`
ID: myext
extends: mymta
_schema-version: '3.1'

modules:
- name: mymodule
  provides:
  - name: myprovides
    public: true
  requires:
  - name: myreq
    list: destinations
  hooks:
  - name: myhook
    requires:
    - name: myhookreq
      list: somelistname
resources:
- name: myresource
  optional: true
  requires:
  - name: myresreq
    list: abc
- name: myresource1
  requires:
  - name: req1
    properties-metadata:
  - name: req2
    parameters-metadata:
- name: myresource2
  requires:
  - name: req1
    properties-metadata:
      a:
`), getTestPath("mtahtml5"), "my.mtaext",
			true, false, true, "")
		Ω(warn).Should(BeNil())
		Ω(err).Should(ConsistOf(
			YamlValidationIssue{fmt.Sprintf(propertyExistsErrorMsg, "public", "modules[0].provides[0]"), 10, 5},
			YamlValidationIssue{fmt.Sprintf(propertyExistsErrorMsg, "list", "modules[0].requires[0]"), 13, 5},
			YamlValidationIssue{fmt.Sprintf(propertyExistsErrorMsg, "list", "modules[0].hooks[0].requires[0]"), 18, 7},
			YamlValidationIssue{`field optional not found in type mta.ResourceExt`, 21, 0},
			YamlValidationIssue{fmt.Sprintf(propertyExistsErrorMsg, "list", "resources[0].requires[0]"), 24, 5},
			YamlValidationIssue{fmt.Sprintf(propertyExistsErrorMsg, "properties-metadata", "resources[1].requires[0]"), 28, 5},
			YamlValidationIssue{fmt.Sprintf(propertyExistsErrorMsg, "parameters-metadata", "resources[1].requires[1]"), 30, 5},
			YamlValidationIssue{fmt.Sprintf(propertyExistsErrorMsg, "properties-metadata", "resources[2].requires[0]"), 34, 5},
		))
	})
})
