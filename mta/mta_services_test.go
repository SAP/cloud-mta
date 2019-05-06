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

	schemaVersion := "1.1"
	oMtaInput := MTA{
		ID:            "test",
		Version:       "1.2",
		SchemaVersion: &schemaVersion,
		Description:   "test mta creation",
	}

	AfterEach(func() {
		os.RemoveAll(getTestPath("result"))
	})
	var _ = Describe("CreateMta", func() {
		It("Sanity", func() {
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

	var _ = Describe("CopyFile", func() {
		It("Copy file content", func() {
			jsonData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			sourceFilePath := getTestPath("result", "temp.mta.yaml")
			targetFilePath := getTestPath("result", "temp2.mta.yaml")
			Ω(CreateMta(sourceFilePath, string(jsonData))).Should(Succeed())
			Ω(CopyFile(sourceFilePath, targetFilePath)).Should(Succeed())
			yamlData, err := ioutil.ReadFile(targetFilePath)
			Ω(err).Should(Succeed())
			oOutput, err := Unmarshal(yamlData)
			Ω(err).Should(Succeed())
			Ω(reflect.DeepEqual(oMtaInput, *oOutput)).Should(BeTrue())
		})
	})

	var _ = Describe("deleteFile", func() {
		It("Delete file", func() {
			jsonData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			mtaPath := getTestPath("result", "temp.mta.yaml")
			Ω(CreateMta(mtaPath, string(jsonData))).Should(Succeed())
			Ω(mtaPath).Should(BeAnExistingFile())
			Ω(DeleteFile(mtaPath)).Should(Succeed())
			Ω(mtaPath).ShouldNot(BeAnExistingFile())
		})
	})
})
