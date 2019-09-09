package validate

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/SAP/cloud-mta/mta"
)

// GetValidationMode converts validation mode flags to validation process flags.
func GetValidationMode(validationFlag string) (bool, bool, error) {
	switch validationFlag {
	case "schema":
		return true, false, nil
	case "semantic":
		return true, true, nil
	case "":
		return true, true, nil
	}
	return false, false,
		fmt.Errorf(`the "%s" validation mode is incorrect; expected one of the following: schema, semantic`,
			validationFlag)
}

// MtaYaml validates an MTA.yaml file.
func MtaYaml(projectPath, mtaFilename string,
	validateSchema, validateSemantic, strict bool, exclude string) (warning string, err error) {
	if validateSemantic || validateSchema {

		mtaPath := filepath.Join(projectPath, mtaFilename)
		// ParseFile contains MTA yaml content.
		yamlContent, e := readFile(mtaPath)

		if e != nil {
			return "", errors.Wrapf(e, `could not read the "%v" file; the validation failed`, mtaPath)
		}
		s := string(yamlContent)
		s = strings.Replace(s, "\r\n", "\r", -1)
		yamlContent = []byte(s)
		// Validates MTA content.
		errIssues, warnIssues := validate(yamlContent, projectPath,
			validateSchema, validateSemantic, strict, exclude)
		errIssues.Sort()
		warnIssues.Sort()
		if len(errIssues) > 0 {
			return warnIssues.String(), errors.Errorf(`the "%v" file is not valid: `+"\n%v",
				mtaPath, errIssues.String())
		}
		return warnIssues.String(), nil
	}

	return "", nil
}

func readFile(file string) ([]byte, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, `failed to read the "%v" file`, file)
	}
	s := string(content)
	s = strings.Replace(s, "\r\n", "\r", -1)
	content = []byte(s)
	return content, nil
}

// validate - validates the MTA descriptor
func validate(yamlContent []byte, projectPath string,
	validateSchema, validateSemantic, strict bool, exclude string) (errIssues YamlValidationIssues, warnIssues YamlValidationIssues) {

	mtaStr, err := mta.Unmarshal(yamlContent)

	if strict && err != nil {
		errIssues = append(errIssues, convertError(err)...)
	} else if err != nil {
		warnIssues = append(warnIssues, convertError(err)...)
	}

	mtaNode, err := getContentNode(yamlContent)
	if err != nil {
		errIssues = convertError(err)
	}

	if validateSchema {
		validations, schemaValidationLog := buildValidationsFromSchemaText(schemaDef)
		if len(schemaValidationLog) > 0 {
			errIssues = append(errIssues, schemaValidationLog...)
			return errIssues, warnIssues
		}
		errIssues = append(errIssues, runSchemaValidations(mtaNode, validations...)...)

		issues := checkBuilderSchema(mtaStr, mtaNode, "")
		if strict {
			errIssues = append(errIssues, issues...)
		} else {
			warnIssues = append(warnIssues, issues...)
		}

		issues = checkMetadataSchema(mtaStr, mtaNode, "")
		if strict {
			errIssues = append(errIssues, issues...)
		} else {
			warnIssues = append(warnIssues, issues...)
		}
	}

	if validateSemantic {
		errs, warns := runSemanticValidations(mtaStr, mtaNode, projectPath, exclude, strict)
		errIssues = append(errIssues, errs...)
		warnIssues = append(warnIssues, warns...)
	}
	return errIssues, warnIssues
}

// convertError - converts unmarshalling errors to the YamlValidationIssue format
// extracting line number to issue Line property
func convertError(err error) []YamlValidationIssue {
	var issues []YamlValidationIssue

	if err != nil {
		reLineNumber := regexp.MustCompile("(.)*line [0-9]+: ")
		reNumber := regexp.MustCompile("[0-9]+")

		errType, ok := err.(*yaml.TypeError)
		var errorsList []string
		if !ok {
			// single error can be received in errString format
			errorsList = append(errorsList, err.Error())
		} else {
			// multiple errors come in TypeError format
			errorsList = errType.Errors
		}
		for _, e := range errorsList {
			// find substring ... line <number>:
			lineNumberStr := reLineNumber.FindString(e)
			// remove this substring from error message
			e = strings.Replace(e, lineNumberStr, "", 1)
			// extract line number from the substring
			lineStr := reNumber.FindString(lineNumberStr)
			line, err := strconv.Atoi(lineStr)
			if err != nil {
				line = 1
			}

			// add converted issue to the issues list
			issues = appendIssue(issues, e, line)
		}
	}
	return issues
}
