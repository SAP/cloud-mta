package validate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SAP/cloud-mta/mta"
)

type yamlProjectCheck func(mta *mta.MTA, path string) []YamlValidationIssue

// validateModules - Validate the MTA file
func validateModules(mta *mta.MTA, projectPath string) []YamlValidationIssue {
	//noinspection GoPreferNilSlice
	issues := []YamlValidationIssue{}
	for _, module := range mta.Modules {
		modulePath := module.Path
		if modulePath == "" {
			modulePath = module.Name
		}
		fullPath := filepath.Join(projectPath, modulePath)
		_, err := os.Stat(fullPath)
		if err != nil {
			issues = append(issues, []YamlValidationIssue{
				{
					Msg: fmt.Sprintf("validation of the modules failed because the %s path of the %s module does not exist",
						modulePath, module.Name),
				}}...)
		}
	}

	return issues
}

func validateNamesUniqueness(mta *mta.MTA, path string) []YamlValidationIssue {
	issues := []YamlValidationIssue{}
	names := make(map[string]string)
	for _, module := range mta.Modules {
		issues = validateName(names, module.Name, "module", issues)
		for _, provide := range module.Provides {
			issues = validateName(names, provide.Name, "provided", issues)
		}
	}
	for _, resource := range mta.Resources {
		issues = validateName(names, resource.Name, "resource", issues)
	}
	return issues
}

func validateName(names map[string]string, name string,
	objectName string, issues []YamlValidationIssue) []YamlValidationIssue {
	result := issues
	prevObjectName, ok := names[name]
	if ok {
		result = append(result, []YamlValidationIssue{
			{
				Msg: fmt.Sprintf("the %s %s is not unique because the %s with the same name defined",
					name, objectName, prevObjectName),
			}}...)
	} else {
		names[name] = objectName
	}
	return result
}

// validateYamlProject - Validate the MTA file
func validateYamlProject(mta *mta.MTA, path string) []YamlValidationIssue {
	validations := []yamlProjectCheck{validateModules, validateNamesUniqueness}
	//noinspection GoPreferNilSlice
	issues := []YamlValidationIssue{}
	for _, validation := range validations {
		validationIssues := validation(mta, path)
		issues = append(issues, validationIssues...)

	}
	return issues
}
