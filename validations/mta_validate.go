package validate

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
)

// GetValidationMode converts validation mode flags to validation process flags.
func GetValidationMode(validationFlag string) (bool, bool, error) {
	switch validationFlag {
	case "schema":
		return true, false, nil
	case "semantic":
		return true, true, nil
	}
	return false, false,
		fmt.Errorf("the %s validation mode is incorrect; expected one of the following: schema, semantic",
			validationFlag)
}

// MtaYaml validates an MTA.yaml file.
func MtaYaml(projectPath, mtaFilename string, validateSchema bool, validateSemantic bool) error {
	if validateSemantic || validateSchema {

		mtaPath := filepath.Join(projectPath, mtaFilename)
		// ParseFile contains MTA yaml content.
		yamlContent, err := ioutil.ReadFile(mtaPath)

		if err != nil {
			return errors.Wrapf(err, "could not read the %v file; the validation failed", mtaPath)
		}
		// Validates MTA content.
		issues := validate(yamlContent, projectPath, validateSchema, validateSemantic)
		if len(issues) > 0 {
			return errors.Errorf("validation of the %v file failed with the following issues: \n%v",
				mtaPath, issues.String())
		}
	}

	return nil
}

// validate - validates the MTA descriptor
func validate(yamlContent []byte, projectPath string, validateSchema bool, validateSemantic bool) YamlValidationIssues {
	var issues []YamlValidationIssue
	if validateSchema {
		validations, schemaValidationLog := buildValidationsFromSchemaText(schemaDef)
		if len(schemaValidationLog) > 0 {
			return schemaValidationLog
		}
		issues = append(issues, runSchemaValidations(yamlContent, validations...)...)

	}
	if validateSemantic {
		issues = append(issues, runSemanticValidations(yamlContent, projectPath)...)
	}
	return issues
}
