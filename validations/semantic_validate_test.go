package validate

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
		issues := runSemanticValidations(mtaContent, getTestPath("testproject"))
		Ω(len(issues)).Should(Equal(2))
		Ω(issues[0].Msg).Should(
			Equal(`the "ui5app2" path of the "ui5app2" module does not exist`))
		Ω(issues[1].Msg).
			Should(Equal(`the "test" resource name is not unique; a provided service was found with the same name`))
	})

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
   parameters:
      service-plan: lite
      service: destination
      service: destination1
`)
		issues := runSemanticValidations(mtaContent, getTestPath("testproject"))
		Ω(len(issues)).Should(Equal(1))
		fmt.Println(issues[0].Msg)
		Ω(issues[0].Msg).Should(
			ContainSubstring(`line 14: key "service" already set in map`))
	})
})
