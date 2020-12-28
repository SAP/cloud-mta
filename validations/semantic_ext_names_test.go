package validate

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/SAP/cloud-mta/mta"
)

var _ = Describe("checkSingleExtendNames", func() {
	It("doesn't return issues when all names are different", func() {
		mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

modules:
 - name: module1
   type: html5
   provides:
   - name: provides1
   - name: provides2
   requires:
   - name: requires1
   - name: requires2
   hooks:
   - name: hook1
     requires:
     - name: requires3
     - name: requires4
   - name: hook2
     requires:
     - name: requires5
     - name: requires6

 - name: module2
   type: html5
   provides:
   - name: provides7
   - name: provides8
   requires:
   - name: requires9
   - name: requires10
   hooks:
   - name: hook3
     requires:
     - name: requires11
     - name: requires12
   - name: hook4
     requires:
     - name: requires13
     - name: requires14

resources:
 - name: resource1
   parameters:
      path: ./xs-security.json
      service-plan: application
   type: com.company.xs.uaa
   requires:
   - name: requires15
   - name: requires16

 - name: resource2
   parameters:
      path: ./xs-security.json
      service-plan: application
   type: com.company.xs.uaa
   requires:
   - name: requires17
   - name: requires18
`)
		mta, _ := mta.UnmarshalExt(mtaContent)
		node, _ := getContentNode(mtaContent)
		issues, _ := checkSingleExtendNames(mta, node, "", true)
		Ω(len(issues)).Should(Equal(0))
	})
	It("doesn't return issues when names are different in their own sections", func() {
		mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

modules:
 - name: a1
   type: html5
   provides:
   - name: a1
   - name: a2
   requires:
   - name: a1
   - name: a2
   hooks:
   - name: a1
     requires:
     - name: a1
     - name: a2
   - name: a2
     requires:
     - name: a1
     - name: a2

 - name: a2
   type: html5
   provides:
   - name: a1
   - name: a2
   requires:
   - name: a1
   - name: a2
   hooks:
   - name: a1
     requires:
     - name: a1
     - name: a2
   - name: a2
     requires:
     - name: a1
     - name: a2

resources:
 - name: a1
   parameters:
      path: ./xs-security.json
      service-plan: application
   type: com.company.xs.uaa
   requires:
   - name: a1
   - name: a2

 - name: a2
   parameters:
      path: ./xs-security.json
      service-plan: application
   type: com.company.xs.uaa
   requires:
   - name: a1
   - name: a2
`)
		mta, _ := mta.UnmarshalExt(mtaContent)
		node, _ := getContentNode(mtaContent)
		issues, _ := checkSingleExtendNames(mta, node, "", true)
		Ω(len(issues)).Should(Equal(0))
	})

	It("returns issue when module is extended twice", func() {
		mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

modules:
 - name: m1
   parameters:
     a: 1

 - name: m1
   properties:
     b: 2
`)
		mta, _ := mta.UnmarshalExt(mtaContent)
		node, _ := getContentNode(mtaContent)
		issues, _ := checkSingleExtendNames(mta, node, "", true)
		Ω(issues).Should(ConsistOf(YamlValidationIssue{getDuplicateExtendsErrorMsg("m1", moduleEntityKind, 7), 11, 10}))
	})
	It("returns issue when module provides is extended twice", func() {
		mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

modules:
 - name: m1
   provides:
     - name: p1
     - name: p1
`)
		mta, _ := mta.UnmarshalExt(mtaContent)
		node, _ := getContentNode(mtaContent)
		issues, _ := checkSingleExtendNames(mta, node, "", true)
		Ω(issues).Should(ConsistOf(YamlValidationIssue{getDuplicateExtendsErrorMsg("p1", providedPropEntityKind, 9), 10, 14}))
	})
	It("returns issue when module requires is extended twice", func() {
		mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

modules:
 - name: m1
   requires:
     - name: r1
     - name: r1
`)
		mta, _ := mta.UnmarshalExt(mtaContent)
		node, _ := getContentNode(mtaContent)
		issues, _ := checkSingleExtendNames(mta, node, "", true)
		Ω(issues).Should(ConsistOf(YamlValidationIssue{getDuplicateExtendsErrorMsg("r1", requiresPropEntityKind, 9), 10, 14}))
	})
	It("returns issue when module hook is extended twice", func() {
		mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

modules:
 - name: m1
   hooks:
   - name: h1
   - name: h1
`)
		mta, _ := mta.UnmarshalExt(mtaContent)
		node, _ := getContentNode(mtaContent)
		issues, _ := checkSingleExtendNames(mta, node, "", true)
		Ω(issues).Should(ConsistOf(YamlValidationIssue{getDuplicateExtendsErrorMsg("h1", hookPropEntityKind, 9), 10, 12}))
	})
	It("returns issue when module hook requires is extended twice", func() {
		mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

modules:
 - name: m1
   hooks:
     - name: p1
       requires:
       - name: r1
       - name: r1
`)
		mta, _ := mta.UnmarshalExt(mtaContent)
		node, _ := getContentNode(mtaContent)
		issues, _ := checkSingleExtendNames(mta, node, "", true)
		Ω(issues).Should(ConsistOf(YamlValidationIssue{getDuplicateExtendsErrorMsg("r1", requiresPropEntityKind, 11), 12, 16}))
	})
	It("returns issue when resource is extended twice", func() {
		mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

resources:
 - name: r1
   parameters:
     a: 1

 - name: r1
   properties:
     b: 2
`)
		mta, _ := mta.UnmarshalExt(mtaContent)
		node, _ := getContentNode(mtaContent)
		issues, _ := checkSingleExtendNames(mta, node, "", true)
		Ω(issues).Should(ConsistOf(YamlValidationIssue{getDuplicateExtendsErrorMsg("r1", resourceEntityKind, 7), 11, 10}))
	})
	It("returns issue when resource requires is extended twice", func() {
		mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

resources:
 - name: r1
   requires:
     - name: req1
     - name: req1
`)
		mta, _ := mta.UnmarshalExt(mtaContent)
		node, _ := getContentNode(mtaContent)
		issues, _ := checkSingleExtendNames(mta, node, "", true)
		Ω(issues).Should(ConsistOf(YamlValidationIssue{getDuplicateExtendsErrorMsg("req1", requiresPropEntityKind, 9), 10, 14}))
	})

	It("returns the expected issues when several entities are extended twice", func() {
		mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

modules:
 - name: m1
   provides:
     - name: p1
     - name: p1
 - name: m1
   requires:
     - name: r1
     - name: r1

resources:
  - name: r1
  - name: r1
`)
		mta, _ := mta.UnmarshalExt(mtaContent)
		node, _ := getContentNode(mtaContent)
		issues, _ := checkSingleExtendNames(mta, node, "", true)
		Ω(issues).Should(ConsistOf(
			YamlValidationIssue{getDuplicateExtendsErrorMsg("m1", moduleEntityKind, 7), 11, 10},
			YamlValidationIssue{getDuplicateExtendsErrorMsg("p1", providedPropEntityKind, 9), 10, 14},
			YamlValidationIssue{getDuplicateExtendsErrorMsg("r1", requiresPropEntityKind, 13), 14, 14},
			YamlValidationIssue{getDuplicateExtendsErrorMsg("r1", resourceEntityKind, 17), 18, 11},
		))
	})
})

func getDuplicateExtendsErrorMsg(name, entityType string, line int) string {
	return fmt.Sprintf(nameAlreadyExtendedMsg, name, entityType, "another", entityType, line)
}
