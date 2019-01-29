package validate

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/SAP/cloud-mta/mta"
)

var _ = Describe("validateNamesUniqueness", func() {
	It("Sanity", func() {
		mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

modules:
 - name: ui5app
   type: html5
   provides:
   - name: test

 - name: ui5app2
   type: html5

resources:
 - name: test
   parameters:
      path: ./xs-security.json
      service-plan: application
   type: com.company.xs.uaa

 - name: dest_mtahtml5
   parameters:
      service-plan: lite
      service: destination
   type: org.cloudfoundry.managed-service
`)
		mta, _ := mta.Unmarshal(mtaContent)
		issues := runSemanticValidations(mta, getTestPath("testproject") )
		Ω(len(issues)).Should(Equal(2))
		Ω(issues[0].Msg).Should(Equal("validation of the modules failed because the ui5app2 path of the ui5app2 module does not exist"))
		Ω(issues[1].Msg).Should(Equal("the test resource is not unique because the provided with the same name defined"))
	})
})