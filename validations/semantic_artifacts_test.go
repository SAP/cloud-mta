package validate

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/SAP/cloud-mta/mta"
)

var _ = Describe("ValidateModulesPaths", func() {
	var mtaContent = []byte(`
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
   path: ui5app2
   parameters:
      disk-quota: 256M
      memory: 256M
   requires:
   - name: uaa_mtahtml5
   - name: dest_mtahtml5

 - name: ui5app3
   type: html5
   build-parameters:
     no-source: true

 - name: ui5app4
   type: html5
   build-parameters:
     no-source: abc

 - name: ui5app5
   type: html5
   path: 
   build-parameters:
     no-source: 
        - a
        - b

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

	It("ifModulePathExists", func() {
		wd, _ := os.Getwd()
		mta, err := mta.Unmarshal(mtaContent)
		Ω(err).Should(Succeed())
		root, _ := getContentNode(mtaContent)
		issues, _ := ifModulePathExists(mta, root, filepath.Join(wd, "testdata", "testproject"), true)
		Ω(len(issues)).Should(Equal(1))
		Ω(issues[0].Msg).Should(
			Equal(`the "ui5app2" path of the "ui5app2" module does not exist`))
	})

	It("ifModulePathEmpty", func() {
		wd, _ := os.Getwd()
		mta, _ := mta.Unmarshal(mtaContent)
		root, _ := getContentNode(mtaContent)
		issues, _ := ifModulePathEmpty(mta, root, filepath.Join(wd, "testdata", "testproject"), true)
		Ω(len(issues)).Should(Equal(2))
		Ω(issues[0].Msg).Should(
			Equal(`the path of the "ui5app4" module is not defined`))
		Ω(issues[0].Line).Should(Equal(32))
		Ω(issues[1].Msg).Should(
			Equal(`the path of the "ui5app5" module is empty`))
		Ω(issues[1].Line).Should(Equal(39))
	})

	It("ifNoSourceParamBool", func() {
		wd, _ := os.Getwd()
		mta, _ := mta.Unmarshal(mtaContent)
		root, _ := getContentNode(mtaContent)
		issues, _ := ifNoSourceParamBool(mta, root, filepath.Join(wd, "testdata", "testproject"), true)
		Ω(len(issues)).Should(Equal(2))
		Ω(issues[0].Msg).Should(
			Equal(`the "no-source" build parameter must be a boolean`))
		Ω(issues[1].Msg).Should(
			Equal(`the "no-source" build parameter must be a boolean`))
	})
})
