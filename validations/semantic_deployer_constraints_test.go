package validate

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/SAP/cloud-mta/mta"
)

var _ = Describe("SemanticDeployerConstraints", func() {
	It("Sanity", func() {
		mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

parameters:
  deploy_mode: html5-repo

modules:
 - name: ui5app1
   type: html5
      
 - name: ui5app2
   type: html5
   build-parameters:
     builder: grunt

 - name: ui5app3
   type: html5
   build-parameters:
     supported-platforms: []

 - name: ui5app4
   type: html5
   build-parameters:
     build-result: dist

 - name: ui5app5
   type: html5
   build-parameters:
     supported-platforms: []
     build-result: dist

 - name: ui5app6
   type: html5
   build-parameters:
     supported-platforms: [cf, neo]
     build-result: dist1
`)
		mta, _ := mta.Unmarshal(mtaContent)
		node, _ := getContentNode(mtaContent)
		errors, warn := checkDeployerConstraints(mta, node, "", true)
		Ω(len(errors)).Should(Equal(4))
		Ω(len(warn)).Should(Equal(0))
		checkDeployerConstrError(errors[0], missingConfigsMsg, 10, "ui5app1")
		checkDeployerConstrError(errors[1], missingConfigsMsg, 13, "ui5app2")
		checkDeployerConstrError(errors[2], missingConfigMsg, 18, "ui5app3", buildResultYamlField)
		checkDeployerConstrError(errors[3], missingConfigMsg, 23, "ui5app4", supportedPlatformsYamlField)
	})
})

func checkDeployerConstrError(issue YamlValidationIssue, message string, line int, params ...interface{}) {
	params = append(params, missingConfigDocLink)
	Ω(issue.Msg).Should(Equal(fmt.Sprintf(message, params...)))
	Ω(issue.Line).Should(Equal(line))

}
