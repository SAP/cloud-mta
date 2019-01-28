package validate

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/SAP/cloud-mta/mta"
)

var _ = Describe("ValidateYamlProject", func() {
	It("Sanity", func() {
		wd, _ := os.Getwd()
		mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

modules:
 - name: ui5app
   type: html5
   path: ui5app
   parameters:
      disk-quota: 256M
      memory: 256M
   requires:
    - name: uaa_mtahtml5
    - name: dest_mtahtml5


 - name: ui5app2
   type: html5
   parameters:
      disk-quota: 256M
      memory: 256M
   requires:
   - name: uaa_mtahtml5
   - name: dest_mtahtml5

resources:
 - name: uaa_mtahtml5
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
		issues := validateYamlProject(mta, filepath.Join(wd, "testdata", "testproject"))
		Ω(issues[0].Msg).Should(Equal("validation of the modules failed because the ui5app2 path of the ui5app2 module does not exist"))
	})
})

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
 - name: uaa_mtahtml5
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
		issues := validateNamesUniqueness(mta, "")
		Ω(len(issues)).Should(Equal(0))
	})

	It("the same module name", func() {
		mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

modules:
 - name: ui5app
   type: html5
   provides:
   - name: test

 - name: ui5app
   type: html5
`)
		mta, _ := mta.Unmarshal(mtaContent)
		issues := validateNamesUniqueness(mta, "")
		Ω(issues[0].Msg).Should(Equal("the ui5app module is not unique because the module with the same name defined"))
	})
	It("module and provides have the same name", func() {
		mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

modules:
 - name: ui5app
   type: html5
   provides:
   - name: ui5app2

 - name: ui5app2
   type: html5
`)
		mta, _ := mta.Unmarshal(mtaContent)
		issues := validateNamesUniqueness(mta, "")
		Ω(issues[0].Msg).Should(Equal("the ui5app2 module is not unique because the provided with the same name defined"))
	})
	It("resource and provides have the same name", func() {
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
`)
		mta, _ := mta.Unmarshal(mtaContent)
		issues := validateNamesUniqueness(mta, "")
		Ω(issues[0].Msg).Should(Equal("the test resource is not unique because the provided with the same name defined"))
	})
})
