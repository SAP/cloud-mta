package validate

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"regexp"
	"sort"
	"strings"

	"github.com/smallfish/simpleyaml"
)

const (
	propertyExistsErrorMsg = `the "%s" key is not allowed inside the "%s"`
)

// YamlValidationIssue - specific issue
type YamlValidationIssue struct {
	// Msg - message content
	Msg string
	// Line - line number indicating issue
	Line int
	// Column - column number of the issue
	Column int
}

// YamlValidationIssues - list of issue's
type YamlValidationIssues []YamlValidationIssue

func (issues YamlValidationIssues) String() string {
	var messages []string
	for _, issue := range issues {
		messages = append(messages, fmt.Sprintf("line %d: %s", issue.Line, issue.Msg))
	}
	return strings.Join(messages, "\n")
}

// Sort sorts the validation issues by line and column number
func (issues YamlValidationIssues) Sort() {
	sort.SliceStable(issues, func(i, j int) bool {
		return (issues[i].Line < issues[j].Line) ||
			(issues[i].Line == issues[j].Line && issues[i].Column < issues[j].Column)
	})
}

// YamlCheck - validation check function type
type YamlCheck func(yNode, yParentNode *yaml.Node, path []string) YamlValidationIssues

// DSL method to execute validations on a sub node(property) of a YAML tree.
// Can be nested to check properties farther and farther down the tree.
func property(propName string, checks ...YamlCheck) YamlCheck {
	return func(yNode, yParentNode *yaml.Node, path []string) YamlValidationIssues {
		var issues YamlValidationIssues
		yPropNode := getPropValueByName(yNode, propName)

		// Will perform all the validations without stopping
		for _, check := range checks {
			newIssues := check(yPropNode, yNode, append(path, propName))
			issues = append(issues, newIssues...)
		}

		return issues
	}
}

// DSL method to execute validations on the name of a sub node(property) of a YAML tree.
// Can be nested to check properties farther and farther down the tree.
func propertyName(propName string, checks ...YamlCheck) YamlCheck {
	return func(yNode, yParentNode *yaml.Node, path []string) YamlValidationIssues {
		var issues YamlValidationIssues
		yPropNode := getPropByName(yNode, propName)

		// Will perform all the validations without stopping
		for _, check := range checks {
			newIssues := check(yPropNode, yNode, append(path, propName))
			issues = append(issues, newIssues...)
		}

		return issues
	}
}

func getPropByName(node *yaml.Node, name string) *yaml.Node {
	if node == nil {
		return nil
	}

	content := node.Content
	if content == nil && node.Alias != nil {
		content = node.Alias.Content
	}

	if content == nil {
		return nil
	}

	// we start from some key and searched key is still not found
	key := true

	// properties are listed in the Content of node as slice of key, value, key, value,...
	for _, propNode := range content {
		// current is key and its value equals to searched name => key found
		if key && propNode.Value == name {
			return propNode
		}
		// key ->value, value->key
		key = !key
	}
	return nil
}

func getPropValueByName(node *yaml.Node, name string) *yaml.Node {
	if node == nil || node.Content == nil {
		return nil
	}

	// we start from some key and searched key is still not found
	key := true
	keyFound := false

	// properties are listed in the Content of node as slice of key, value, key, value,...
	for _, propNode := range node.Content {
		if keyFound {
			// previous node one is found key, current is its value
			return propNode
		}
		// current is key and its value equals to searched name => key found
		if key && propNode.Value == name {
			keyFound = true
		}
		// key ->value, value->key
		key = !key
	}
	return nil
}

func getPropContent(node *yaml.Node, name string) []*yaml.Node {
	propNode := getPropValueByName(node, name)
	if propNode != nil {
		return propNode.Content
	}
	return nil
}

// DSL method to execute validations in order and break early as soon as the first one fails
// This is very useful if a certain validation cannot be executed without the previous ones succeeding.
// For example: matching vs a regExp should not be performed for a property that is not a string.
func sequence(
	checks ...YamlCheck) YamlCheck {

	return sequenceInternal(false, checks...)
}

// DSL method to execute validations in order and break early as soon as the first one fails
// This is very useful if a certain validation cannot be executed without the previous ones succeeding.
// For example: matching vs a regExp should not be performed for a property that is not a string.
func sequenceFailFast(
	checks ...YamlCheck) YamlCheck {

	return sequenceInternal(true, checks...)
}

func sequenceInternal(failfast bool,
	checks ...YamlCheck) YamlCheck {

	return func(yNode, yParentNode *yaml.Node, path []string) YamlValidationIssues {
		var issues YamlValidationIssues

		for _, check := range checks {
			newIssues := check(yNode, yParentNode, path)
			// Only perform the next validation, if the previous one succeeded
			if len(newIssues) > 0 {
				issues = append(issues, newIssues...)
				if failfast {
					break
				}
			}
		}

		return issues
	}
}

// DSL method to iterate over a YAML array items
func forEach(checks ...YamlCheck) YamlCheck {

	return func(yPropNode, yParentNode *yaml.Node, path []string) YamlValidationIssues {
		var issues YamlValidationIssues

		if yPropNode == nil {
			return issues
		}

		arrSize := len(yPropNode.Content)

		validation := sequence(checks...)

		for i := 0; i < arrSize; i++ {
			child := yPropNode.Content[i]
			elemErrors := validation(child, yPropNode, append(path, fmt.Sprintf("[%d]", i)))
			issues = append(issues, elemErrors...)
		}

		return issues
	}
}

// DSL method to iterate over a YAML map values
func forEachProperty(checks ...YamlCheck) YamlCheck {

	return func(yPropNode, yParentNode *yaml.Node, path []string) YamlValidationIssues {
		var issues YamlValidationIssues

		if yPropNode == nil {
			return issues
		}

		validation := sequence(checks...)

		// Properties are listed in the Content of node as slice of key, value, key, value,...
		isKeyNode := true
		var key string
		for _, child := range yPropNode.Content {
			if isKeyNode {
				// the current node is the key of the next node
				key = child.Value
			} else {
				// current is the value => run the validation
				elemErrors := validation(child, yPropNode, append(path, key))
				issues = append(issues, elemErrors...)
			}
			// key->value, value->key
			isKeyNode = !isKeyNode
		}

		return issues
	}
}

// DSL method to ensure a property exists.
// Note that this has no context, the property being checked is provided externally
// via the "property" DSL method.
func required() YamlCheck {
	return func(yNode, yParentNode *yaml.Node, path []string) YamlValidationIssues {
		if yNode == nil {
			return []YamlValidationIssue{
				{
					Msg: fmt.Sprintf(`missing the "%s" required property in the %s .yaml node`,
						last(path), buildPathString(dropRight(path))),
					Line:   yParentNode.Line,
					Column: yParentNode.Column,
				},
			}
		}

		return []YamlValidationIssue{}
	}
}

// DSL method to ensure a property doesn't exist.
// Useful for properties that exist in the structure but should not exist in the yaml, for example
// in extensions.
// Note that this has no context, the property being checked is provided externally
// via the "property" DSL method.
func doesNotExist() YamlCheck {
	return func(yNode, yParentNode *yaml.Node, path []string) YamlValidationIssues {
		if yNode != nil {
			return []YamlValidationIssue{
				{
					Msg: fmt.Sprintf(propertyExistsErrorMsg,
						last(path), buildPathString(dropRight(path))),
					Line:   yNode.Line,
					Column: yNode.Column,
				},
			}
		}

		return []YamlValidationIssue{}
	}
}

// DSL method that will only perform validations if the property exists
// Useful to avoid executing validations on none mandatory properties which are not present.
func optional(checks ...YamlCheck) YamlCheck {
	return func(yNode, yParentNode *yaml.Node, path []string) YamlValidationIssues {
		var issues YamlValidationIssues

		// If an optional property is not found
		// no sense in executing the validations.
		if yNode == nil {
			return issues
		}

		for _, check := range checks {
			newIssues := check(yNode, yParentNode, path)
			issues = append(issues, newIssues...)
		}

		return issues
	}
}

func typeIsNotMapArray() YamlCheck {
	return func(yNode, yParentNode *yaml.Node, path []string) YamlValidationIssues {
		if yNode == nil {
			return []YamlValidationIssue{}
		}

		if yNode.Kind == yaml.SequenceNode || yNode.Kind == yaml.MappingNode {
			return []YamlValidationIssue{
				{
					Msg:    fmt.Sprintf(`the "%s" property must be a string`, buildPathString(path)),
					Line:   yNode.Line,
					Column: yNode.Column,
				},
			}
		}

		return []YamlValidationIssue{}
	}
}

func typeIsArray() YamlCheck {
	return func(yNode, yParentNode *yaml.Node, path []string) YamlValidationIssues {

		if yNode != nil {
			if yNode.Kind != yaml.SequenceNode {
				return []YamlValidationIssue{
					{
						Msg:    fmt.Sprintf(`the "%s" property must be an array`, buildPathString(path)),
						Line:   yNode.Line,
						Column: yNode.Column,
					},
				}
			}
		}

		return []YamlValidationIssue{}
	}
}

func typeIsMap() YamlCheck {
	return func(yNode, yParentNode *yaml.Node, path []string) YamlValidationIssues {
		if yNode != nil {
			if yNode.Kind != yaml.MappingNode {
				return []YamlValidationIssue{
					{
						Msg:    fmt.Sprintf(`the "%s" property must be a map`, buildPathString(path)),
						Line:   yNode.Line,
						Column: yNode.Column,
					},
				}
			}
		}

		return []YamlValidationIssue{}
	}
}

func typeIsBoolean() YamlCheck {
	return func(yNode, yParentNode *yaml.Node, path []string) YamlValidationIssues {
		if yNode != nil {
			if yNode.Tag != "!!bool" {
				return []YamlValidationIssue{

					{Msg: fmt.Sprintf(`the "%s" property must be a boolean`, buildPathString(path)),
						Line:   yNode.Line,
						Column: yNode.Column,
					},
				}
			}
		}

		return []YamlValidationIssue{}
	}
}

func matchesRegExp(pattern string) YamlCheck {
	regExp, _ := regexp.Compile(pattern)

	return func(yNode, yParentNode *yaml.Node, path []string) YamlValidationIssues {
		if yNode == nil {
			return []YamlValidationIssue{}
		}

		strValue := yNode.Value

		if !regExp.MatchString(strValue) {
			return []YamlValidationIssue{
				{
					Msg: fmt.Sprintf(`the "%s" value of the "%s" property does not match the "%s" pattern`,
						strValue, buildPathString(path), pattern),
					Line:   yNode.Line,
					Column: yNode.Column,
				},
			}
		}

		return []YamlValidationIssue{}
	}
}

// Validates that value matches to one of defined enums values
func matchesEnumValues(enumValues []string) YamlCheck {
	expectedSubset := ""
	i := 0
	for _, enumValue := range enumValues {
		i++
		if i > 4 {
			break
		}
		if i > 1 {
			expectedSubset = expectedSubset + ","
		}
		expectedSubset = expectedSubset + enumValue
	}

	return func(yNode, yParentNode *yaml.Node, path []string) YamlValidationIssues {
		if yNode == nil {
			return []YamlValidationIssue{}
		}

		value := yNode.Value
		found := false
		for _, enumValue := range enumValues {
			if enumValue == value {
				found = true
				break
			}
		}
		if !found {
			return []YamlValidationIssue{
				{
					Msg: fmt.Sprintf(
						`the "%s" value of the "%s" enum property is invalid; expected one of the following: %s`,
						value, buildPathString(path), expectedSubset),
					Line:   yNode.Line,
					Column: yNode.Column,
				},
			}
		}

		return []YamlValidationIssue{}
	}
}

func prettifyPath(path string) string {
	wrongIdxSyntax, _ := regexp.Compile(`\.\[`)

	return wrongIdxSyntax.ReplaceAllString(path, "[")
}

func buildPathString(path []string) string {
	if len(path) == 0 {
		return "root"
	}

	if len(path) == 1 {
		return buildPathString(append([]string{"root"}, path...))
	}
	pathStr := strings.Join(path, ".")

	prettyPathStr := prettifyPath(pathStr)

	return prettyPathStr
}

func last(sl []string) string {
	return sl[len(sl)-1]
}

func dropRight(sl []string) []string {
	return sl[:len(sl)-1]
}

func getLiteralStringValue(y *simpleyaml.Yaml) string {
	strVal, strErr := y.String()

	if strErr == nil {
		return strVal
	}

	boolVal, boolErr := y.Bool()
	if boolErr == nil {
		return fmt.Sprintf("%t", boolVal)
	}

	IntVal, IntErr := y.Int()
	if IntErr == nil {
		return fmt.Sprintf("%d", IntVal)
	}

	FloatVal, FloatErr := y.Float()
	if FloatErr == nil {
		return fmt.Sprintf("%g", FloatVal)
	}

	return ""
}

func getContentNode(yamlContent []byte) (*yaml.Node, error) {

	dec := yaml.NewDecoder(bytes.NewReader(yamlContent))
	dec.KnownFields(false)

	var document yaml.Node
	err := dec.Decode(&document)
	if err != nil {
		return &document, err
	}

	var contentNode yaml.Node
	if document.Content != nil {
		contentNode = *document.Content[0]
	}
	return &contentNode, nil
}

// runSchemaValidations - Given a YAML text and a set of validations will execute them and will return relevant issue slice
func runSchemaValidations(mtaNode *yaml.Node, validations ...YamlCheck) []YamlValidationIssue {
	var issues []YamlValidationIssue

	for _, validation := range validations {
		issues = append(issues, validation(mtaNode, nil, []string{})...)
	}

	return issues
}
