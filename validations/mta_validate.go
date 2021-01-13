package validate

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/SAP/cloud-mta/internal/fs"
	"github.com/SAP/cloud-mta/mta"
)

const (
	SeverityError   = "error"
	SeverityWarning = "warning"
)

type ValidationResult map[string][]FileValidationIssue
type FileValidationIssue struct {
	// Severity - the severity of the message. Possible values: "error", "warning".
	Severity string `json:"severity"`
	// Message - the validation message
	Message string `json:"message"`
	// Line - the line number of the issue
	Line int `json:"line"`
	// Column - the column number of the issue
	Column int `json:"column"`
}

// Validate validates an mta.yaml file and a list of mta extension files, and returns the issues for each file.
func Validate(mtaPath string, extensions []string) ValidationResult {
	allIssues := make(ValidationResult)
	// Assuming the project path is the folder of the mta.yaml
	projectPath := filepath.Dir(mtaPath)
	mtaYamlFileName := filepath.Base(mtaPath)
	// Paths validation is excluded because it doesn't prevent building the MTA (the paths can be created during the build)
	errorIssues, warningIssues, e := validateMtaYaml(projectPath, mtaYamlFileName, true, true, true, pathsValidation)
	if e != nil {
		errorIssues = appendIssue(errorIssues, e.Error(), 0, 0)
	}
	allIssues[mtaPath] = createFileIssues(warningIssues, errorIssues)

	for _, extPath := range extensions {
		errorIssues, warningIssues, e = validateMtaext(projectPath, extPath, true, true, true, "")
		if e != nil {
			errorIssues = appendIssue(errorIssues, e.Error(), 0, 0)
		}
		allIssues[extPath] = createFileIssues(warningIssues, errorIssues)
	}

	_, _, e = mta.GetMtaFromFile(mtaPath, extensions, true)
	if e != nil {
		// Ignore errors which are not on a specific extension (if they are on the mta.yaml we already got them earlier)
		// and parse errors from extensions (we already got them earlier too)
		if extErr, ok := e.(*mta.ExtensionError); ok && !extErr.IsParseError {
			allIssues[extErr.FileName] = append(allIssues[extErr.FileName], FileValidationIssue{SeverityError, e.Error(), 0, 0})
		}
	}

	return allIssues
}

func createFileIssues(warningIssues YamlValidationIssues, errorIssues YamlValidationIssues) []FileValidationIssue {
	allIssues := make([]FileValidationIssue, 0)
	for _, issue := range errorIssues {
		allIssues = append(allIssues, FileValidationIssue{SeverityError, issue.Msg, issue.Line, issue.Column})
	}
	for _, issue := range warningIssues {
		allIssues = append(allIssues, FileValidationIssue{SeverityWarning, issue.Msg, issue.Line, issue.Column})
	}

	return allIssues
}

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
	errIssues, warnIssues, err := validateMtaYaml(projectPath, mtaFilename, validateSchema, validateSemantic, strict, exclude)
	if err != nil {
		return "", err
	}
	if len(errIssues) > 0 {
		return warnIssues.String(), errors.Errorf(`the %q file is not valid: `+"\n%v",
			filepath.Join(projectPath, mtaFilename), errIssues.String())
	}
	return warnIssues.String(), nil
}

func validateMtaYaml(projectPath, mtaFilename string, validateSchema, validateSemantic, strict bool,
	exclude string) (errIssues YamlValidationIssues, warnIssues YamlValidationIssues, err error) {
	if validateSemantic || validateSchema {
		mtaPath := filepath.Join(projectPath, mtaFilename)
		// ParseFile contains MTA yaml content.
		yamlContent, e := fs.ReadFile(mtaPath)
		if e != nil {
			return nil, nil, errors.Wrapf(e, `could not read the %q file; the validation failed`, mtaPath)
		}

		// Validates MTA content.
		errIssues, warnIssues := validate(yamlContent, projectPath, validateSchema, validateSemantic, strict, exclude)
		errIssues.Sort()
		warnIssues.Sort()
		return errIssues, warnIssues, nil
	}

	return nil, nil, nil
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

			// Add converted issue to the issues list. We only have the line number here.
			issues = appendIssue(issues, e, line, 0)
		}
	}
	return issues
}
