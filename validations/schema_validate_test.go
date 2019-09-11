package validate

import (
	"fmt"
	"sort"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/types"
	"github.com/smallfish/simpleyaml"
)

var _ = Describe("runSchemaValidations Validation", func() {

	DescribeTable("Valid runSchemaValidations", func(data string, validations ...YamlCheck) {
		node, _ := getContentNode([]byte(data))
		validateIssues := runSchemaValidations(node, validations...)

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

		Entry("doesNotExist on property value", `
firstName: Donald
lastName: duck
`, property("middleName", doesNotExist())),

		Entry("doesNotExist on property name", `
firstName: Donald
lastName: duck
`, propertyName("middleName", doesNotExist())),

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

		Entry("For Each Property", `
firstName: Hello
lastName: World
classes:
  biology:
    grade: 90
    optional: true
    room: MR113

  history:
    grade: 83
    optional: false
    room: MR225

`, property("classes", sequence(
			required(),
			typeIsMap(),
			forEachProperty(
				property("grade", required()),
				property("optional", typeIsBoolean()),
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
		node, _ := getContentNode([]byte(data))
		validateIssues := runSchemaValidations(node, validations...)

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

		Entry("doesNotExist on property name", `
firstName: Donald
lastName: 
  - duck
`, fmt.Sprintf(propertyExistsErrorMsg, "lastName", "root"), 3,
			propertyName("lastName", doesNotExist())),

		Entry("doesNotExist on property value", `
firstName: Donald
lastName: 
  - duck
`, fmt.Sprintf(propertyExistsErrorMsg, "lastName", "root"), 4,
			property("lastName", doesNotExist())),

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
		node, _ := getContentNode([]byte(data))
		issues := runSchemaValidations(node, property("lastName", required()))
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

optionalClasses:
  biology: true
  history: false
  english: unknown
`)
		validations := sequence(
			property("classes", sequence(
				required(),
				typeIsArray(),
				forEach(
					property("name", required()),
					property("room", matchesRegExp("^[0-9]+$"))))),
			property("optionalClasses",
				forEachProperty(typeIsBoolean())))

		node, _ := getContentNode([]byte(data))
		validateIssues := runSchemaValidations(node, validations)

		Ω(validateIssues).Should(ConsistOf(
			YamlValidationIssue{`the "oops" value of the "classes[0].room" property does not match the "^[0-9]+$" pattern`, 6},
			YamlValidationIssue{`missing the "name" required property in the classes[1] .yaml node`, 8},
			YamlValidationIssue{`the "optionalClasses.english" property must be a boolean`, 13},
		))
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

var _ = Describe("getPropByName", func() {
	var data = []byte(`
firstName: Hello
lastName: World`)
	It("sanity", func() {
		node, _ := getContentNode([]byte(data))
		node = getPropByName(node, "lastName")
		Ω(node.Value).Should(Equal("lastName"))
		Ω(node.Line).Should(Equal(3))
	})
	It("nil node", func() {
		Ω(getPropByName(nil, "x")).Should(BeNil())
	})
	It("property not exists", func() {
		node, _ := getContentNode([]byte(data))
		Ω(getPropByName(node, "x")).Should(BeNil())
	})
})

var _ = DescribeTable("sort validation issues", func(lines []int) {
	var issues YamlValidationIssues
	if lines == nil {
		issues = nil
	} else {
		issues = make(YamlValidationIssues, len(lines))
		for i, line := range lines {
			issues[i] = YamlValidationIssue{Line: line, Msg: fmt.Sprintf("line %d issue", line)}
		}

		issues.Sort()

		// Check it's sorted
		isSorted := sort.SliceIsSorted(issues, func(i, j int) bool {
			return issues[i].Line < issues[j].Line
		})
		Ω(isSorted).Should(Equal(true), fmt.Sprintf("slice is not sorted: %v", issues))

		// Check the values are correct
		for _, issue := range issues {
			Ω(issue.Msg).Should(Equal(fmt.Sprintf("line %d issue", issue.Line)))
		}
	}
},
	Entry("nil slice", nil),
	Entry("empty slice", []int{}),
	Entry("slice with one value", []int{300}),
	Entry("sorted slice", []int{1, 2, 3, 4, 12, 65}),
	Entry("backwards sorted slice", []int{12, 2, 0}),
	Entry("slice with equal values", []int{1, 1, 4, 23, 32, 5, 32}),
	Entry("unsorted slice", []int{3, 84, 600, 2, 0, 5, 0, 7, 5, 12}),
)
