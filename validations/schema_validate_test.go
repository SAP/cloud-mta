package validate

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/types"
	"github.com/smallfish/simpleyaml"
)

var _ = Describe("runSchemaValidations Validation", func() {

	DescribeTable("Valid runSchemaValidations", func(data string, validations ...YamlCheck) {
		validateIssues := runSchemaValidations(getMtaNode([]byte(data)), validations...)

		assertNoValidationErrors(validateIssues)
	},
		Entry("matchesRegExp", `
firstName: Donald
lastName: duck
`, property("lastName", matchesRegExp("^[A-Za-z0-9_\\-\\.]+$"))),

		Entry("required", `
firstName: Donald
lastName: duck
`, property("firstName", required())),

		Entry("Type Is String", `
firstName: Donald
lastName: duck
`, property("firstName", typeIsNotMapArray())),

		Entry("Type Is Bool", `
name: bisli
registered: false
`, property("registered", typeIsBoolean())),

		Entry("Type Is Array", `
firstName:
   - 1
   - 2
   - 3
lastName: duck
`, property("firstName", typeIsArray())),

		Entry("sequenceFailFast", `
firstName: Hello
lastName: World
`, property("firstName", sequence(required(), matchesRegExp("^[A-Za-z0-9]+$")))),

		Entry("Type Is Map", `
firstName:
   - 1
   - 2
   - 3
lastName:
   a : 1
   b : 2
`, property("lastName", typeIsMap())),

		Entry("For Each", `
firstName: Hello
lastName: World
classes:
 - name: biology
   room: MR113

 - name: history
   room: MR225

`, property("classes", sequence(
			required(),
			typeIsArray(),
			forEach(
				property("name", required()),
				property("room", matchesRegExp("^MR[0-9]+$")))))),

		Entry("optional Exists", `
firstName: Donald
lastName: duck
`, property("firstName", optional(typeIsNotMapArray()))),

		Entry("optional Missing", `
lastName: duck
`, property("firstName", optional(typeIsNotMapArray()))),
	)

	DescribeTable("Invalid runSchemaValidations", func(data, message string, line int, validations ...YamlCheck) {
		validateIssues := runSchemaValidations(getMtaNode([]byte(data)), validations...)

		expectSingleValidationError(validateIssues, message, line)
	},
		Entry("matchesRegExp", `
firstName: Donald
lastName: duck
`, `the "Donald" value of the "root.firstName" property does not match the "^[0-9_\-\.]+$" pattern`, 2,
			property("firstName", matchesRegExp("^[0-9_\\-\\.]+$"))),

		Entry("required", `
firstName: Donald
lastName: duck
`, `missing the "age" required property in the root .yaml node`, 2,
			property("age", required())),

		Entry("TypeIsString", `
firstName:
   - 1
   - 2
   - 3
lastName: duck
`, `the "root.firstName" property must be a string`, 3,
			property("firstName", typeIsNotMapArray())),

		Entry("TypeIsBool", `
name: bamba
registered: 123
`, `the "root.registered" property must be a boolean`, 3,
			property("registered", typeIsBoolean())),

		Entry("typeIsArray", `
firstName:
   - 1
   - 2
   - 3
lastName: duck
`, `the "root.lastName" property must be an array`, 6,
			property("lastName", typeIsArray())),

		Entry("typeIsMap", `
firstName:
   - 1
   - 2
   - 3
lastName:
   a : 1
   b : 2
`, `the "root.firstName" property must be a map`, 3,
			property("firstName", typeIsMap())),

		Entry("sequenceFailFast", `
firstName: Hello
lastName: World
`, `missing the "missing" required property in the root .yaml node`, 2,
			property("missing", sequenceFailFast(
				required(),
				// This second validation should not be executed as sequence breaks early.
				matchesRegExp("^[0-9]+$")))),

		Entry("OptionalExists", `
firstName:
  - 1
  - 2
lastName: duck
`, `the "root.firstName" property must be a string`, 3,
			property("firstName", optional(typeIsNotMapArray()))),
	)

	It("InvalidYamlHandling", func() {
		data := []byte(`
firstName: Donald
  lastName: duck # invalid indentation
		`)
		issues := runSchemaValidations(getMtaNode([]byte(data)), property("lastName", required()))
		Ω(len(issues)).Should(Equal(1))
	})

	It("ForEachInValid", func() {
		data := []byte(`
firstName: Hello
lastName: World
classes:
 - name: biology
   room: oops

 - room: 225

`)
		validations := property("classes", sequence(
			required(),
			typeIsArray(),
			forEach(
				property("name", required()),
				property("room", matchesRegExp("^[0-9]+$")))))

		validateIssues := runSchemaValidations(getMtaNode([]byte(data)), validations)

		expectMultipleValidationError(validateIssues,
			[]string{
				`the "oops" value of the "classes[0].room" property does not match the "^[0-9]+$" pattern`,
				`missing the "name" required property in the classes[1] .yaml node`})
	})
})

var _ = DescribeTable("GetLiteralStringValue", func(data string, matcher GomegaMatcher) {
	y, _ := simpleyaml.NewYaml([]byte(data))
	value := getLiteralStringValue(y)
	Ω(value).Should(matcher)
},
	Entry("Invalid", `
  [a,b]
`, BeEmpty()),
	Entry("Valid", fmt.Sprintf("%g", 0.55), Equal("0.55")),
)
