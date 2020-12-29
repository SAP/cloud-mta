package validate

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("MTA Schema Validations", func() {
	It("Main", func() {

		var data = []byte(`
type: map
mapping:
   _schema-version:  {required: true}
   ID: {required: true, pattern: '/^[A-Za-z0-9_\-\.]+$/'}
   description:
   version: {required: true, pattern: '/^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(-(0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(\.(0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*)?(\+[0-9a-zA-Z-]+(\.[0-9a-zA-Z-]+)*)?$/'}
   provider:
   copyright:
   modules:
      type: seq
      sequence:
       - type: map
         mapping:
            name: {required: true, unique: true, pattern: '/^[A-Za-z0-9_\-\.]+$/'}
            type: {required: true}
            description:
            path:

`)
		schemaValidations, _ := buildValidationsFromSchemaText(data)
		input := []byte(`
_schema-version: "2.0.0"
ID: mta_proj
version: 1.0.0

modules:
  - name: webide-feature-management
    type: html5
    path: public
    provides:
      - name: webide-feature-management
        public: true
    build-parameters:
      builder: npm
      build-result: "dist"
      timeout: 15m
      requires:
        - name: webide-feature-management-client
          artifacts: ["dist/*"]
          target-path: "dist_client_tmp"

  - name: webide-feature-management-client
    typoe: html5
    path: client
    build-parameters:
      builder: npm
      supported-platforms: []

`)
		node, _ := getContentNode(input)
		validateIssues := runSchemaValidations(node, schemaValidations...)
		expectSingleValidationError(validateIssues, `missing the "type" required property in the modules[1] .yaml node`, 22, 5)
	})
})
