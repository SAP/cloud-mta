package validate

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/SAP/cloud-mta/mta"
)

const (
	filenameValidation = "filename"

	mtaextExtension = ".mtaext"

	badExtensionErrorMsg = `MTA extension descriptor file name must have the "mtaext" file extension`
)

// Mtaext validates an MTA extension file.
func Mtaext(projectPath, extFilename string,
	validateSchema, validateSemantic, strict bool, exclude string) (warning string, err error) {
	if validateSemantic || validateSchema {
		var errIssues, warnIssues YamlValidationIssues

		extPath := filepath.Join(projectPath, extFilename)
		// ParseFile contains MTA yaml content.
		yamlContent, e := readFile(extPath)

		if e != nil {
			return "", errors.Wrapf(e, `could not read the "%v" file; the validation failed`, extPath)
		}
		s := string(yamlContent)
		s = strings.Replace(s, "\r\n", "\r", -1)
		yamlContent = []byte(s)

		// Validates MTA content.
		contentErrIssues, contentWarnIssues := validateExt(yamlContent, projectPath, extFilename,
			validateSchema, validateSemantic, strict, exclude)
		errIssues = append(errIssues, contentErrIssues...)
		warnIssues = append(warnIssues, contentWarnIssues...)
		if len(errIssues) > 0 {
			return warnIssues.String(), errors.Errorf(`the "%v" file is not valid: `+"\n%v",
				extPath, errIssues.String())
		}
		return warnIssues.String(), nil
	}

	return "", nil
}

func validateExtFileName(name string, strict bool) (errIssues YamlValidationIssues, warnIssues YamlValidationIssues) {
	var issues YamlValidationIssues

	if filepath.Ext(name) != mtaextExtension {
		issues = append(issues, YamlValidationIssue{Msg: badExtensionErrorMsg, Line: 0})
	}

	if strict {
		return issues, nil
	}
	return nil, issues
}

// validateExt validates the MTA extension descriptor
func validateExt(yamlContent []byte, projectPath string, extFileName string,
	validateSchema, validateSemantic, strict bool, exclude string) (errIssues YamlValidationIssues, warnIssues YamlValidationIssues) {

	// This is a special case semantic validation, on the file name and not its content
	if validateSemantic && !strings.Contains(exclude, filenameValidation) {
		errIssues, warnIssues = validateExtFileName(extFileName, strict)
	}

	mtaExt, err := mta.UnmarshalExt(yamlContent)

	if strict && err != nil {
		errIssues = append(errIssues, convertError(err)...)
	} else if err != nil {
		warnIssues = append(warnIssues, convertError(err)...)
	}

	extNode, err := getContentNode(yamlContent)
	if err != nil {
		errIssues = convertError(err)
	}

	if validateSchema {
		validations, schemaValidationLog := buildValidationsFromSchemaText(extSchemaDef)
		if len(schemaValidationLog) > 0 {
			errIssues = append(errIssues, schemaValidationLog...)
			return errIssues, warnIssues
		}
		errIssues = append(errIssues, runSchemaValidations(extNode, validations...)...)
	}

	if validateSemantic {
		errs, warns := runExtSemanticValidations(mtaExt, extNode, projectPath, exclude, strict)
		errIssues = append(errIssues, errs...)
		warnIssues = append(warnIssues, warns...)
	}
	return errIssues, warnIssues
}
