package validate

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/SAP/cloud-mta/mta"
)

var _ = Describe("checkSingleExtendNames", func() {
	It("Sanity", func() {
		mtaContent := []byte(`
ID: mtahtml5ext
_schema-version: '2.1'
version: 0.0.1
extends: mtahtml5

modules:
 - name: ui5app
   parameters:
     A: 1
   provides:
   - name: test

 - name: ui5app

resources:
 - name: test
   parameters:
      path: ./xs-security2.json

 - name: test
   parameters:
      service-plan: premium
`)
		mtaExt, e := mta.UnmarshalExt(mtaContent)
		Ω(e).Should(Succeed())
		root, _ := getContentNode(mtaContent)
		issues, _ := runExtSemanticValidations(mtaExt, root, getTestPath("testproject"), "", true)
		Ω(issues).Should(ConsistOf(
			YamlValidationIssue{fmt.Sprintf(nameAlreadyExtendedMsg, "ui5app", "module", "another", "module", 8), 14, 10},
			YamlValidationIssue{fmt.Sprintf(nameAlreadyExtendedMsg, "test", "resource", "another", "resource", 17), 21, 10},
		))
		issues, _ = runExtSemanticValidations(mtaExt, root, getTestPath("testproject"), "names", true)
		Ω(len(issues)).Should(Equal(0))
	})

})
