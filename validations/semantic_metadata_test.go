package validate

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/SAP/cloud-mta/mta"
)

var _ = Describe("metadata semantic validations", func() {
	var _ = Describe("checkParamsAndPropertiesMetadata", func() {
		It("Sanity", func() {
			mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

parameters:
  a: 1
parameters-metadata:
  a:
    overwritable: true
  b:
    optional: false

modules:
 - name: ui5app1
   type: html5
   parameters:
     memory:
   parameters-metadata:
     memory:
        overwritable: false
        optional: false

 - name: ui5app2
   type: html5
   properties-metadata:
     x:
        optional: false
   hooks:
     - name: h1
       parameters:
         a: 1
       parameters-metadata:
         b:
           overwritable: true
       requires:
         - name: hr1
           parameters:
             b:
           parameters-metadata:
             a:
               overwritable: true

 - name: ui5app3
   type: html5
   requires:
     - name: r1
       properties:
         a: 1
       properties-metadata:
         b:
           overwritable: true
       parameters:
         b: 1
       parameters-metadata:
         a:
           overwritable: true

 - name: ui5app4
   type: html5
   provides:
     - name: p1
       properties-metadata:
         b:
           overwritable: true

resources:
 - name: res1
   type: custom
   parameters:
     a:
   parameters-metadata:
     m:
        overwritable: false
        optional: false

 - name: res2
   type: custom
   properties:
     b:
   properties-metadata:
     b:
        overwritable: false
        optional: false

 - name: res3
   type: custom
   requires:
     - name: req1
       properties:
         a: 1
       properties-metadata:
         b:
           overwritable: true
       parameters:
         b: 1
       parameters-metadata:
         a:
           overwritable: true
`)
			mta, _ := mta.Unmarshal(mtaContent)
			node, _ := getContentNode(mtaContent)
			errors, warn := checkParamsAndPropertiesMetadata(mta, node, "", true)
			Ω(len(warn)).Should(Equal(0))
			Ω(errors).Should(ConsistOf(
				matchValidationIssue(11, fmt.Sprintf(unknownNameInMetadataMsg, "b", "parameter")),
				matchValidationIssue(18, fmt.Sprintf(emptyRequiredFieldMsg, "memory", "parameter")),
				matchValidationIssue(27, fmt.Sprintf(unknownNameInMetadataMsg, "x", "property")),
				matchValidationIssue(34, fmt.Sprintf(unknownNameInMetadataMsg, "b", "parameter")),
				matchValidationIssue(41, fmt.Sprintf(unknownNameInMetadataMsg, "a", "parameter")),
				matchValidationIssue(51, fmt.Sprintf(unknownNameInMetadataMsg, "b", "property")),
				matchValidationIssue(56, fmt.Sprintf(unknownNameInMetadataMsg, "a", "parameter")),
				matchValidationIssue(64, fmt.Sprintf(unknownNameInMetadataMsg, "b", "property")),
				matchValidationIssue(73, fmt.Sprintf(unknownNameInMetadataMsg, "m", "parameter")),
				matchValidationIssue(80, fmt.Sprintf(emptyRequiredFieldMsg, "b", "property")),
				matchValidationIssue(93, fmt.Sprintf(unknownNameInMetadataMsg, "b", "property")),
				matchValidationIssue(98, fmt.Sprintf(unknownNameInMetadataMsg, "a", "parameter")),
			))
		})

	})
})

func matchValidationIssue(line int, msg string) types.GomegaMatcher {
	return SatisfyAll(
		WithTransform(func(v YamlValidationIssue) int {
			return v.Line
		}, Equal(line)),
		WithTransform(func(v YamlValidationIssue) string {
			return v.Msg
		}, Equal(msg)))
}
