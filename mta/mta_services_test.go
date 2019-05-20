package mta

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MtaServices", func() {

	schemaVersion := "1.1"
	oMtaInput := MTA{
		ID:            "test",
		Version:       "1.2",
		SchemaVersion: &schemaVersion,
		Description:   "test mta creation",
	}

	oModule := Module{
		Name: "testModule",
		Type: "testType",
		Path: "test",
	}

	oResource := Resource{
		Name: "testResource",
		Type: "testType",
	}

	oProvides := Provides{
		Name: "testProvides",
	}

	AfterEach(func() {
		err := os.RemoveAll(getTestPath("result"))
		Ω(err).Should(Succeed())
	})

	var _ = Describe("CreateMta", func() {
		It("Create MTA", func() {
			jsonData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			mtaPath := getTestPath("result", "temp.mta.yaml")
			Ω(CreateMta(mtaPath, string(jsonData), os.MkdirAll)).Should(Succeed())
			Ω(mtaPath).Should(BeAnExistingFile())
			yamlData, err := ioutil.ReadFile(mtaPath)
			Ω(err).Should(Succeed())
			oMtaOutput, err := Unmarshal(yamlData)
			Ω(err).Should(Succeed())
			Ω(reflect.DeepEqual(oMtaInput, *oMtaOutput)).Should(BeTrue())
		})

		It("Create MTA with wrong json format", func() {
			wrongJSON := "{Name:fff"
			mtaPath := getTestPath("result", "temp.mta.yaml")
			Ω(CreateMta(mtaPath, wrongJSON, os.MkdirAll)).Should(HaveOccurred())
		})

		It("Create MTA fail to create file", func() {
			jsonData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			mtaPath := getTestPath("result", "temp.mta.yaml")
			Ω(CreateMta(mtaPath, string(jsonData), mkDirsErr)).Should(HaveOccurred())
		})
	})

	var _ = Describe("CopyFile", func() {
		It("Copy file content", func() {
			jsonData, err := json.Marshal(oMtaInput)
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
			Ω(reflect.DeepEqual(oMtaInput, *oOutput)).Should(BeTrue())
		})

		It("Copy file with non existing path", func() {
			sourceFilePath := "c:/temp/test1"
			targetFilePath := "c:/temp/test2"
			Ω(CopyFile(sourceFilePath, targetFilePath, os.Create)).Should(HaveOccurred())
		})

		It("Copy file fail to create destination file", func() {
			jsonData, err := json.Marshal(oMtaInput)
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
			jsonData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			mtaPath := getTestPath("result", "temp.mta.yaml")
			Ω(CreateMta(mtaPath, string(jsonData), os.MkdirAll)).Should(Succeed())
			Ω(mtaPath).Should(BeAnExistingFile())
			Ω(DeleteFile(mtaPath)).Should(Succeed())
			Ω(mtaPath).ShouldNot(BeAnExistingFile())
		})
	})

	var _ = Describe("addModule", func() {
		It("Add module", func() {
			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonModuleData, err := json.Marshal(oModule)
			Ω(err).Should(Succeed())
			Ω(AddModule(mtaPath, string(jsonModuleData), Marshal)).Should(Succeed())

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
			jsonRootData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			Ω(AddModule(mtaPath, wrongJSON, Marshal)).Should(HaveOccurred())
		})

		It("Add module fails to marshal", func() {
			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonModuleData, err := json.Marshal(oModule)
			Ω(err).Should(Succeed())
			Ω(AddModule(mtaPath, string(jsonModuleData), marshalErr)).Should(HaveOccurred())
		})
	})

	var _ = Describe("getModules", func() {
		It("Get modules", func() {
			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonModuleData, err := json.Marshal(oModule)
			Ω(err).Should(Succeed())
			Ω(AddModule(mtaPath, string(jsonModuleData), Marshal)).Should(Succeed())

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

			jsonRootData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonResourceData, err := json.Marshal(oResource)
			Ω(err).Should(Succeed())
			Ω(AddResource(mtaPath, string(jsonResourceData), Marshal)).Should(Succeed())

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
			jsonRootData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			Ω(AddResource(mtaPath, wrongJSON, Marshal)).Should(HaveOccurred())
		})

		It("Add resource fails to marshal", func() {
			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonResourceData, err := json.Marshal(oResource)
			Ω(err).Should(Succeed())
			Ω(AddResource(mtaPath, string(jsonResourceData), marshalErr)).Should(HaveOccurred())
		})
	})

	var _ = Describe("getResources", func() {
		It("Get resources", func() {
			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonResourceData, err := json.Marshal(oResource)
			Ω(err).Should(Succeed())
			Ω(AddResource(mtaPath, string(jsonResourceData), Marshal)).Should(Succeed())

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
			mtaPath := getTestPath("result", "mta.yaml")

			jsonRootData, err := json.Marshal(oMtaInput)
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			oModule.Provides = append(oModule.Provides, oProvides)
			jsonModuleData, err := json.Marshal(oModule)
			Ω(err).Should(Succeed())
			Ω(AddModule(mtaPath, string(jsonModuleData), Marshal)).Should(Succeed())
			oMtaInput.Modules = append(oMtaInput.Modules, &oModule)

			jsonResourceData, err := json.Marshal(oResource)
			Ω(err).Should(Succeed())
			Ω(AddResource(mtaPath, string(jsonResourceData), Marshal)).Should(Succeed())
			oMtaInput.Resources = append(oMtaInput.Resources, &oResource)

			Ω(mtaPath).Should(BeAnExistingFile())

			exists, err := IsNameUnique(mtaPath, oModule.Name)
			Ω(err).Should(Succeed())
			Ω(exists).Should(BeTrue())

			exists, err = IsNameUnique(mtaPath, oModule.Name+"aaaaa")
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
		err = ModifyMta(mtaPath, func() error {
			return AddModule(mtaPath, moduleJSON, Marshal)
		}, mtaHashCode, false)
		Ω(err).Should(Succeed())
		// wrong yaml
		err = ModifyMta(getTestPath("result", "mtaX.yaml"), func() error {
			return AddModule(getTestPath("result", "mtaX.yaml"), moduleJSON, Marshal)
		}, mtaHashCode, false)
		Ω(err).Should(HaveOccurred())
		Ω(err.Error()).Should(ContainSubstring("file does not exist"))
		oModule.Name = "test1"
		jsonData, err = json.Marshal(oModule)
		moduleJSON = string(jsonData)
		// hashcode of the mta.yaml is wrong now
		err = ModifyMta(mtaPath, func() error {
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
			err1 = ModifyMta(mtaPath, func() error {
				time.Sleep(time.Second)
				return nil
			}, mtaHashCode, false)
			defer wg.Done()
		}()
		time.Sleep(time.Millisecond * 200)
		wg.Add(1)
		var err2 error
		go func() {
			err2 = ModifyMta(mtaPath, func() error {
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

func mkDirsErr(path string, perm os.FileMode) error {
	return errors.New("err")
}

func createErr(path string) (*os.File, error) {
	return nil, errors.New("err")
}

func marshalErr(o *MTA) ([]byte, error) {
	return nil, errors.New("err")
}
