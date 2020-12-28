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

		Entry("Type is String when property doesn't exist", `
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

	DescribeTable("Invalid runSchemaValidations", func(data, message string, line int, column int, validations ...YamlCheck) {
		node, _ := getContentNode([]byte(data))
		validateIssues := runSchemaValidations(node, validations...)

		expectSingleValidationError(validateIssues, message, line, column)
	},
		Entry("matchesRegExp", `
firstName: Donald
lastName: duck
`, `the "Donald" value of the "root.firstName" property does not match the "^[0-9_\-\.]+$" pattern`, 2, 12,
			property("firstName", matchesRegExp("^[0-9_\\-\\.]+$"))),

		Entry("required", `
firstName: Donald
lastName: duck
`, `missing the "age" required property in the root .yaml node`, 2, 1,
			property("age", required())),

		Entry("doesNotExist on property name", `
firstName: Donald
lastName: 
  - duck
`, fmt.Sprintf(propertyExistsErrorMsg, "lastName", "root"), 3, 1,
			propertyName("lastName", doesNotExist())),

		Entry("doesNotExist on property value", `
firstName: Donald
lastName: 
  - duck
`, fmt.Sprintf(propertyExistsErrorMsg, "lastName", "root"), 4, 3,
			property("lastName", doesNotExist())),

		Entry("TypeIsString", `
firstName:
   - 1
   - 2
   - 3
lastName: duck
`, `the "root.firstName" property must be a string`, 3, 4,
			property("firstName", typeIsNotMapArray())),

		Entry("TypeIsBool", `
name: bamba
registered: 123
`, `the "root.registered" property must be a boolean`, 3, 13,
			property("registered", typeIsBoolean())),

		Entry("typeIsArray", `
firstName:
   - 1
   - 2
   - 3
lastName: duck
`, `the "root.lastName" property must be an array`, 6, 11,
			property("lastName", typeIsArray())),

		Entry("typeIsMap", `
firstName:
   - 1
   - 2
   - 3
lastName:
   a : 1
   b : 2
`, `the "root.firstName" property must be a map`, 3, 4,
			property("firstName", typeIsMap())),

		Entry("sequenceFailFast", `
firstName: Hello
lastName: World
`, `missing the "missing" required property in the root .yaml node`, 2, 1,
			property("missing", sequenceFailFast(
				required(),
				// This second validation should not be executed as sequence breaks early.
				matchesRegExp("^[0-9]+$")))),

		Entry("OptionalExists", `
firstName:
  - 1
  - 2
lastName: duck
`, `the "root.firstName" property must be a string`, 3, 3,
			property("firstName", optional(typeIsNotMapArray()))),
	)

	It("InvalidYamlHandling", func() {
		data := []byte(`
firstName: Donald
  lastName: duck # invalid indentation
		`)
		node, _ := getContentNode(data)
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

		node, _ := getContentNode(data)
		validateIssues := runSchemaValidations(node, validations)

		Ω(validateIssues).Should(ConsistOf(
			YamlValidationIssue{`the "oops" value of the "classes[0].room" property does not match the "^[0-9]+$" pattern`, 6, 10},
			YamlValidationIssue{`missing the "name" required property in the classes[1] .yaml node`, 8, 4},
			YamlValidationIssue{`the "optionalClasses.english" property must be a boolean`, 13, 12},
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
lastName: World
aliases:
  - &alias1
    veryLastName: Bye
prop1: *alias1
   `)
	It("sanity", func() {
		node, _ := getContentNode(data)
		node = getPropByName(node, "lastName")
		Ω(node.Value).Should(Equal("lastName"))
		Ω(node.Line).Should(Equal(3))
	})
	It("nil node", func() {
		Ω(getPropByName(nil, "x")).Should(BeNil())
	})
	It("property not exists", func() {
		node, _ := getContentNode(data)
		Ω(getPropByName(node, "x")).Should(BeNil())
	})
	It("aliases usage", func() {
		node, _ := getContentNode(data)
		prop1 := getPropValueByName(node, "prop1")
		Ω(prop1).ShouldNot(BeNil())
		Ω(getPropByName(prop1, "veryLastName")).ShouldNot(BeNil())
		Ω(getPropByName(prop1, "y")).Should(BeNil())
	})
	It("aliases usage; aliases content is nil", func() {
		node, _ := getContentNode(data)
		prop := getPropValueByName(node, "prop1")
		Ω(prop).ShouldNot(BeNil())
		prop.Alias = nil
		prop.Content = nil
		Ω(getPropByName(prop, "veryLastName")).Should(BeNil())
	})
})

var _ = DescribeTable("sort validation issues", func(lines []int, columns []int) {
	var issues YamlValidationIssues
	if lines == nil {
		issues = nil
	} else {
		if len(columns) == 0 {
			columns = []int{0}
		}
		issues = make(YamlValidationIssues, len(lines)*len(columns))
		for i, line := range lines {
			for j, column := range columns {
				issues[i*len(columns)+j] = YamlValidationIssue{Line: line, Column: column, Msg: fmt.Sprintf("line %d col %d issue", line, column)}
			}
		}

		issues.Sort()

		// Check it's sorted
		isSorted := sort.SliceIsSorted(issues, func(i, j int) bool {
			return (issues[i].Line < issues[j].Line) ||
				(issues[i].Line == issues[j].Line && issues[i].Column < issues[j].Column)
		})
		Ω(isSorted).Should(Equal(true), fmt.Sprintf("slice is not sorted: %v", issues))

		// Check the values are correct
		for _, issue := range issues {
			Ω(issue.Msg).Should(Equal(fmt.Sprintf("line %d col %d issue", issue.Line, issue.Column)))
		}
	}
},
	Entry("nil slice", nil, nil),
	Entry("empty slice", []int{}, nil),
	Entry("slice with one value", []int{300}, nil),
	Entry("sorted slice", []int{1, 2, 3, 4, 12, 65}, nil),
	Entry("backwards sorted slice", []int{12, 2, 0}, nil),
	Entry("slice with equal values", []int{1, 1, 4, 23, 32, 5, 32}, nil),
	Entry("slice with equal values and several unsorted columns for each line", []int{1, 1, 4, 23, 32, 5, 32}, []int{2, 1}),
	Entry("unsorted slice", []int{3, 84, 600, 2, 0, 5, 0, 7, 5, 12}, nil),
	Entry("unsorted slice with several columns for each line", []int{3, 84, 600, 2, 0, 5, 0, 7, 5, 12}, []int{5, 6}),
	Entry("unsorted slice with several unsorted columns for each line", []int{3, 84, 600, 2, 0, 5, 0, 7, 5, 12}, []int{2, 1}),
)
