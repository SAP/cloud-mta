package validate

import (
	"fmt"

	"github.com/SAP/cloud-mta/mta"
)

// validateRequested - validates that required named are defined in modules provided sections or in resources
func validateRequested(mta *mta.MTA, source string) []YamlValidationIssue {
	var issues []YamlValidationIssue
	providedSet := make(map[string]interface{})

	// fill set of all provided names by modules
	for _, module := range mta.Modules {
		for _, provided := range module.Provides {
			providedSet[provided.Name] = nil
		}
	}

	// fill set of all provided names by resources
	for _, resource := range mta.Resources {
		providedSet[resource.Name] = nil
	}

	for _, module := range mta.Modules {
		for _, requires := range module.Requires {
			if _, contains := providedSet[requires.Name]; !contains {
				issues = appendIssue(issues,
					fmt.Sprintf(`the "%s" name required by the "%s" module is not provided`,
					requires.Name, module.Name))
			}
		}
	}
	return issues
}
