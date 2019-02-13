package validate

import (
	"fmt"

	"github.com/SAP/cloud-mta/mta"
)

// ifRequiredDefined - validates that required property sets are defined in modules, provided sections or resources
func ifRequiredDefined(mta *mta.MTA, source string) []YamlValidationIssue {
	var issues []YamlValidationIssue

	// init set of all provided property sets
	providedSet := make(map[string]interface{})

	for _, module := range mta.Modules {
		// add module to provided property sets
		providedSet[module.Name] = nil
		// add all property sets provided by module
		for _, provided := range module.Provides {
			providedSet[provided.Name] = nil
		}
	}

	// add resources to provided property sets
	for _, resource := range mta.Resources {
		providedSet[resource.Name] = nil
	}

	for _, module := range mta.Modules {
		// check that each required property set was provided in mta.yaml
		for _, requires := range module.Requires {
			if _, contains := providedSet[requires.Name]; !contains {
				issues = appendIssue(issues,
					fmt.Sprintf(`the "%s" property set required by the "%s" module is not defined`,
						requires.Name, module.Name))
			}
		}
	}
	return issues
}
