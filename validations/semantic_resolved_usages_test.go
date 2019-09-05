package validate

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/SAP/cloud-mta/mta"
)

var _ = Describe("SemanticResolvedUsages", func() {
	It("Sanity", func() {
		mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

modules:
 - name: ui5app
   type: html5
   provides:
   - name: test

 - name: ui5app2
   type: html5
   requires:
   - name: test
   - name: test1
   - name: uaa_mtahtml5
   - name: ui5app

resources:
 - name: uaa_mtahtml5
   parameters:
      path: ./xs-security.json
      service-plan: application
   type: com.company.xs.uaa
   requires:
   - name: test
   - name: test1

 - name: dest_mtahtml5
   parameters:
      service-plan: lite
      service: destination
   type: org.cloudfoundry.managed-service
`)
		mta, _ := mta.Unmarshal(mtaContent)
		node, _ := getContentNode(mtaContent)
		issues, _ := ifRequiredDefined(mta, node, "", true)
		Ω(len(issues)).Should(Equal(2))
		Ω(issues[0].Msg).Should(Equal(`the "test1" property set required by the "ui5app2" module is not defined`))
		Ω(issues[0].Line).Should(Equal(16))
		Ω(issues[1].Msg).Should(Equal(`the "test1" property set required by the "uaa_mtahtml5" resource is not defined`))
		Ω(issues[1].Line).Should(Equal(28))
	})

	It("Sanity - resource of type configuration", func() {
		mtaContent := []byte(`
ID: test.consumer
_schema-version: 3.3.0
version: 0.0.0

parameters:
  env: dev

resources:
- name: external-configurations
  type: configuration
  parameters: 
    provider-id: test.provider:configuration_${env}
    target: 
      org: ${org}
      space: ${space}
    version: ">=0.0.0"

modules:
  - name: staticapp
    type: staticfile
    path: Staticfile.zip
    parameters:
      no-route: true
      no-start: true
      the-config-in-module-parameter: ~{external-configurations/a-config-property}.inModuleParameter
    properties: 
      property-directly-referencing-the-config: ~{external-configurations/a-config-property}.directly
      via-module-parameter-value: ${the-config-in-module-parameter}
    requires:
      - name: test1
      - name: external-configurations
`)
		mta, _ := mta.Unmarshal(mtaContent)
		node, _ := getContentNode(mtaContent)
		issues, _ := ifRequiredDefined(mta, node, "", true)
		Ω(len(issues)).Should(Equal(1))
		Ω(issues[0].Msg).Should(Equal(`the "test1" property set required by the "staticapp" module is not defined`))
		Ω(issues[0].Line).Should(Equal(31))
	})

	It("check required properties (placeholders usage)", func() {
		mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

modules:
 - name: pricing-ui
   type: javascript.nodejs
   requires:
   - name: price_opt 
     properties:
       conn_string: "~{protocol}://~{uri}/odata/" 
       conn_string1: "~{protocol1}://~{uri}/odata/"
       x: 
         xa: "~{protocol}://~{uri}/odata/"
   - name: unknown
     properties:
       conn_string2: "~{protocol}://~{uri}/odata/"

   properties: 
     conn_string3: "~{protocol}://~{uri}/odata/" 
     a: "~{price_opt/protocol}://~{price_opt/uri}/odata/"
     b: "~{price_opt/protocol1}://~{price_opt/uri}/odata/"
     c: "~{price_opt1/protocol}://~{price_opt/uri}/odata/"
     complex: 
       a: "~{price_opt1/protocol}://~{price_opt/uri}/odata/"
       b: "~{price_opt/address}://~{price_opt/address1}/odata/"
   parameters: 
     aaa: "~{protocol}://~{uri}/odata/" 

 - name: pricing-backend
   type: html5
   provides:
   - name: price_opt
     properties:
       protocol: http
       uri: myhost.mydomain
       uri1: ~{aaaa}
       uri2: ~{pricing-ui/a}
       address: 
         protocolX: http
         uriX: myhost.mydomain

 - name1: unnamed
   type: html5
   properties: 
     conn_string: "~{price_opt/protocol}://~{price_opt/uri}/odata/" 
`)
		mta, err := mta.Unmarshal(mtaContent)
		Ω(err).Should(HaveOccurred())
		node, _ := getContentNode(mtaContent)
		issues, _ := ifRequiredDefined(mta, node, "", true)
		Ω(len(issues)).Should(Equal(13))
	})
})
