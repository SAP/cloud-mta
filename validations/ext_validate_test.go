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

	It("bad file extension", func() {
		err, warn := validateExt([]byte(`
ID: mymtaext
extends: somemta
_schema-version: '3.1'
`), getTestPath("mtahtml5"), "ext.yaml",
			false, true, true, "")
		Ω(warn).Should(BeNil())
		Ω(err).Should(ConsistOf(YamlValidationIssue{badExtensionErrorMsg, 0}))
	})

	DescribeTable("validateExtFileName", func(filename string, expectedSuccess bool) {
		if expectedSuccess {
			errIssues, warnIssues := validateExtFileName(filename, true)
			Ω(errIssues).Should(BeNil())
			Ω(warnIssues).Should(BeNil())
			errIssues, warnIssues = validateExtFileName(filename, false)
			Ω(errIssues).Should(BeNil())
			Ω(warnIssues).Should(BeNil())
		} else {
			errIssues, warnIssues := validateExtFileName(filename, true)
			Ω(errIssues).ShouldNot(BeNil())
			Ω(len(errIssues)).Should(Equal(1))
			Ω(warnIssues).Should(BeNil())
			errIssues, warnIssues = validateExtFileName(filename, false)
			Ω(errIssues).Should(BeNil())
			Ω(warnIssues).ShouldNot(BeNil())
			Ω(len(warnIssues)).Should(Equal(1))
		}
	},
		Entry("file name is mtaext", "mtaext", false),
		Entry("file name has mtaext extension", "a.mtaext", true),
		Entry("file name has yaml extension", "ext.yaml", false),
		Entry("file name is mtaext and has yaml extension", "mtaext.yaml", false),
	)
})
