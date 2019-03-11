package validate

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/SAP/cloud-mta/mta"
)

var _ = Describe("SemanticResolvedUsages", func() {
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
   requires:
   - name: test
   - name: test1
   - name: uaa_mtahtml5
   - name: ui5app

resources:
 - name: uaa_mtahtml5
   parameters:
      path: ./xs-security.json
      service-plan: application
   type: com.company.xs.uaa
   requires:
   - name: test
   - name: test1

 - name: dest_mtahtml5
   parameters:
      service-plan: lite
      service: destination
   type: org.cloudfoundry.managed-service
`)
		mta, _ := mta.Unmarshal(mtaContent)
		issues := ifRequiredDefined(mta, "")
		Ω(len(issues)).Should(Equal(2))
		Ω(issues[0].Msg).Should(Equal(`the "test1" property set required by the "ui5app2" module is not defined`))
		Ω(issues[1].Msg).Should(Equal(`the "test1" property set required by the "uaa_mtahtml5" resource is not defined`))	})
})
