package validate

import (
	"fmt"
	"github.com/SAP/cloud-mta/mta"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SemanticDeprecatedOpts", func() {
	It("Sanity", func() {
		mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

modules:
 - name: ui5app1
   type: html5
   build-parameters:
     builder: npm
     npm-opts: abc
      
 - name: ui5app2
   type: html5
   build-parameters:
     builder: grunt
     grunt-opts: 
       - opt1
       - opt2

 - name: ui5app3
   type: html5
   build-parameters:
     builder: mvn
     maven-opts: 1
`)
		mta, _ := mta.Unmarshal(mtaContent)
		node, _ := getContentNode(mtaContent)
		errors, warn := checkDeprecatedOpts(mta, node, "", true)
		Ω(len(errors)).Should(Equal(3))
		Ω(len(warn)).Should(Equal(0))
		checkDeprecatedOptError(errors[0], npmOptsYamlField, 11)
		checkDeprecatedOptError(errors[1], gruntOptsYamlField, 17)
		checkDeprecatedOptError(errors[2], mavenOptsYamlField, 25)
	})
})

func checkDeprecatedOptError(issue YamlValidationIssue, optName string, line int) {
	Ω(issue.Msg).Should(Equal(fmt.Sprintf(deprecatedOptMsg, optName, customBuilderDocLink)))
	Ω(issue.Line).Should(Equal(line))
}
