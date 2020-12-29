// By naming this file with _test suffix it is not measured
// in the coverage report, although we do end-up with a strange file name...
package validate

import (
	"fmt"

	. "github.com/onsi/gomega"
)

func assertNoValidationErrors(errors []YamlValidationIssue) {
	Ω(len(errors)).Should(Equal(0), fmt.Sprintf("Validation issues detected: %v", errors))
}

func expectSingleValidationError(actual []YamlValidationIssue, expectedMsg string, expectedLine int, expectedColumn int) {
	Ω(actual).Should(ConsistOf(YamlValidationIssue{expectedMsg, expectedLine, expectedColumn}))
}
