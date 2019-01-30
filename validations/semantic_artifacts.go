package validate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SAP/cloud-mta/mta"
)

// validateModulesPaths - validates the existence of modules paths used in the MTA descriptor
func validateModulesPaths(mta *mta.MTA, source string) []YamlValidationIssue {
	var issues []YamlValidationIssue
	for _, module := range mta.Modules {
		modulePath := module.Path
		// "path" property not defined -> use module name as a path
		if modulePath == "" {
			modulePath = module.Name
		}
		// build full path
		fullPath := filepath.Join(source, modulePath)
		// check existence of file/folder
		_, err := os.Stat(fullPath)
		if err != nil {
			// path not exists -> add an issue
			issues = appendIssue(issues, fmt.Sprintf("validation of the modules failed because the %s path of the %s module does not exist",
				modulePath, module.Name))
		}
	}

	return issues
}
