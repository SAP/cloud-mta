package mta

import (
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"reflect"
)

var _ = Describe("MtaServices", func() {
	AfterEach(func() {
		os.RemoveAll(getTestPath("result"))
	})
	var _ = Describe("CreateMta", func() {
		It("Sanity", func() {
			schemaVersion := "1.1"
			oMtaInput := MTA{
				ID:            "test",
				Version:       "1.2",
				SchemaVersion: &schemaVersion,
				Description:   "test mta creation",
			}
			jsonData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			mtaPath := getTestPath("result", "temp.mta.yaml")
			Ω(CreateMta(mtaPath, string(jsonData))).Should(Succeed())
			yamlData, err := ioutil.ReadFile(mtaPath)
			Ω(err).Should(Succeed())
			oMtaOutput, err := Unmarshal(yamlData)
			Ω(err).Should(Succeed())
			Ω(reflect.DeepEqual(oMtaInput, *oMtaOutput)).Should(BeTrue())
		})
	})
})
