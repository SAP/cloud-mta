package validate

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/SAP/cloud-mta/mta"
)

var _ = Describe("checkBuilderSchema", func() {
	It("Sanity", func() {
		mtaContent := []byte(`
ID: mtahtml5
_schema-version: '2.1'
version: 0.0.1

modules:
 - name: ui5app1
   type: html5
   build-parameters:
     builder: custom
     commands: command

 - name: ui5app2
   type: html5
   build-parameters:
     builder: 
       - a

 - name: ui5app3
   type: html5
   build-parameters:
     builder: custom
     commands: 
       - command1
       -
         - command2

 - name: ui5app4
   type: html5
   build-parameters:
     builder: custom
     commands: 
       - command1
`)
		mta, _ := mta.Unmarshal(mtaContent)
		node, _ := getContentNode(mtaContent)
		issues := checkBuilderSchema(mta, node, "")
		Ω(len(issues)).Should(Equal(3))
		Ω(issues[0].Msg).Should(Equal(`the "commands" property is defined incorrectly; the property must be a sequence of strings`))
		Ω(issues[0].Line).Should(Equal(11))
		Ω(issues[1].Msg).Should(Equal(`the "builder" property is defined incorrectly; the property must be a string`))
		Ω(issues[1].Line).Should(Equal(17))
		Ω(issues[2].Msg).Should(Equal(`the "commands" property is defined incorrectly; the property must be a sequence of strings`))
		Ω(issues[2].Line).Should(Equal(24))
	})

})
