package validate

import (
	"fmt"
	"github.com/SAP/cloud-mta/mta"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SemanticDeprecatedOpts", func() {
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

	mtaExtContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
extends: example
version: 0.0.1

modules:
 - name: ui5app1
   build-parameters:
     builder: npm
     npm-opts: abc
      
 - name: ui5app2
   build-parameters:
     builder: grunt
     grunt-opts: 
       - opt1
       - opt2

 - name: ui5app3
   build-parameters:
     builder: mvn
     maven-opts: 1
`)

	It("checkDeprecatedOpts - Sanity", func() {

		mta, err := mta.Unmarshal(mtaContent)
		Ω(err).Should(Succeed())
		node, err := getContentNode(mtaContent)
		Ω(err).Should(Succeed())
		errors, warn := checkDeprecatedOpts(mta, node, "", true)
		Ω(len(errors)).Should(Equal(3))
		Ω(len(warn)).Should(Equal(0))
		checkDeprecatedOptError(errors[0], npmOptsYamlField, 11)
		checkDeprecatedOptError(errors[1], gruntOptsYamlField, 17)
		checkDeprecatedOptError(errors[2], mavenOptsYamlField, 25)
	})

	It("checkExtDeprecatedOpts - Sanity", func() {

		mta, err := mta.UnmarshalExt(mtaExtContent)
		Ω(err).Should(Succeed())
		node, err := getContentNode(mtaExtContent)
		Ω(err).Should(Succeed())
		errors, warn := checkExtDeprecatedOpts(mta, node, "", true)
		Ω(len(errors)).Should(Equal(3))
		Ω(len(warn)).Should(Equal(0))
		checkDeprecatedOptError(errors[0], npmOptsYamlField, 11)
		checkDeprecatedOptError(errors[1], gruntOptsYamlField, 16)
		checkDeprecatedOptError(errors[2], mavenOptsYamlField, 23)
	})

	It("aliases usage", func() {
		mtaContent := []byte(`
_schema1-version: 3.1.0
ID: app
version: 1.0.0

parameters:
  defaults:
    - &build-parameters-app
      npm-opts:
        pre-param: ci

modules:
  - name: mod1
    type: html5
    path: path1
    build-parameters: *build-parameters-app
`)
		mta, _ := mta.Unmarshal(mtaContent)
		node, _ := getContentNode(mtaContent)
		errors, warn := checkDeprecatedOpts(mta, node, "", true)
		Ω(len(errors)).Should(Equal(1))
		Ω(len(warn)).Should(Equal(0))
		checkDeprecatedOptError(errors[0], npmOptsYamlField, 9)
	})
})

func checkDeprecatedOptError(issue YamlValidationIssue, optName string, line int) {
	Ω(issue.Msg).Should(Equal(fmt.Sprintf(deprecatedOptMsg, optName, customBuilderDocLink)))
	Ω(issue.Line).Should(Equal(line))
}
