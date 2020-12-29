package validate

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/SAP/cloud-mta/mta"
)

var _ = Describe("checkMetadataSchema", func() {
	It("gives an error for datatype field in parameters-metadata", func() {
		mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

parameters:
  a: 1
parameters-metadata:
  a:
    datatype: str
    optional: false

modules:
- name: ui5app1
  type: html5
  parameters:
    memory:
    env: abc
  parameters-metadata:
    memory:
      overwritable: false
      datatype: str
    env:
      optional: true
- name: ui5app2
  type: html5
  hooks:
  - name: h1
    parameters:
      b: 1
    parameters-metadata:
      b:
        datatype: true
    requires:
    - name: hr1
      parameters:
        a:
      parameters-metadata:
        a:
          datatype: true
- name: ui5app3
  type: html5
  requires:
  - name: r1
    properties:
      a: 1
    properties-metadata:
      b:
        datatype: int
    parameters:
      b: 1
    parameters-metadata:
      a:
        datatype: true
resources:
- name: res1
  type: custom
  parameters:
    m:
  parameters-metadata:
    m:
      overwritable: false
      datatype: float
- name: res3
  type: custom
  requires:
  - name: req1
    properties:
      a: 1
    properties-metadata:
      a:
        datatype: bool
    parameters-metadata:
      b:
        datatype: bool
`)
		mta, err := mta.Unmarshal(mtaContent)
		Ω(err).Should(Succeed())
		node, _ := getContentNode(mtaContent)
		issues := checkMetadataSchema(mta, node, "")

		datatypeNotAllowedForParametersMetadata := fmt.Sprintf(propertyExistsErrorMsg, datatypeYamlField, parametersMetadataField)
		Ω(issues).Should(ConsistOf(
			YamlValidationIssue{datatypeNotAllowedForParametersMetadata, 10, 5},
			YamlValidationIssue{datatypeNotAllowedForParametersMetadata, 22, 7},
			YamlValidationIssue{datatypeNotAllowedForParametersMetadata, 33, 9},
			YamlValidationIssue{datatypeNotAllowedForParametersMetadata, 40, 11},
			YamlValidationIssue{datatypeNotAllowedForParametersMetadata, 54, 9},
			YamlValidationIssue{datatypeNotAllowedForParametersMetadata, 63, 7},
			YamlValidationIssue{datatypeNotAllowedForParametersMetadata, 75, 9},
		))
	})
})
