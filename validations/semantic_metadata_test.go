package validate

import (
	"fmt"
	"github.com/SAP/cloud-mta/mta"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
				YamlValidationIssue{fmt.Sprintf(unknownNameInMetadataMsg, "b", "parameter"), 11},
				YamlValidationIssue{fmt.Sprintf(emptyRequiredFieldMsg, "memory", "parameter"), 18},
				YamlValidationIssue{fmt.Sprintf(unknownNameInMetadataMsg, "x", "property"), 27},
				YamlValidationIssue{fmt.Sprintf(unknownNameInMetadataMsg, "b", "parameter"), 34},
				YamlValidationIssue{fmt.Sprintf(unknownNameInMetadataMsg, "a", "parameter"), 41},
				YamlValidationIssue{fmt.Sprintf(unknownNameInMetadataMsg, "b", "property"), 51},
				YamlValidationIssue{fmt.Sprintf(unknownNameInMetadataMsg, "a", "parameter"), 56},
				YamlValidationIssue{fmt.Sprintf(unknownNameInMetadataMsg, "b", "property"), 64},
				YamlValidationIssue{fmt.Sprintf(unknownNameInMetadataMsg, "m", "parameter"), 73},
				YamlValidationIssue{fmt.Sprintf(emptyRequiredFieldMsg, "b", "property"), 80},
				YamlValidationIssue{fmt.Sprintf(unknownNameInMetadataMsg, "b", "property"), 93},
				YamlValidationIssue{fmt.Sprintf(unknownNameInMetadataMsg, "a", "parameter"), 98},
			))
		})

	})
})
