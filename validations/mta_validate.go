package validate

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/SAP/cloud-mta/mta"
)

// GetValidationMode converts validation mode flags to validation process flags.
func GetValidationMode(validationFlag string) (bool, bool, error) {
	switch validationFlag {
	case "":
		return true, true, nil
	case "schema":
		return true, false, nil
	case "project":
		return false, true, nil
	}
	return false, false,
		fmt.Errorf("the %s validation mode is incorrect; expected one of the following: all, schema, project",
			validationFlag)
}

// MtaYaml validates an MTA.yaml file.
func MtaYaml(projectPath, mtaFilename string, validateSchema bool, validateProject bool) error {
	if validateProject || validateSchema {

		mtaPath := filepath.Join(projectPath, mtaFilename)
		// ParseFile contains MTA yaml content.
		yamlContent, err := ioutil.ReadFile(mtaPath)

		if err != nil {
			return errors.Wrapf(err, "could not read the %v file; the validation failed", mtaPath)
		}
		// Validates MTA content.
		issues, err := validate(yamlContent, projectPath, validateSchema, validateProject)
		if err != nil {
			issues = appendIssue(issues, err.Error())
		}
		if len(issues) > 0 {
			return errors.Errorf("validation of the %v file failed with the following issues: \n%v", mtaPath, issues.String())
		}
	}

	return nil
}

// validate - validates the MTA descriptor
func validate(yamlContent []byte, projectPath string, validateSchema bool, validateProject bool) (YamlValidationIssues, error) {
	var issues []YamlValidationIssue
	if validateSchema {
		validations, schemaValidationLog := BuildValidationsFromSchemaText(schemaDef)
		if len(schemaValidationLog) > 0 {
			return schemaValidationLog, nil
		}
		yamlValidationLog, err := Yaml(yamlContent, validations...)
		if err != nil && len(yamlValidationLog) == 0 {
			yamlValidationLog = appendIssue(yamlValidationLog, "validation failed because: "+err.Error())
		}
		issues = append(issues, yamlValidationLog...)

	}
	if validateProject {
		mtaStr := mta.MTA{}
		err := yaml.Unmarshal(yamlContent, &mtaStr)
		if err != nil {
			return nil, errors.Wrap(err, "validation failed when unmarshalling the MTA file")
		}
		projectIssues := runSemanticValidations(&mtaStr, projectPath)
		issues = append(issues, projectIssues...)
	}
	return issues, nil
}
