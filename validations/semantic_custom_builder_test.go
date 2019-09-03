package validate

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/SAP/cloud-mta/mta"
)

var _ = Describe("SemanticCustomBuilder", func() {
	var _ = Describe("checkBuilderSchema", func() {
		It("Sanity", func() {
			mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

build-parameters:
  before-all:
    - builder: npm
      commands: 
        -  command1
    - builder: custom
      commands: 
        -  command1
    - builder: custom

modules:
 - name: ui5app1
   type: html5
   build-parameters:
     builder: custom
      

 - name: ui5app2
   type: html5
   build-parameters:
     builder: npm
     commands: 
       - command1

 - name: ui5app2
   type: html5
   build-parameters:
     builder: custom
     commands: 
       - command1

 - name: ui5app4
   type: html5
   build-parameters:
     builder: npm

 - name: ui5app4
   type: html5
   build-parameters:
     builder: 
       - npm
`)
			mta, _ := mta.Unmarshal(mtaContent)
			node, _ := getContentNode(mtaContent)
			errors, warn := checkBuildersSemantic(mta, node, "", true)
			Ω(len(errors)).Should(Equal(4))
			Ω(len(warn)).Should(Equal(0))
			Ω(errors[0].Msg).Should(Equal(`the "commands" property is not supported by the "npm" builder`))
			Ω(errors[0].Line).Should(Equal(10))
			Ω(errors[1].Msg).Should(Equal(`the "commands" property is missing in the "custom" builder`))
			Ω(errors[1].Line).Should(Equal(14))
			Ω(errors[2].Msg).Should(Equal(`the "commands" property is missing in the "custom" builder`))
			Ω(errors[2].Line).Should(Equal(20))
			Ω(errors[3].Msg).Should(Equal(`the "commands" property is not supported by the "npm" builder`))
			Ω(errors[3].Line).Should(Equal(28))
		})

	})

})
