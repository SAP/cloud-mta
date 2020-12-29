package validate

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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
       group: a

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
        optional: null

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
       list: "abc"

 - name: res4
   type: custom
   requires:
   - name: req2
     properties-metadata:
     group:
   - name: req3
     properties-metadata:
     list:
`)
			mta, err := mta.Unmarshal(mtaContent)
			Ω(err).Should(Succeed())
			node, _ := getContentNode(mtaContent)
			errors, warn := checkParamsAndPropertiesMetadata(mta, node, "", true)
			Ω(len(warn)).Should(Equal(0))
			Ω(errors).Should(ConsistOf(
				YamlValidationIssue{fmt.Sprintf(unknownNameInMetadataMsg, "b", "parameter"), 11, 3},
				YamlValidationIssue{fmt.Sprintf(emptyRequiredFieldMsg, "memory", "parameter"), 18, 6},
				YamlValidationIssue{fmt.Sprintf(unknownNameInMetadataMsg, "x", "property"), 27, 6},
				YamlValidationIssue{fmt.Sprintf(unknownNameInMetadataMsg, "b", "parameter"), 34, 10},
				YamlValidationIssue{fmt.Sprintf(unknownNameInMetadataMsg, "a", "parameter"), 41, 14},
				YamlValidationIssue{fmt.Sprintf(unknownNameInMetadataMsg, "b", "property"), 51, 10},
				YamlValidationIssue{fmt.Sprintf(unknownNameInMetadataMsg, "a", "parameter"), 56, 10},
				YamlValidationIssue{propertiesMetadataWithListOrGroupMsg, 50, 8},
				YamlValidationIssue{fmt.Sprintf(unknownNameInMetadataMsg, "b", "property"), 65, 10},
				YamlValidationIssue{fmt.Sprintf(unknownNameInMetadataMsg, "m", "parameter"), 74, 6},
				YamlValidationIssue{fmt.Sprintf(emptyRequiredFieldMsg, "b", "property"), 81, 6},
				YamlValidationIssue{fmt.Sprintf(unknownNameInMetadataMsg, "b", "property"), 94, 10},
				YamlValidationIssue{fmt.Sprintf(unknownNameInMetadataMsg, "a", "parameter"), 99, 10},
				YamlValidationIssue{propertiesMetadataWithListOrGroupMsg, 93, 8},
				YamlValidationIssue{propertiesMetadataWithListOrGroupMsg, 107, 6},
				YamlValidationIssue{propertiesMetadataWithListOrGroupMsg, 110, 6},
			))
		})

	})
})

var _ = Describe("isPropertyOverWritable", func() {
	It("default value", func() {
		Ω(isPropertyOverWritable(nil)).Should(BeTrue())
	})
	It("non default", func() {
		falseValue := false
		Ω(isPropertyOverWritable(&falseValue)).Should(BeFalse())
	})
})

var _ = Describe("isPropertyOptional", func() {
	It("default value", func() {
		Ω(isPropertyOptional(nil)).Should(BeFalse())
	})
	It("non default", func() {
		trueValue := true
		Ω(isPropertyOverWritable(&trueValue)).Should(BeTrue())
	})
})
