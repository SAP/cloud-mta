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
		It("Create MTA", func() {
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

		It("Create MTA with wrong json format", func() {
			wrongJSON := "{Name:fff"
			mtaPath := getTestPath("result", "temp.mta.yaml")
			Ω(CreateMta(mtaPath, wrongJSON)).ShouldNot(Succeed())
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
			Ω(targetFilePath).Should(BeAnExistingFile())
			yamlData, err := ioutil.ReadFile(targetFilePath)
			Ω(err).Should(Succeed())
			oOutput, err := Unmarshal(yamlData)
			Ω(err).Should(Succeed())
			Ω(reflect.DeepEqual(oMtaInput, *oOutput)).Should(BeTrue())
		})

		It("Copy file with non existing path", func() {
			sourceFilePath := "c:/temp/test1"
			targetFilePath := "c:/temp/test2"
			Ω(CopyFile(sourceFilePath, targetFilePath)).ShouldNot(Succeed())
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

	var _ = Describe("addModule", func() {
		It("Add module", func() {
			oModule := Module{
				Name: "testModule",
				Type: "testType",
				Path: "test",
			}

			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData))).Should(Succeed())

			jsonModuleData, err := json.Marshal(oModule)
			Ω(err).Should(Succeed())
			Ω(AddModule(mtaPath, string(jsonModuleData))).Should(Succeed())

			oMtaInput.Modules = append(oMtaInput.Modules, &oModule)

			yamlData, err := ioutil.ReadFile(mtaPath)
			Ω(err).Should(Succeed())
			oMtaOutput, err := Unmarshal(yamlData)
			Ω(err).Should(Succeed())
			Ω(reflect.DeepEqual(oMtaInput, *oMtaOutput)).Should(BeTrue())
		})

		It("Add Module to non existing mta.yaml file", func() {
			json := "{name:fff}"
			mtaPath := getTestPath("result", "mta.yaml")
			Ω(AddModule(mtaPath, json)).ShouldNot(Succeed())
		})

		It("Add Module to wrong mta.yaml format", func() {
			wrongJSON := "{TEST:fff}"
			oModule := Module{
				Name: "testModule",
				Type: "testType",
				Path: "test",
			}

			mtaPath := getTestPath("result", "mta.yaml")
			Ω(CreateMta(mtaPath, wrongJSON)).Should(Succeed())

			jsonModuleData, err := json.Marshal(oModule)
			Ω(err).Should(Succeed())
			Ω(AddModule(mtaPath, string(jsonModuleData))).ShouldNot(Succeed())
		})

		It("Add Module with wrong json format", func() {
			wrongJSON := "{name:fff"

			mtaPath := getTestPath("result", "temp.mta.yaml")
			jsonRootData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData))).Should(Succeed())

			Ω(AddModule(mtaPath, wrongJSON)).ShouldNot(Succeed())
		})
	})

	var _ = Describe("addResource", func() {
		It("Add resource", func() {
			oResource := Resource{
				Name: "testResource",
				Type: "testType",
			}

			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData))).Should(Succeed())

			jsonResourceData, err := json.Marshal(oResource)
			Ω(err).Should(Succeed())
			Ω(AddResource(mtaPath, string(jsonResourceData))).Should(Succeed())

			oMtaInput.Resources = append(oMtaInput.Resources, &oResource)

			yamlData, err := ioutil.ReadFile(mtaPath)
			Ω(err).Should(Succeed())
			oMtaOutput, err := Unmarshal(yamlData)
			Ω(err).Should(Succeed())
			Ω(reflect.DeepEqual(oMtaInput, *oMtaOutput)).Should(BeTrue())
		})

		It("Add Resource to non existing mta.yaml file", func() {
			json := "{name:fff}"
			mtaPath := getTestPath("result", "mta.yaml")
			Ω(AddResource(mtaPath, json)).ShouldNot(Succeed())
		})

		It("Add Resource to wrong mta.yaml format", func() {
			wrongJSON := "{TEST:fff}"
			oResource := Resource{
				Name: "testResource",
				Type: "testType",
			}

			mtaPath := getTestPath("result", "mta.yaml")
			Ω(CreateMta(mtaPath, wrongJSON)).Should(Succeed())

			jsonResourceData, err := json.Marshal(oResource)
			Ω(err).Should(Succeed())
			Ω(AddResource(mtaPath, string(jsonResourceData))).ShouldNot(Succeed())
		})

		It("Add Resource with wrong json format", func() {
			wrongJSON := "{name:fff"

			mtaPath := getTestPath("result", "temp.mta.yaml")
			jsonRootData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData))).Should(Succeed())

			Ω(AddResource(mtaPath, wrongJSON)).ShouldNot(Succeed())
		})
	})
})
