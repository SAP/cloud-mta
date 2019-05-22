package mta

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getMtaInput() MTA {
	// Return a new object every time so we don't accidentally change it for the other tests
	schemaVersion := "1.1"
	oMtaInput := MTA{
		ID:            "test",
		Version:       "1.2",
		SchemaVersion: &schemaVersion,
		Description:   "test mta creation",
	}
	return oMtaInput
}

var _ = Describe("MtaServices", func() {

	oModule := Module{
		Name: "testModule",
		Type: "testType",
		Path: "test",
	}

	oResource := Resource{
		Name: "testResource",
		Type: "testType",
	}

	AfterEach(func() {
		err := os.RemoveAll(getTestPath("result"))
		Ω(err).Should(Succeed())
	})

	var _ = Describe("CreateMta", func() {
		It("Create MTA", func() {
			jsonData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			mtaPath := getTestPath("result", "temp.mta.yaml")
			Ω(CreateMta(mtaPath, string(jsonData), os.MkdirAll)).Should(Succeed())
			Ω(mtaPath).Should(BeAnExistingFile())
			yamlData, err := ioutil.ReadFile(mtaPath)
			Ω(err).Should(Succeed())
			oMtaOutput, err := Unmarshal(yamlData)
			Ω(err).Should(Succeed())
			Ω(reflect.DeepEqual(getMtaInput(), *oMtaOutput)).Should(BeTrue())
		})

		It("Create MTA with wrong json format", func() {
			wrongJSON := "{Name:fff"
			mtaPath := getTestPath("result", "temp.mta.yaml")
			Ω(CreateMta(mtaPath, wrongJSON, os.MkdirAll)).Should(HaveOccurred())
		})

		It("Create MTA fail to create file", func() {
			jsonData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			mtaPath := getTestPath("result", "temp.mta.yaml")
			Ω(CreateMta(mtaPath, string(jsonData), mkDirsErr)).Should(HaveOccurred())
		})
	})

	var _ = Describe("CopyFile", func() {
		It("Copy file content", func() {
			jsonData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			sourceFilePath := getTestPath("result", "temp.mta.yaml")
			targetFilePath := getTestPath("result", "temp2.mta.yaml")
			Ω(CreateMta(sourceFilePath, string(jsonData), os.MkdirAll)).Should(Succeed())
			Ω(CopyFile(sourceFilePath, targetFilePath, os.Create)).Should(Succeed())
			Ω(targetFilePath).Should(BeAnExistingFile())
			yamlData, err := ioutil.ReadFile(targetFilePath)
			Ω(err).Should(Succeed())
			oOutput, err := Unmarshal(yamlData)
			Ω(err).Should(Succeed())
			Ω(reflect.DeepEqual(getMtaInput(), *oOutput)).Should(BeTrue())
		})

		It("Copy file with non existing path", func() {
			sourceFilePath := "c:/temp/test1"
			targetFilePath := "c:/temp/test2"
			Ω(CopyFile(sourceFilePath, targetFilePath, os.Create)).Should(HaveOccurred())
		})

		It("Copy file fail to create destination file", func() {
			jsonData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			sourceFilePath := getTestPath("result", "temp.mta.yaml")
			targetFilePath := getTestPath("result", "temp2.mta.yaml")
			Ω(CreateMta(sourceFilePath, string(jsonData), os.MkdirAll)).Should(Succeed())
			Ω(CopyFile(sourceFilePath, targetFilePath, createErr)).Should(HaveOccurred())
			Ω(targetFilePath).ShouldNot(BeAnExistingFile())
		})
	})

	var _ = Describe("deleteFile", func() {
		It("Delete file", func() {
			jsonData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			mtaPath := getTestPath("result", "temp.mta.yaml")
			Ω(CreateMta(mtaPath, string(jsonData), os.MkdirAll)).Should(Succeed())
			Ω(mtaPath).Should(BeAnExistingFile())
			Ω(DeleteFile(mtaPath)).Should(Succeed())
			Ω(mtaPath).ShouldNot(BeAnExistingFile())
		})
	})

	var _ = Describe("printResult", func() {
		var printed string
		printer := func(s ...interface{}) (int, error) {
			printed = s[0].(string)
			return 0, nil
		}
		BeforeEach(func() {
			printed = ""
		})

		It("Writes only the hashcode when the result and error are nil", func() {
			err := printResult(nil, 123, nil, printer)
			Ω(err).Should(Succeed())
			Ω(printed).Should(Equal(`{"hashcode":123}`))
		})

		It("Writes error message when the error is not nil", func() {
			err := printResult("123", 123, errors.New("error message"), printer)
			Ω(err).Should(Succeed())
			Ω(printed).Should(Equal(`{"message":"error message"}`))
		})

		It("Writes hashcode and result when the result is sent and there is no error", func() {
			err := printResult("1234", 3, nil, printer)
			Ω(err).Should(Succeed())
			Ω(printed).Should(Equal(`{"result":"1234","hashcode":3}`))
		})

		It("Writes complex result", func() {
			modules := []Module{
				{
					Name: "m1",
					Type: "type1",
				},
				{
					Name: "m2",
					Type: "type2",
				},
			}
			err := printResult(modules, 0, nil, printer)
			Ω(err).Should(Succeed())
			Ω(printed).Should(Equal(`{"result":[{"name":"m1","type":"type1"},{"name":"m2","type":"type2"}],"hashcode":0}`))
		})

		It("Returns print error if print fails", func() {
			printerErr := func(s ...interface{}) (int, error) {
				return 0, errors.New("error in print")
			}
			err := printResult(nil, 1, nil, printerErr)
			Ω(err).Should(MatchError("error in print"))
		})
	})

	var _ = Describe("addModule", func() {
		It("Add module", func() {
			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonModuleData, err := json.Marshal(oModule)
			Ω(err).Should(Succeed())
			Ω(AddModule(mtaPath, string(jsonModuleData), Marshal)).Should(Succeed())

			oMtaInput := getMtaInput()
			oMtaInput.Modules = append(oMtaInput.Modules, &oModule)
			Ω(mtaPath).Should(BeAnExistingFile())
			yamlData, err := ioutil.ReadFile(mtaPath)
			Ω(err).Should(Succeed())
			oMtaOutput, err := Unmarshal(yamlData)
			Ω(err).Should(Succeed())
			Ω(reflect.DeepEqual(oMtaInput, *oMtaOutput)).Should(BeTrue())
		})

		It("Add module to non existing mta.yaml file", func() {
			json := "{name:fff}"
			mtaPath := getTestPath("result", "mta.yaml")
			Ω(AddModule(mtaPath, json, Marshal)).Should(HaveOccurred())
		})

		It("Add module to wrong mta.yaml format", func() {
			wrongJSON := "{TEST:fff}"
			mtaPath := getTestPath("result", "mta.yaml")
			Ω(CreateMta(mtaPath, wrongJSON, os.MkdirAll)).Should(Succeed())

			jsonModuleData, err := json.Marshal(oModule)
			Ω(err).Should(Succeed())
			Ω(AddModule(mtaPath, string(jsonModuleData), Marshal)).Should(HaveOccurred())
		})

		It("Add module with wrong json format", func() {
			wrongJSON := "{name:fff"

			mtaPath := getTestPath("result", "temp.mta.yaml")
			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			Ω(AddModule(mtaPath, wrongJSON, Marshal)).Should(HaveOccurred())
		})

		It("Add module fails to marshal", func() {
			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonModuleData, err := json.Marshal(oModule)
			Ω(err).Should(Succeed())
			Ω(AddModule(mtaPath, string(jsonModuleData), marshalErr)).Should(HaveOccurred())
		})
	})

	var _ = Describe("updateModule", func() {
		It("fails when mta.yaml doesn't exist", func() {
			json := "{name:fff}"
			mtaPath := getTestPath("result", "mta.yaml")
			err := UpdateModule(mtaPath, json, Marshal)
			Ω(err).Should(HaveOccurred())
		})

		It("fails when mta has wrong format", func() {
			wrongJSON := "{TEST:fff}"

			mtaPath := getTestPath("result", "mta.yaml")
			Ω(CreateMta(mtaPath, wrongJSON, os.MkdirAll)).Should(Succeed())

			jsonModuleData, err := json.Marshal(oModule)
			Ω(err).Should(Succeed())
			err = UpdateModule(mtaPath, string(jsonModuleData), Marshal)
			Ω(err).Should(MatchError(MatchRegexp("yaml: unmarshal errors")))
		})

		It("fails when input is bad json format", func() {
			wrongJSON := "{name:fff"

			mtaPath := getTestPath("result", "temp.mta.yaml")
			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			err = UpdateModule(mtaPath, wrongJSON, Marshal)
			Ω(err).Should(MatchError(MatchRegexp("line 1: did not find expected")))
		})

		It("fails when module with this name doesn't exist", func() {
			oOriginalModule := Module{
				Name: "testModule",
				Type: "testType",
				Path: "test",
			}

			oUpdatedModule := Module{
				Name: "testModule2",
				Type: "testType2",
				Path: "test2",
			}

			mtaPath := getTestPath("result", "temp.mta.yaml")

			oMtaInput := getMtaInput()
			oMtaInput.Modules = append(oMtaInput.Modules, &oOriginalModule)
			jsonRootData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonModuleData, err := json.Marshal(oUpdatedModule)
			Ω(err).Should(Succeed())
			err = UpdateModule(mtaPath, string(jsonModuleData), Marshal)
			Ω(err).Should(MatchError("module with name testModule2 does not exist"))
		})

		It("fails when marshal to mta.yaml fails", func() {
			oOriginalModule := Module{
				Name: "testModule",
				Type: "testType",
				Path: "test",
			}

			oUpdatedModule := Module{
				Name: "testModule",
				Type: "testType2",
				Path: "test2",
			}

			mtaPath := getTestPath("result", "temp.mta.yaml")

			oMtaInput := getMtaInput()
			oMtaInput.Modules = append(oMtaInput.Modules, &oOriginalModule)
			jsonRootData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonModuleData, err := json.Marshal(oUpdatedModule)
			Ω(err).Should(Succeed())
			err = UpdateModule(mtaPath, string(jsonModuleData), marshalErr)
			Ω(err).Should(MatchError("could not marshal mta.yaml file"))
		})

		It("updates module when a module with this name exists", func() {
			oOriginalModule := Module{
				Name: "testModule",
				Type: "testType",
				Path: "test",
			}

			oUpdatedModule := Module{
				Name: "testModule",
				Type: "testType2",
				Path: "test2",
			}

			mtaPath := getTestPath("result", "temp.mta.yaml")

			oMtaInput := getMtaInput()
			oMtaInput.Modules = append(oMtaInput.Modules, &oOriginalModule)
			jsonRootData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonModuleData, err := json.Marshal(oUpdatedModule)
			Ω(err).Should(Succeed())
			Ω(UpdateModule(mtaPath, string(jsonModuleData), Marshal)).Should(Succeed())

			Ω(mtaPath).Should(BeAnExistingFile())
			yamlData, err := ioutil.ReadFile(mtaPath)
			Ω(err).Should(Succeed())
			oMtaInput = getMtaInput()
			oMtaInput.Modules = append(oMtaInput.Modules, &oUpdatedModule)
			oMtaOutput, err := Unmarshal(yamlData)
			Ω(err).Should(Succeed())
			Ω(reflect.DeepEqual(oMtaInput, *oMtaOutput)).Should(BeTrue())
		})

		It("updates one of the modules when 2 modules with this name exist", func() {
			oOriginalModule := Module{
				Name: "testModule",
				Type: "testType",
				Path: "test",
			}

			oUpdatedModule := Module{
				Name: "testModule",
				Type: "testType2",
				Path: "test2",
			}

			mtaPath := getTestPath("result", "temp.mta.yaml")

			oMtaInput := getMtaInput()
			oMtaInput.Modules = append(oMtaInput.Modules, &oOriginalModule, &oOriginalModule)
			jsonRootData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonModuleData, err := json.Marshal(oUpdatedModule)
			Ω(err).Should(Succeed())
			Ω(UpdateModule(mtaPath, string(jsonModuleData), Marshal)).Should(Succeed())

			Ω(mtaPath).Should(BeAnExistingFile())
			yamlData, err := ioutil.ReadFile(mtaPath)
			Ω(err).Should(Succeed())
			oMtaInput = getMtaInput()
			oMtaInput.Modules = append(oMtaInput.Modules, &oUpdatedModule, &oOriginalModule)
			oMtaOutput, err := Unmarshal(yamlData)
			Ω(err).Should(Succeed())
			Ω(reflect.DeepEqual(oMtaInput, *oMtaOutput)).Should(BeTrue())
		})
	})

	var _ = Describe("getModules", func() {
		It("Get modules", func() {
			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonModuleData, err := json.Marshal(oModule)
			Ω(err).Should(Succeed())
			Ω(AddModule(mtaPath, string(jsonModuleData), Marshal)).Should(Succeed())

			oMtaInput := getMtaInput()
			oMtaInput.Modules = append(oMtaInput.Modules, &oModule)
			Ω(mtaPath).Should(BeAnExistingFile())

			modules, err := GetModules(mtaPath)
			Ω(err).Should(Succeed())
			Ω(reflect.DeepEqual(oMtaInput.Modules, modules)).Should(BeTrue())
		})

		It("Get modules from a non existing mta.yaml file", func() {
			mtaPath := getTestPath("result", "mta.yaml")
			modules, err := GetModules(mtaPath)
			Ω(err).Should(HaveOccurred())
			Ω(modules).Should(BeNil())
		})
	})

	var _ = Describe("addResource", func() {
		It("Add resource", func() {
			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonResourceData, err := json.Marshal(oResource)
			Ω(err).Should(Succeed())
			Ω(AddResource(mtaPath, string(jsonResourceData), Marshal)).Should(Succeed())

			oMtaInput := getMtaInput()
			oMtaInput.Resources = append(oMtaInput.Resources, &oResource)
			Ω(mtaPath).Should(BeAnExistingFile())
			yamlData, err := ioutil.ReadFile(mtaPath)
			Ω(err).Should(Succeed())
			oMtaOutput, err := Unmarshal(yamlData)
			Ω(err).Should(Succeed())
			Ω(reflect.DeepEqual(oMtaInput, *oMtaOutput)).Should(BeTrue())
		})

		It("Add resource to non existing mta.yaml file", func() {
			json := "{name:fff}"
			mtaPath := getTestPath("result", "mta.yaml")
			Ω(AddResource(mtaPath, json, Marshal)).Should(HaveOccurred())
		})

		It("Add resource to wrong mta.yaml format", func() {
			wrongJSON := "{TEST:fff}"
			mtaPath := getTestPath("result", "mta.yaml")
			Ω(CreateMta(mtaPath, wrongJSON, os.MkdirAll)).Should(Succeed())

			jsonResourceData, err := json.Marshal(oResource)
			Ω(err).Should(Succeed())
			Ω(AddResource(mtaPath, string(jsonResourceData), Marshal)).Should(HaveOccurred())
		})

		It("Add resource with wrong json format", func() {
			wrongJSON := "{name:fff"

			mtaPath := getTestPath("result", "temp.mta.yaml")
			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			Ω(AddResource(mtaPath, wrongJSON, Marshal)).Should(HaveOccurred())
		})

		It("Add resource fails to marshal", func() {
			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonResourceData, err := json.Marshal(oResource)
			Ω(err).Should(Succeed())
			Ω(AddResource(mtaPath, string(jsonResourceData), marshalErr)).Should(HaveOccurred())
		})
	})

	var _ = Describe("updateResource", func() {
		It("fails when mta.yaml doesn't exist", func() {
			json := "{name:fff}"
			mtaPath := getTestPath("result", "mta.yaml")
			err := UpdateResource(mtaPath, json, Marshal)
			Ω(err).Should(HaveOccurred())
		})

		It("fails when mta has wrong format", func() {
			wrongJSON := "{TEST:fff}"

			mtaPath := getTestPath("result", "mta.yaml")
			Ω(CreateMta(mtaPath, wrongJSON, os.MkdirAll)).Should(Succeed())

			jsonResourceData, err := json.Marshal(oResource)
			Ω(err).Should(Succeed())
			err = UpdateResource(mtaPath, string(jsonResourceData), Marshal)
			Ω(err).Should(MatchError(MatchRegexp("yaml: unmarshal errors")))
		})

		It("fails when input is bad json format", func() {
			wrongJSON := "{name:fff"

			mtaPath := getTestPath("result", "temp.mta.yaml")
			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			err = UpdateResource(mtaPath, wrongJSON, Marshal)
			Ω(err).Should(MatchError(MatchRegexp("line 1: did not find expected")))
		})

		It("fails when resource with this name doesn't exist", func() {
			oOriginalResource := Resource{
				Name: "testResource",
				Type: "testType",
			}

			oUpdatedResource := Resource{
				Name: "testResource2",
				Type: "testType2",
			}

			mtaPath := getTestPath("result", "temp.mta.yaml")

			oMtaInput := getMtaInput()
			oMtaInput.Resources = append(oMtaInput.Resources, &oOriginalResource)
			jsonRootData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonResourceData, err := json.Marshal(oUpdatedResource)
			Ω(err).Should(Succeed())
			err = UpdateResource(mtaPath, string(jsonResourceData), Marshal)
			Ω(err).Should(MatchError("resource with name testResource2 does not exist"))
		})

		It("fails when marshal to mta.yaml fails", func() {
			oOriginalResource := Resource{
				Name: "testResource",
				Type: "testType",
			}

			oUpdatedResource := Resource{
				Name: "testResource",
				Type: "testType2",
			}

			mtaPath := getTestPath("result", "temp.mta.yaml")

			oMtaInput := getMtaInput()
			oMtaInput.Resources = append(oMtaInput.Resources, &oOriginalResource)
			jsonRootData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonResourceData, err := json.Marshal(oUpdatedResource)
			Ω(err).Should(Succeed())
			err = UpdateResource(mtaPath, string(jsonResourceData), marshalErr)
			Ω(err).Should(MatchError("could not marshal mta.yaml file"))
		})

		It("updates resource when a resource with this name exists", func() {
			oOriginalResource := Resource{
				Name: "testResource",
				Type: "testType",
			}

			oUpdatedResource := Resource{
				Name: "testResource",
				Type: "testType2",
			}

			mtaPath := getTestPath("result", "temp.mta.yaml")

			oMtaInput := getMtaInput()
			oMtaInput.Resources = append(oMtaInput.Resources, &oOriginalResource)
			jsonRootData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonResourceData, err := json.Marshal(oUpdatedResource)
			Ω(err).Should(Succeed())
			Ω(UpdateResource(mtaPath, string(jsonResourceData), Marshal)).Should(Succeed())

			Ω(mtaPath).Should(BeAnExistingFile())
			yamlData, err := ioutil.ReadFile(mtaPath)
			Ω(err).Should(Succeed())
			oMtaInput = getMtaInput()
			oMtaInput.Resources = append(oMtaInput.Resources, &oUpdatedResource)
			oMtaOutput, err := Unmarshal(yamlData)
			Ω(err).Should(Succeed())
			Ω(reflect.DeepEqual(oMtaInput, *oMtaOutput)).Should(BeTrue())
		})

		It("updates one of the resources when 2 resources with this name exist", func() {
			oOriginalResource := Resource{
				Name: "testResource",
				Type: "testType",
			}

			oUpdatedResource := Resource{
				Name: "testResource",
				Type: "testType2",
			}

			mtaPath := getTestPath("result", "temp.mta.yaml")

			oMtaInput := getMtaInput()
			oMtaInput.Resources = append(oMtaInput.Resources, &oOriginalResource, &oOriginalResource)
			jsonRootData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonResourceData, err := json.Marshal(oUpdatedResource)
			Ω(err).Should(Succeed())
			Ω(UpdateResource(mtaPath, string(jsonResourceData), Marshal)).Should(Succeed())

			Ω(mtaPath).Should(BeAnExistingFile())
			yamlData, err := ioutil.ReadFile(mtaPath)
			Ω(err).Should(Succeed())
			oMtaInput = getMtaInput()
			oMtaInput.Resources = append(oMtaInput.Resources, &oUpdatedResource, &oOriginalResource)
			oMtaOutput, err := Unmarshal(yamlData)
			Ω(err).Should(Succeed())
			Ω(reflect.DeepEqual(oMtaInput, *oMtaOutput)).Should(BeTrue())
		})
	})

	var _ = Describe("getResources", func() {
		It("Get resources", func() {
			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonResourceData, err := json.Marshal(oResource)
			Ω(err).Should(Succeed())
			Ω(AddResource(mtaPath, string(jsonResourceData), Marshal)).Should(Succeed())

			oMtaInput := getMtaInput()
			oMtaInput.Resources = append(oMtaInput.Resources, &oResource)
			Ω(mtaPath).Should(BeAnExistingFile())

			resources, err := GetResources(mtaPath)
			Ω(err).Should(Succeed())
			Ω(reflect.DeepEqual(oMtaInput.Resources, resources)).Should(BeTrue())
		})

		It("Get resources from a non existing mta.yaml file", func() {
			mtaPath := getTestPath("result", "mta.yaml")
			resources, err := GetResources(mtaPath)
			Ω(err).Should(HaveOccurred())
			Ω(resources).Should(BeNil())
		})
	})

	var _ = Describe("isNameUnique", func() {
		It("Check if name exists in mta.yaml", func() {
			mtaPath := getTestPath("mta.yaml")

			//verify module name exists
			exists, err := IsNameUnique(mtaPath, "backend")
			Ω(err).Should(Succeed())
			Ω(exists).Should(BeTrue())

			//verify provides name exists
			exists, err = IsNameUnique(mtaPath, "backend_task")
			Ω(err).Should(Succeed())
			Ω(exists).Should(BeTrue())

			//verify resource name exists
			exists, err = IsNameUnique(mtaPath, "database")
			Ω(err).Should(Succeed())
			Ω(exists).Should(BeTrue())

			//verify random name doesn't exist
			exists, err = IsNameUnique(mtaPath, "blablabla")
			Ω(err).Should(Succeed())
			Ω(exists).ShouldNot(BeTrue())
		})

		It("Check if name exists in a non existing mta.yaml file ", func() {
			mtaPath := getTestPath("result", "mta.yaml")
			_, err := IsNameUnique(mtaPath, oModule.Name)
			Ω(err).Should(HaveOccurred())
		})
	})
})

var _ = Describe("Module", func() {
	oModule := Module{
		Name: "testModule",
		Type: "testType",
		Path: "test",
	}

	AfterEach(func() {
		err := os.RemoveAll(getTestPath("result"))
		Ω(err).Should(Succeed())
	})

	It("Sanity", func() {
		os.MkdirAll(getTestPath("result"), os.ModePerm)
		mtaPath := getTestPath("result", "mta.yaml")
		Ω(CopyFile(getTestPath("mta.yaml"), mtaPath, os.Create)).Should(Succeed())

		var err error
		mtaHashCode, exists, err := GetMtaHash(mtaPath)
		Ω(err).Should(Succeed())
		Ω(exists).Should(BeTrue())

		jsonData, err := json.Marshal(oModule)
		moduleJSON := string(jsonData)
		mtaHashCodeResult, err := ModifyMta(mtaPath, func() error {
			return AddModule(mtaPath, moduleJSON, Marshal)
		}, mtaHashCode, false)
		Ω(err).Should(Succeed())
		Ω(mtaHashCodeResult).ShouldNot(Equal(mtaHashCode))
		mtaHashCodeAfterModify, _, err := GetMtaHash(mtaPath)
		Ω(err).Should(Succeed())
		Ω(mtaHashCodeResult).Should(Equal(mtaHashCodeAfterModify))
		// wrong yaml
		_, err = ModifyMta(getTestPath("result", "mtaX.yaml"), func() error {
			return AddModule(getTestPath("result", "mtaX.yaml"), moduleJSON, Marshal)
		}, mtaHashCode, false)
		Ω(err).Should(HaveOccurred())
		Ω(err.Error()).Should(ContainSubstring("file does not exist"))
		oModule.Name = "test1"
		jsonData, err = json.Marshal(oModule)
		moduleJSON = string(jsonData)
		// hashcode of the mta.yaml is wrong now
		_, err = ModifyMta(mtaPath, func() error {
			return AddModule(mtaPath, moduleJSON, Marshal)
		}, mtaHashCode, false)
		Ω(err).Should(HaveOccurred())
	})
	It("2 parallel processes, second fails to make locking", func() {
		os.MkdirAll(getTestPath("result"), os.ModePerm)
		mtaPath := getTestPath("result", "mta.yaml")
		Ω(CopyFile(getTestPath("mta.yaml"), mtaPath, os.Create)).Should(Succeed())
		mtaHashCode, _, err := GetMtaHash(mtaPath)
		Ω(err).Should(Succeed())
		var wg sync.WaitGroup
		wg.Add(1)
		var err1 error
		go func() {
			_, err1 = ModifyMta(mtaPath, func() error {
				time.Sleep(time.Second)
				return nil
			}, mtaHashCode, false)
			defer wg.Done()
		}()
		time.Sleep(time.Millisecond * 200)
		wg.Add(1)
		var err2 error
		go func() {
			_, err2 = ModifyMta(mtaPath, func() error {
				time.Sleep(time.Second)
				return nil
			}, mtaHashCode, false)
			defer wg.Done()
		}()
		wg.Wait()
		Ω(err1 == nil && err2 != nil || err1 != nil && err2 == nil).Should(BeTrue())
		if err1 == nil {
			Ω(err2.Error()).Should(ContainSubstring("is locked"))
		} else {
			Ω(err1.Error()).Should(ContainSubstring("is locked"))
		}
	})
})

var _ = Describe("RunE helper functions", func() {
	AfterEach(func() {
		err := os.RemoveAll(getTestPath("result"))
		Ω(err).Should(Succeed())
	})

	It("RunModifyAndWriteHash performs the action and writes the hashcode to the output when there is no error", func() {
		err := os.MkdirAll(getTestPath("result"), os.ModePerm)
		Ω(err).Should(Succeed())
		mtaPath := getTestPath("result", "temp.mta.yaml")
		json, err := json.Marshal(getMtaInput())
		Ω(err).Should(Succeed())
		output := executeAndProvideOutput(func() {
			err = RunModifyAndWriteHash("info message", mtaPath, func() error {
				return CreateMta(mtaPath, string(json), os.MkdirAll)
			}, 0, true)
			Ω(err).Should(Succeed())
		})
		// Check the last line of the result is a json with hashcode and that it's is not 0
		Ω(output).Should(MatchRegexp(`{"hashcode":[1-9][0-9]*}$`))
		// Note: the info message is written to the logger but we don't test it because the logger is initialized
		// with stdout before it's replaced in the test
	})

	It("RunModifyAndWriteHash writes the error when the action fails", func() {
		err := os.MkdirAll(getTestPath("result"), os.ModePerm)
		Ω(err).Should(Succeed())
		mtaPath := getTestPath("result", "temp.mta.yaml")
		output := executeAndProvideOutput(func() {
			err := RunModifyAndWriteHash("info message", mtaPath, func() error {
				return errors.New("some error")
			}, 0, true)
			Ω(err).Should(MatchError("some error"))
		})
		// Check the last line of the result is a json with hashcode and that it's is not 0
		Ω(output).Should(MatchRegexp(`{"message":"some error"}$`))
		// Note: the info message is written to the logger but we don't test it because the logger is initialized
		// with stdout before it's replaced in the test
	})

	It("RunAndWriteResultAndHash performs the action and writes the hashcode and result to the output when there is no error", func() {
		err := os.MkdirAll(getTestPath("result"), os.ModePerm)
		Ω(err).Should(Succeed())
		mtaPath := getTestPath("result", "temp.mta.yaml")
		output := executeAndProvideOutput(func() {
			err = RunAndWriteResultAndHash("info message", mtaPath, func() (interface{}, error) {
				err1 := CopyFile(getTestPath("mta.yaml"), mtaPath, os.Create)
				Ω(err1).Should(Succeed())
				return 1, nil
			})
			Ω(err).Should(Succeed())
		})
		// Check the last line of the result is a json with hashcode and that it's is not 0
		Ω(output).Should(MatchRegexp(`{"result":1,"hashcode":[1-9][0-9]*}$`))
		// Note: the info message is written to the logger but we don't test it because the logger is initialized
		// with stdout before it's replaced in the test
	})

	It("RunAndWriteResultAndHash writes the error when the action fails", func() {
		err := os.MkdirAll(getTestPath("result"), os.ModePerm)
		Ω(err).Should(Succeed())
		mtaPath := getTestPath("result", "temp.mta.yaml")
		output := executeAndProvideOutput(func() {
			err := RunAndWriteResultAndHash("info message", mtaPath, func() (interface{}, error) {
				return nil, errors.New("some error")
			})
			Ω(err).Should(MatchError("some error"))
		})
		// Check the last line of the result is a json with hashcode and that it's is not 0
		Ω(output).Should(MatchRegexp(`{"message":"some error"}$`))
		// Note: the info message is written to the logger but we don't test it because the logger is initialized
		// with stdout before it's replaced in the test
	})
})

func mkDirsErr(path string, perm os.FileMode) error {
	return errors.New("err")
}

func createErr(path string) (*os.File, error) {
	return nil, errors.New("err")
}

func marshalErr(o *MTA) ([]byte, error) {
	return nil, errors.New("could not marshal mta.yaml file")
}

func executeAndProvideOutput(execute func()) string {
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	execute()

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		if err != nil {
			fmt.Println(err)
		}
		outC <- buf.String()
	}()

	// back to normal state
	_ = w.Close()
	os.Stdout = old // restoring the real stdout
	out := <-outC
	return out
}
