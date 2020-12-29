package validate

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func assertNoSchemaIssues(errors []YamlValidationIssue) {
	Ω(len(errors)).Should(Equal(0), "Schema issues detected")
}

var _ = Describe("Schema tests Issues", func() {

	var _ = DescribeTable("Schema issues",
		func(schema string, message string) {
			_, schemaIssues := buildValidationsFromSchemaText([]byte(schema))
			expectSingleValidationError(schemaIssues, message, 0, 0) // We don't pass the line and column numbers to schema issues
		},
		Entry("Parsing", `
type: map
# bad indentation
 mapping:
   firstName:  {required: true}`, `validation failed when parsing the MTA schema file: unmarshal []byte to yaml failed: yaml: line 3: did not find expected key`),

		Entry("Mapping", `
type: map
mapping: NotAMap`, `invalid .yaml file schema: the mapping node must be a map`),

		Entry("SchemaSequenceIssue", `
type: seq
sequence: NotASequence
`, `invalid .yaml file schema: the sequence node must be an array`),

		Entry("sequence One Item - more than 1 item", `
type: seq
sequence:
- 1
- 2
`, `invalid .yaml file schema: the sequence node must have exactly one item`),

		Entry("sequence One Item - empty sequence", `
type: seq
sequence: []
`, `invalid .yaml file schema: the sequence node must have exactly one item`),

		Entry("required value not bool", `
type: map
mapping:
  firstName:  {required: 123}
`, `invalid .yaml file schema: the required node must be a boolean `),

		Entry("sequence NestedTypeNotString", `
type: map
mapping:
  firstName:  {type: [1,2] }
`, `invalid .yaml file schema: the type node must be a string`),

		Entry("Pattern NotString", `
type: map
mapping:
  firstName:  {pattern: [1,2] }
`, `invalid .yaml file schema: the pattern node must be a string`),

		Entry("Pattern InvalidRegex", `
type: map
mapping:
  firstName:  {required: true, pattern: '/[a-zA-Z+/'}
`, "invalid .yaml file schema: the pattern node is invalid because: error parsing regexp: missing closing ]: `[a-zA-Z+`"),

		Entry("Enum NotString", `
type: str
enum:
  duck : 1
  dog  : 2
`, `invalid .yaml file schema: enums values must be listed as an array`),

		Entry("Enum ValueNotSimple", `
type: str
enum:
  [duck, [dog, cat]]
`, `invalid .yaml file schema: enum values must be simple`),
	)

	var _ = DescribeTable("Valid input",
		func(schema, input string) {
			schemaValidations, schemaIssues := buildValidationsFromSchemaText([]byte(schema))
			assertNoSchemaIssues(schemaIssues)
			node, _ := getContentNode([]byte(input))
			validateIssues := runSchemaValidations(node, schemaValidations...)
			assertNoValidationErrors(validateIssues)
		},
		Entry("required", `
type: map
mapping:
 firstName:  {required: true}
`, `
firstName: Donald
lastName: duck`),
		Entry("required field inside mapping, when parent is null", `
type: map
mapping:
  inner:
    type: map
    mapping:
      firstName:  {required: true}
`, `{}`),
		Entry("Enum value", `
enum:
  - duck
  - dog
`, `duck`),
		Entry("null enum value", `
type: map
mapping:
  animal:
   enum:
    - duck
    - dog
`, `{}`),
		Entry("sequence", `
type: seq
sequence:
- type: map
  mapping:
    name: {required: true}
`, `
- name: Donald
  lastName: duck

- name: Bugs
  lastName: Bunny

`),
		Entry("sequence of mapping with default value - value is sequence", `
type: seq
sequence:
- type: map
  mapping:
    =:
      type: seq
      sequence:
      - type: bool
`, `
- firstKey:
  - true
  - false
  key2: []
`),
		Entry("mapping with default value - value is mapping", `
type: map
mapping:
  =:
    type: map
    mapping:
      name: {required: true}
`, `
firstKey:
  name: Donald
  lastName: duck
key2:
  name: Bugs
  lastName: Bunny
`),
		Entry("mapping with default value - value is any", `
type: map
mapping:
  =:
    type: any
`, `
firstMapKey_map:
  name: Donald
  lastName: duck
key2str: a
key3int: 1
key4seq:
- 1
- 2
`),
		Entry("Pattern", `
type: map
mapping:
   firstName:  {required: true, pattern: '/^[a-zA-Z]+$/'}
`, `
firstName: Donald
lastName: duck
`),
		Entry("Pattern when value is null", `
type: map
mapping:
   firstName:  {pattern: '/^[a-zA-Z]+$/'}
`, `
lastName: duck
`),
		Entry("optional", `
type: map
mapping:
   firstName:  {required: false, pattern: '/^[a-zA-Z]+$/'}
`, `
lastName: duck
`),
		Entry("Type Is Bool", `
type: map
mapping:
   isHappy:  {type: bool}
`, `
firstName: Tim
isHappy: false
`),
	)

	var _ = DescribeTable("Invalid input",
		func(schema, input, message string, line int, column int) {
			schemaValidations, schemaIssues := buildValidationsFromSchemaText([]byte(schema))
			assertNoSchemaIssues(schemaIssues)
			node, _ := getContentNode([]byte(input))
			validateIssues := runSchemaValidations(node, schemaValidations...)
			expectSingleValidationError(validateIssues, message, line, column)
		},
		Entry("required", `
type: map
mapping:
   age:  {required: true}
`, `
firstName: Donald
lastName: duck
`, `missing the "age" required property in the root .yaml node`, 2, 1),

		Entry("required mapping field", `
type: map
mapping:
  inner:
    type: map
    required: true
    mapping:
      firstName: {type: string}
`, `{}`, `missing the "inner" required property in the root .yaml node`, 1, 1),

		Entry("required mapping field with inner required field", `
type: map
mapping:
  inner:
    type: map
    required: true
    mapping:
      firstName: {required: true}
`, `{}`, `missing the "inner" required property in the root .yaml node`, 1, 1),

		Entry("Enum", `
type: str
enum:
   - duck
   - dog
   - cat
   - mouse
   - elephant
`, `bird`, `the "bird" value of the "root" enum property is invalid; expected one of the following: duck,dog,cat,mouse`, 1, 1),

		Entry("sequence", `
type: seq
sequence:
- type: map
  mapping:
    name: {required: true}
`, `
- name: Donald
  lastName: duck

- age: 80
  lastName: Bunny
`, `missing the "name" required property in the root[1] .yaml node`, 5, 3),

		Entry("sequence of mapping with default value - value is sequence", `
type: seq
sequence:
- type: map
  mapping:
    =:
      type: seq
      sequence:
      - type: bool
`, `
- firstKey:
  - true
  - b
  key2: []
`, `the "[0].firstKey[1]" property must be a boolean`, 4, 5),

		Entry("mapping with default value - value is mapping", `
type: map
mapping:
  =:
    type: map
    mapping:
      name: {required: true}
`, `
firstKey:
  name: Donald
  lastName: duck
key2:
  age: 30
  lastName: Bunny
`, `missing the "name" required property in the root.key2 .yaml node`, 6, 3),

		Entry("Pattern", `
type: map
mapping:
   age:  {pattern: '/^[0-9]+$/'}
`, `
name: Bamba
age: NaN
`, `the "NaN" value of the "root.age" property does not match the "^[0-9]+$" pattern`, 3, 6),

		Entry("optional With Pattern", `
type: map
mapping:
   firstName:  {required: false, pattern: '/^[a-zA-Z]+$/'}
`, `
firstName: Donald123
lastName: duck
`, `the "Donald123" value of the "root.firstName" property does not match the "^[a-zA-Z]+$" pattern`, 2, 12),

		Entry("Type Is Bool", `
type: map
mapping:
   isHappy:  {type: bool}
`, `
firstName: John
isHappy: 123
`, `the "root.isHappy" property must be a boolean`, 3, 10),
	)
})
