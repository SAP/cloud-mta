package mta

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/SAP/cloud-mta/internal/fs"
)

func getMtaInput() MTA {
	// Return a new object every time so we don't accidentally change it for the other tests
	schemaVersion := "1.1"
	oMtaInput := MTA{
		ID:            "test",
		Version:       "1.2",
		SchemaVersion: &schemaVersion,
		Description:   "test mta creation",
		Parameters:    map[string]interface{}{"param1": "value1", "param2": "value2"},
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
		err2 := os.RemoveAll(getTestPath("result2"))
		Ω(err2).Should(Succeed())
	})

	Describe("GetMtaFromFile", func() {
		wd, _ := os.Getwd()
		mtaYamlSchemaVersion := "2.1"

		It("returns MTA for valid filename without extensions", func() {
			mta, messages, err := GetMtaFromFile(filepath.Join(wd, "testdata", "mtaValid.yaml"), nil, true)
			Ω(err).Should(Succeed())
			Ω(messages).Should(BeEmpty())
			Ω(mta).ShouldNot(BeNil())
			Ω(*mta).Should(Equal(MTA{
				ID:            "demo",
				SchemaVersion: &mtaYamlSchemaVersion,
				Version:       "0.0.1",
				Modules: []*Module{
					{
						Name: "srv",
						Type: "java",
						Path: "srv",
						Properties: map[string]interface{}{
							"APPC_LOG_LEVEL":              "info",
							"VSCODE_JAVA_DEBUG_LOG_LEVEL": "ALL",
						},
						Parameters: map[string]interface{}{
							"memory": "512M",
						},
						Provides: []Provides{
							{
								Name: "srv_api",
								Properties: map[string]interface{}{
									"url": "${default-url}",
								},
							},
						},
						Requires: []Requires{
							{
								Name: "db",
								Properties: map[string]interface{}{
									"JBP_CONFIG_RESOURCE_CONFIGURATION": "[tomcat/webapps/ROOT/META-INF/context.xml: {\"service_name_for_DefaultDB\" : \"~{hdi-container-name}\"}]",
								},
							},
						},
					},
					{
						Name: "ui",
						Type: "html5",
						Path: "ui",
						Parameters: map[string]interface{}{
							"disk-quota": "256M",
							"memory":     "256M",
						},
						BuildParams: map[string]interface{}{
							"builder": "grunt",
						},
						Requires: []Requires{
							{
								Name:  "srv_api",
								Group: "destinations",
								Properties: map[string]interface{}{
									"forwardAuthToken": true,
									"strictSSL":        false,
									"name":             "srv_api",
									"url":              "~{url}",
								},
							},
						},
					},
				},
				Resources: []*Resource{
					{
						Name: "hdi_db",
						Type: "com.company.xs.hdi-container",
						Properties: map[string]interface{}{
							"hdi-container-name": "${service-name}",
						},
					},
				},
			}))
		})
		It("returns error for invalid filename", func() {
			_, messages, err := GetMtaFromFile(filepath.Join(wd, "testdata", "mtaNonExisting.yaml"), nil, true)
			Ω(err).Should(HaveOccurred())
			Ω(messages).Should(BeEmpty())
		})
		It("returns error for invalid mta yaml file", func() {
			_, messages, err := GetMtaFromFile(filepath.Join(wd, "testdata", "mtaInvalid.yaml"), nil, true)
			Ω(err).Should(HaveOccurred())
			Ω(messages).Should(BeEmpty())
		})
		It("returns MTA with merged extensions for valid mta.yaml and extensions", func() {
			mtaPath := filepath.Join(wd, "testdata", "testext", "mta.yaml")
			extPath := filepath.Join(wd, "testdata", "testext", "cf-mtaext.yaml")
			mta, messages, err := GetMtaFromFile(mtaPath, []string{extPath}, true)
			Ω(err).Should(Succeed())
			Ω(messages).Should(BeEmpty())
			Ω(mta).ShouldNot(BeNil())
			activeFalse := false
			expected := MTA{
				ID:            "mtahtml5",
				SchemaVersion: &mtaYamlSchemaVersion,
				Version:       "0.0.1",
				Modules: []*Module{
					{
						Name: "ui5app",
						Type: "html5",
						Path: "ui5app",
						Parameters: map[string]interface{}{
							"disk-quota": "256M",
							"memory":     "256M",
						},
						Properties: map[string]interface{}{
							"my_prop": 1,
						},
						Requires: []Requires{
							{
								Name: "uaa_mtahtml5",
							},
						},
						BuildParams: map[string]interface{}{
							"builder": "zip",
							"ignore":  []interface{}{"ui5app/"},
						},
					},
					{
						Name: "ui5app2",
						Type: "html5",
						Parameters: map[string]interface{}{
							"disk-quota": "256M",
							"memory":     "512M",
						},
						Requires: []Requires{
							{
								Name: "uaa_mtahtml5",
							},
						},
					},
				},
				Resources: []*Resource{
					{
						Name: "uaa_mtahtml5",
						Type: "com.company.xs.uaa",
						Parameters: map[string]interface{}{
							"path":         "./xs-security.json",
							"service-plan": "application",
						},
						Active: &activeFalse,
					},
				},
			}
			Ω(*mta).Should(Equal(expected))
		})
		It("returns extensions errors in messages when extensions are invalid and returnMergeError is false", func() {
			mtaPath := filepath.Join(wd, "testdata", "testext", "mta.yaml")
			extPath := filepath.Join(wd, "testdata", "testext", "unknown_extends.mtaext")
			mta, messages, err := GetMtaFromFile(mtaPath, []string{extPath}, false)
			Ω(err).Should(Succeed())
			Ω(messages).Should(ConsistOf(ContainSubstring(unknownExtendsMsg, "")))
			Ω(mta).ShouldNot(BeNil())
			expected := MTA{
				ID:            "mtahtml5",
				SchemaVersion: &mtaYamlSchemaVersion,
				Version:       "0.0.1",
				Modules: []*Module{
					{
						Name: "ui5app",
						Type: "html5",
						Path: "ui5app",
						Parameters: map[string]interface{}{
							"disk-quota": "256M",
							"memory":     "256M",
						},
						Requires: []Requires{
							{
								Name: "uaa_mtahtml5",
							},
						},
						BuildParams: map[string]interface{}{
							"builder": "zip",
							"ignore":  []interface{}{"ui5app/"},
						},
					},
					{
						Name: "ui5app2",
						Type: "html5",
						Parameters: map[string]interface{}{
							"disk-quota": "256M",
							"memory":     "256M",
						},
						Requires: []Requires{
							{
								Name: "uaa_mtahtml5",
							},
						},
					},
				},
				Resources: []*Resource{
					{
						Name: "uaa_mtahtml5",
						Type: "com.company.xs.uaa",
						Parameters: map[string]interface{}{
							"path":         "./xs-security.json",
							"service-plan": "application",
						},
					},
				},
			}
			Ω(*mta).Should(Equal(expected))
		})
		It("returns extensions errors as error when extensions are invalid and returnMergeError is true", func() {
			mtaPath := filepath.Join(wd, "testdata", "testext", "mta.yaml")
			extPath := filepath.Join(wd, "testdata", "testext", "unknown_extends.mtaext")
			mta, messages, err := GetMtaFromFile(mtaPath, []string{extPath}, true)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(ContainSubstring(unknownExtendsMsg, ""))
			Ω(messages).Should(BeEmpty())
			Ω(mta).ShouldNot(BeNil())
			expected := MTA{
				ID:            "mtahtml5",
				SchemaVersion: &mtaYamlSchemaVersion,
				Version:       "0.0.1",
				Modules: []*Module{
					{
						Name: "ui5app",
						Type: "html5",
						Path: "ui5app",
						Parameters: map[string]interface{}{
							"disk-quota": "256M",
							"memory":     "256M",
						},
						Requires: []Requires{
							{
								Name: "uaa_mtahtml5",
							},
						},
						BuildParams: map[string]interface{}{
							"builder": "zip",
							"ignore":  []interface{}{"ui5app/"},
						},
					},
					{
						Name: "ui5app2",
						Type: "html5",
						Parameters: map[string]interface{}{
							"disk-quota": "256M",
							"memory":     "256M",
						},
						Requires: []Requires{
							{
								Name: "uaa_mtahtml5",
							},
						},
					},
				},
				Resources: []*Resource{
					{
						Name: "uaa_mtahtml5",
						Type: "com.company.xs.uaa",
						Parameters: map[string]interface{}{
							"path":         "./xs-security.json",
							"service-plan": "application",
						},
					},
				},
			}
			Ω(*mta).Should(Equal(expected))
		})
		It("returns error on extension when an extension version mismatches the MTA version", func() {
			mtaPath := filepath.Join(wd, "testdata", "testext", "mta.yaml")
			extPath := filepath.Join(wd, "testdata", "testext", "bad_version.mtaext")
			_, _, err := GetMtaFromFile(mtaPath, []string{extPath}, true)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(ContainSubstring(versionMismatchMsg, "3.1", extPath, mtaYamlSchemaVersion))
		})
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

	var _ = Describe("DeleteMta", func() {
		It("Delete MTA", func() {
			jsonData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			mtaPath := getTestPath("result", "temp.mta.yaml")
			Ω(CreateMta(mtaPath, string(jsonData), os.MkdirAll)).Should(Succeed())
			Ω(mtaPath).Should(BeAnExistingFile())
			mtaProject := getTestPath("result")
			Ω(DeleteMta(mtaProject)).Should(Succeed())
			Ω(mtaPath).ShouldNot(BeAnExistingFile())
			Ω(mtaProject).ShouldNot(BeADirectory())
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

		It("Copy file creates the destination folder", func() {
			jsonData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			sourceFilePath := getTestPath("result", "temp.mta.yaml")
			targetFilePath := getTestPath("result2", "temp2.mta.yaml")
			Ω(CreateMta(sourceFilePath, string(jsonData), os.MkdirAll)).Should(Succeed())
			Ω(CopyFile(sourceFilePath, targetFilePath, os.Create)).Should(Succeed())
			Ω(targetFilePath).Should(BeAnExistingFile())
		})

		It("Copy file fail to create destination folder", func() {
			jsonData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			sourceFilePath := getTestPath("result", "temp.mta.yaml")
			targetFolderPath := getTestPath("result2")
			Ω(CreateMta(sourceFilePath, string(jsonData), os.MkdirAll)).Should(Succeed())
			Ω(CreateMta(targetFolderPath, string(jsonData), os.MkdirAll)).Should(Succeed())
			targetFilePath := getTestPath("result2", "temp2.mta.yaml")
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

	var _ = Describe("DeleteFile", func() {
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

		It("Writes only the hashcode when the result, messages and error are nil", func() {
			err := printResult(nil, nil, 123, nil, printer, json.Marshal)
			Ω(err).Should(Succeed())
			Ω(printed).Should(Equal(`{"hashcode":123}`))
		})

		It("Writes error message when the error is not nil", func() {
			err := printResult("123", nil, 123, errors.New("error message"), printer, json.Marshal)
			Ω(err).Should(Succeed())
			Ω(printed).Should(Equal(`{"message":"error message"}`))
		})

		It("Writes hashcode, messages and result when the result is sent and there is no error", func() {
			err := printResult("1234", []string{"some message"}, 3, nil, printer, json.Marshal)
			Ω(err).Should(Succeed())
			Ω(printed).Should(Equal(`{"result":"1234","messages":["some message"],"hashcode":3}`))
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
			err := printResult(modules, nil, 0, nil, printer, json.Marshal)
			Ω(err).Should(Succeed())
			Ω(printed).Should(Equal(`{"result":[{"name":"m1","type":"type1"},{"name":"m2","type":"type2"}],"hashcode":0}`))
		})

		It("Returns print error if print fails", func() {
			printerErr := func(s ...interface{}) (int, error) {
				return 0, errors.New("error in print")
			}
			err := printResult(nil, nil, 1, nil, printerErr, json.Marshal)
			Ω(err).Should(MatchError("error in print"))
		})

		It("Returns and writes error if the result cannot be serialized to JSON", func() {
			var unserializableResult UnmarshalableString = "a"
			err := printResult(unserializableResult, nil, 0, nil, printer, json.Marshal)
			Ω(err).Should(MatchError(ContainSubstring("cannot marshal value a")))
			Ω(printed).Should(ContainSubstring("cannot marshal value a"))
		})

		It("Returns and writes error if the error message cannot be serialized to JSON", func() {
			err := printResult(nil, nil, 0, errors.New("some error"), printer, jsonMarshalErr)
			Ω(err).Should(MatchError("could not marshal to json"))
			// Both error messages should be printed to the output
			Ω(printed).Should(ContainSubstring("could not marshal to json"))
			Ω(printed).Should(ContainSubstring("some error"))
		})
	})

	var _ = Describe("AddModule", func() {
		It("Add module", func() {
			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonModuleData, err := json.Marshal(oModule)
			Ω(err).Should(Succeed())
			messages, err := AddModule(mtaPath, string(jsonModuleData), Marshal)
			Ω(err).Should(Succeed())
			Ω(messages).Should(BeEmpty())

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
			const json = "{name:fff}"
			mtaPath := getTestPath("result", "mta.yaml")
			_, err := AddModule(mtaPath, json, Marshal)
			Ω(err).Should(HaveOccurred())
		})

		It("Add module to wrong mta.yaml format", func() {
			const wrongJSON = "{TEST:fff}"
			mtaPath := getTestPath("result", "mta.yaml")
			Ω(CreateMta(mtaPath, wrongJSON, os.MkdirAll)).Should(Succeed())

			jsonModuleData, err := json.Marshal(oModule)
			Ω(err).Should(Succeed())
			_, err = AddModule(mtaPath, string(jsonModuleData), Marshal)
			Ω(err).Should(HaveOccurred())
		})

		It("Add module with wrong json format", func() {
			const wrongJSON = "{name:fff"

			mtaPath := getTestPath("result", "temp.mta.yaml")
			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			_, err = AddModule(mtaPath, wrongJSON, Marshal)
			Ω(err).Should(HaveOccurred())
		})

		It("Add module fails to marshal", func() {
			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonModuleData, err := json.Marshal(oModule)
			Ω(err).Should(Succeed())
			_, err = AddModule(mtaPath, string(jsonModuleData), marshalErr)
			Ω(err).Should(HaveOccurred())
		})
	})

	var _ = Describe("UpdateModule", func() {
		It("fails when mta.yaml doesn't exist", func() {
			const json = "{name:fff}"
			mtaPath := getTestPath("result", "mta.yaml")
			_, err := UpdateModule(mtaPath, json, Marshal)
			Ω(err).Should(HaveOccurred())
		})

		It("fails when mta has wrong format", func() {
			const wrongJSON = "{TEST:fff}"

			mtaPath := getTestPath("result", "mta.yaml")
			Ω(CreateMta(mtaPath, wrongJSON, os.MkdirAll)).Should(Succeed())

			jsonModuleData, err := json.Marshal(oModule)
			Ω(err).Should(Succeed())
			_, err = UpdateModule(mtaPath, string(jsonModuleData), Marshal)
			Ω(err).Should(MatchError(MatchRegexp("yaml: unmarshal errors")))
		})

		It("fails when input is bad json format", func() {
			const wrongJSON = "{name:fff"

			mtaPath := getTestPath("result", "temp.mta.yaml")
			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			_, err = UpdateModule(mtaPath, wrongJSON, Marshal)
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
			_, err = UpdateModule(mtaPath, string(jsonModuleData), Marshal)
			Ω(err).Should(MatchError("the 'testModule2' module does not exist"))
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
			_, err = UpdateModule(mtaPath, string(jsonModuleData), marshalErr)
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
			messages, err := UpdateModule(mtaPath, string(jsonModuleData), Marshal)
			Ω(err).Should(Succeed())
			Ω(messages).Should(BeEmpty())

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
			messages, err := UpdateModule(mtaPath, string(jsonModuleData), Marshal)
			Ω(err).Should(Succeed())
			Ω(messages).Should(BeEmpty())

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

	var _ = Describe("GetModules", func() {
		It("returns the modules from the mta.yaml when there are no extensions", func() {
			mtaPath := getTestPath("mta_module.yaml")

			modules, messages, err := GetModules(mtaPath, nil)
			Ω(err).Should(Succeed())
			Ω(modules).Should(Equal([]*Module{&oModule}))
			Ω(messages).Should(BeEmpty())
		})

		It("returns the modules from the mta.yaml when the extensions don't exist", func() {
			mtaPath := getTestPath("mta_module.yaml")

			mtaExtPath := getTestPath("result", "nonExisting.mtaext")
			modules, messages, err := GetModules(mtaPath, []string{mtaExtPath})
			Ω(err).Should(Succeed())
			Ω(modules).Should(Equal([]*Module{&oModule}))
			Ω(messages).Should(ContainElement(ContainSubstring(fs.PathNotFoundMsg, mtaExtPath)))
		})

		It("returns the merged modules from the mta.yaml and extensions", func() {
			mtaPath := getTestPath("mta_module.yaml")

			mergedModule := Module{
				Name: "testModule",
				Type: "testType",
				Path: "test",
				Properties: map[string]interface{}{
					`commonProp`: `value2`,
					"newProp":    "new value",
					"newProp2":   "new value2",
				},
			}
			expectedModules := []*Module{&mergedModule}

			modules, messages, err := GetModules(mtaPath, []string{
				getTestPath("module_valid1.mtaext"),
				getTestPath("module_valid2.mtaext"),
			})
			Ω(err).Should(Succeed())
			Ω(modules).Should(Equal(expectedModules))
			Ω(messages).Should(BeEmpty())
		})

		It("returns the partially merged modules from the mta.yaml and extensions until reaching an invalid extension", func() {
			mtaPath := getTestPath("mta_module.yaml")

			mergedModule := Module{
				Name: "testModule",
				Type: "testType",
				Path: "test",
				Properties: map[string]interface{}{
					`commonProp`: `value2`,
					"newProp":    "new value",
					"newProp2":   "new value2",
				},
			}
			expectedModules := []*Module{&mergedModule}

			modules, messages, err := GetModules(mtaPath, []string{
				getTestPath("module_valid1.mtaext"),
				getTestPath("module_invalid2.mtaext"),
				getTestPath("module_valid3.mtaext"),
			})
			Ω(err).Should(Succeed())
			Ω(modules).Should(Equal(expectedModules))
			Ω(messages).Should(ContainElement(ContainSubstring(`testModuleNonExisting`)))
		})

		It("returns an error when mta.yaml does not exist", func() {
			mtaPath := getTestPath("result", "mta.yaml")
			modules, messages, err := GetModules(mtaPath, nil)
			Ω(err).Should(HaveOccurred())
			Ω(modules).Should(BeNil())
			Ω(messages).Should(BeEmpty())
		})
	})

	var _ = Describe("AddResource", func() {
		It("Add resource", func() {
			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonResourceData, err := json.Marshal(oResource)
			Ω(err).Should(Succeed())
			messages, err := AddResource(mtaPath, string(jsonResourceData), Marshal)
			Ω(err).Should(Succeed())
			Ω(messages).Should(BeEmpty())

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
			const json = "{name:fff}"
			mtaPath := getTestPath("result", "mta.yaml")
			_, err := AddResource(mtaPath, json, Marshal)
			Ω(err).Should(HaveOccurred())
		})

		It("Add resource to wrong mta.yaml format", func() {
			wrongJSON := "{TEST:fff}"
			mtaPath := getTestPath("result", "mta.yaml")
			Ω(CreateMta(mtaPath, wrongJSON, os.MkdirAll)).Should(Succeed())

			jsonResourceData, err := json.Marshal(oResource)
			Ω(err).Should(Succeed())
			_, err = AddResource(mtaPath, string(jsonResourceData), Marshal)
			Ω(err).Should(HaveOccurred())
		})

		It("Add resource with wrong json format", func() {
			const wrongJSON = "{name:fff"

			mtaPath := getTestPath("result", "temp.mta.yaml")
			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			_, err = AddResource(mtaPath, wrongJSON, Marshal)
			Ω(err).Should(HaveOccurred())
		})

		It("Add resource fails to marshal", func() {
			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonResourceData, err := json.Marshal(oResource)
			Ω(err).Should(Succeed())
			_, err = AddResource(mtaPath, string(jsonResourceData), marshalErr)
			Ω(err).Should(HaveOccurred())
		})
	})

	var _ = Describe("UpdateResource", func() {
		It("fails when mta.yaml doesn't exist", func() {
			json := "{name:fff}"
			mtaPath := getTestPath("result", "mta.yaml")
			_, err := UpdateResource(mtaPath, json, Marshal)
			Ω(err).Should(HaveOccurred())
		})

		It("fails when mta has wrong format", func() {
			wrongJSON := "{TEST:fff}"

			mtaPath := getTestPath("result", "mta.yaml")
			Ω(CreateMta(mtaPath, wrongJSON, os.MkdirAll)).Should(Succeed())

			jsonResourceData, err := json.Marshal(oResource)
			Ω(err).Should(Succeed())
			_, err = UpdateResource(mtaPath, string(jsonResourceData), Marshal)
			Ω(err).Should(MatchError(MatchRegexp("yaml: unmarshal errors")))
		})

		It("fails when input is bad json format", func() {
			wrongJSON := "{name:fff"

			mtaPath := getTestPath("result", "temp.mta.yaml")
			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			_, err = UpdateResource(mtaPath, wrongJSON, Marshal)
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
			_, err = UpdateResource(mtaPath, string(jsonResourceData), Marshal)
			Ω(err).Should(MatchError("the 'testResource2' resource does not exist"))
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
			_, err = UpdateResource(mtaPath, string(jsonResourceData), marshalErr)
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
			messages, err := UpdateResource(mtaPath, string(jsonResourceData), Marshal)
			Ω(err).Should(Succeed())
			Ω(messages).Should(BeEmpty())

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
			messages, err := UpdateResource(mtaPath, string(jsonResourceData), Marshal)
			Ω(err).Should(Succeed())
			Ω(messages).Should(BeEmpty())

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

	var _ = Describe("GetResources", func() {
		var createMta = func() string {
			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonResourceData, err := json.Marshal(oResource)
			Ω(err).Should(Succeed())
			messages, err := AddResource(mtaPath, string(jsonResourceData), Marshal)
			Ω(err).Should(Succeed())
			Ω(messages).Should(BeEmpty())

			Ω(mtaPath).Should(BeAnExistingFile())
			return mtaPath
		}

		It("returns the resources from the mta.yaml when there are no extensions", func() {
			mtaPath := createMta()

			oMtaInput := getMtaInput()
			oMtaInput.Resources = append(oMtaInput.Resources, &oResource)

			resources, messages, err := GetResources(mtaPath, nil)
			Ω(err).Should(Succeed())
			Ω(resources).Should(Equal([]*Resource{&oResource}))
			Ω(messages).Should(BeEmpty())
		})

		It("returns the resources from the mta.yaml when the extensions don't exist", func() {
			mtaPath := createMta()

			mtaExtPath := getTestPath("result", "nonExisting.mtaext")
			resources, messages, err := GetResources(mtaPath, []string{mtaExtPath})
			Ω(err).Should(Succeed())
			Ω(resources).Should(Equal([]*Resource{&oResource}))
			Ω(messages).Should(ContainElement(ContainSubstring(fs.PathNotFoundMsg, mtaExtPath)))
		})

		It("returns the merged resources from the mta.yaml and extensions", func() {
			mtaPath := createMta()

			mergedResource := Resource{
				Name: "testResource",
				Type: "testType",
				Properties: map[string]interface{}{
					"newProp":  "new value",
					"newProp2": "new value2",
				},
			}
			expectedResources := []*Resource{&mergedResource}

			resources, messages, err := GetResources(mtaPath, []string{
				getTestPath("resource_valid1.mtaext"),
				getTestPath("resource_valid2.mtaext"),
			})
			Ω(err).Should(Succeed())
			Ω(resources).Should(Equal(expectedResources))
			Ω(messages).Should(BeEmpty())
		})

		It("returns the partially merged resources from the mta.yaml and extensions until reaching an invalid extension", func() {
			mtaPath := createMta()

			mergedResource := Resource{
				Name: "testResource",
				Type: "testType",
				Properties: map[string]interface{}{
					"newProp":  "new value",
					"newProp2": "new value2",
				},
			}
			expectedResources := []*Resource{&mergedResource}

			resources, messages, err := GetResources(mtaPath, []string{
				getTestPath("resource_valid1.mtaext"),
				getTestPath("resource_invalid2.mtaext"),
				getTestPath("resource_valid3.mtaext"),
			})
			Ω(err).Should(Succeed())
			Ω(resources).Should(Equal(expectedResources))
			Ω(messages).Should(ContainElement(ContainSubstring(`testResourceNotExisting`)))
		})

		It("returns an error when mta.yaml does not exist", func() {
			mtaPath := getTestPath("result", "mta.yaml")
			resources, messages, err := GetResources(mtaPath, nil)
			Ω(err).Should(HaveOccurred())
			Ω(resources).Should(BeNil())
			Ω(messages).Should(BeEmpty())
		})
	})

	var _ = Describe("GetResourceConfig", func() {
		It("returns error when mta.yaml does not exist", func() {
			mtaPath := getTestPath("result", "mta.yaml")
			resourceConfig, messages, err := GetResourceConfig(mtaPath, nil, "myResource", "")
			Ω(err).Should(HaveOccurred())
			Ω(resourceConfig).Should(BeNil())
			Ω(messages).Should(BeEmpty())
		})

		It("returns error when resource does not exist", func() {
			mtaPath := getTestPath("mtaConfig.yaml")
			resourceConfig, messages, err := GetResourceConfig(mtaPath, nil, "nonExistingResource", "")
			Ω(err).Should(HaveOccurred())
			Ω(resourceConfig).Should(BeNil())
			Ω(messages).Should(BeEmpty())
		})

		It("returns error when path references a file that doesn't exist", func() {
			mtaPath := getTestPath("mtaConfig.yaml")
			resourceConfig, messages, err := GetResourceConfig(mtaPath, nil, "resourceWithPath2", "")
			Ω(err).Should(HaveOccurred())
			Ω(resourceConfig).Should(BeNil())
			Ω(messages).Should(BeEmpty())
		})

		It("returns error when path references a file that is not json", func() {
			mtaPath := getTestPath("mtaConfig.yaml")
			resourceConfig, messages, err := GetResourceConfig(mtaPath, nil, "resourceWithPath", getTestPath("otherWorkdir"))
			Ω(err).Should(HaveOccurred())
			Ω(resourceConfig).Should(BeNil())
			Ω(messages).Should(BeEmpty())
		})

		// Both xs-security.json and otherWorkdir/xs-security2.json contain this content
		fileContent := map[string]interface{}{
			"xsappname":     "nameFromPath",
			"paramFromPath": "paramValueFromPath",
			"deepParam": map[string]interface{}{
				"otherValue": "deepValueFromPath",
			},
		}

		It("returns empty map when config and path aren't defined", func() {
			mtaPath := getTestPath("mtaConfig.yaml")
			resourceConfig, messages, err := GetResourceConfig(mtaPath, nil, "resourceWithNoConfigAndPath", "")
			Ω(err).Should(Succeed())
			Ω(resourceConfig).ShouldNot(BeNil())
			result := map[string]interface{}{}
			Ω(resourceConfig).Should(Equal(result))
			Ω(messages).Should(BeEmpty())
		})

		It("returns file content when config is not defined and path exists", func() {
			mtaPath := getTestPath("mtaConfig.yaml")
			resourceConfig, messages, err := GetResourceConfig(mtaPath, nil, "resourceWithPath", "")
			Ω(err).Should(Succeed())
			Ω(resourceConfig).ShouldNot(BeNil())
			result := fileContent
			Ω(resourceConfig).Should(Equal(result))
			Ω(messages).Should(BeEmpty())
		})

		It("returns file content when config is not defined and path exists relative to workdir", func() {
			mtaPath := getTestPath("mtaConfig.yaml")
			resourceConfig, messages, err := GetResourceConfig(mtaPath, nil, "resourceWithPath2", getTestPath("otherWorkdir"))
			Ω(err).Should(Succeed())
			Ω(resourceConfig).ShouldNot(BeNil())
			result := fileContent
			Ω(resourceConfig).Should(Equal(result))
			Ω(messages).Should(BeEmpty())
		})

		It("returns file content when config is not a map and path exists", func() {
			mtaPath := getTestPath("mtaConfig.yaml")
			resourceConfig, messages, err := GetResourceConfig(mtaPath, nil, "resourceWithPathAndBadConfig", "")
			Ω(err).Should(Succeed())
			Ω(resourceConfig).ShouldNot(BeNil())
			result := fileContent
			Ω(resourceConfig).Should(Equal(result))
			Ω(messages).Should(BeEmpty())
		})

		It("returns config when it is defined and path is not defined", func() {
			mtaPath := getTestPath("mtaConfig.yaml")
			resourceConfig, messages, err := GetResourceConfig(mtaPath, nil, "resourceWithConfig", "")
			Ω(err).Should(Succeed())
			Ω(resourceConfig).ShouldNot(BeNil())
			result := map[string]interface{}{
				"xsappname": "testName",
			}
			Ω(resourceConfig).Should(Equal(result))
			Ω(messages).Should(BeEmpty())
		})

		It("merges config and file when both config and path are defined", func() {
			// Keys from config override keys from file path, and the override is shallow
			mtaPath := getTestPath("mtaConfig.yaml")
			resourceConfig, messages, err := GetResourceConfig(mtaPath, nil, "resourceWithConfigAndPath", "")
			Ω(err).Should(Succeed())
			Ω(resourceConfig).ShouldNot(BeNil())
			result := map[string]interface{}{
				"xsappname":       "nameFromConfig",
				"paramFromPath":   "paramValueFromPath",
				"paramFromConfig": "paramValueFromConfig",
				"deepParam": map[string]interface{}{
					"someValue": "deepValueFromConfig",
				},
			}
			Ω(resourceConfig).Should(Equal(result))
			Ω(messages).Should(BeEmpty())
		})

		It("returns config from mta.yaml only when extensions don't exist", func() {
			mtaPath := getTestPath("mtaConfig.yaml")
			mtaExtPath := getTestPath("nonExisting.mtaext")
			resourceConfig, messages, err := GetResourceConfig(mtaPath, []string{mtaExtPath}, "resourceWithConfig", "")
			Ω(err).Should(Succeed())
			Ω(resourceConfig).ShouldNot(BeNil())
			result := map[string]interface{}{
				"xsappname": "testName",
			}
			Ω(resourceConfig).Should(Equal(result))
			Ω(messages).Should(ContainElement(ContainSubstring(fs.PathNotFoundMsg, mtaExtPath)))
		})

		It("returns config from mta.yaml only when extensions are invalid", func() {
			mtaPath := getTestPath("mtaConfig.yaml")
			mtaExtPath := getTestPath("resourceConfig_invalid1.mtaext")
			resourceConfig, messages, err := GetResourceConfig(mtaPath, []string{mtaExtPath}, "resourceWithConfig", "")
			Ω(err).Should(Succeed())
			Ω(resourceConfig).ShouldNot(BeNil())
			result := map[string]interface{}{
				"xsappname": "testName",
			}
			Ω(resourceConfig).Should(Equal(result))
			Ω(messages).Should(ContainElement(ContainSubstring("resourceWithConfig_notExisting")))
		})

		It("merges config and file when file is defined in extension", func() {
			mtaPath := getTestPath("mtaConfig.yaml")
			mtaExtPath := getTestPath("resourceConfig_withFile.mtaext")
			resourceConfig, messages, err := GetResourceConfig(mtaPath, []string{mtaExtPath}, "resourceWithConfig", "")
			Ω(err).Should(Succeed())
			Ω(resourceConfig).ShouldNot(BeNil())
			result := map[string]interface{}{
				"xsappname":     "testName",
				"paramFromPath": "paramValueFromPath",
				"deepParam": map[string]interface{}{
					"otherValue": "deepValueFromPath",
				},
			}
			Ω(resourceConfig).Should(Equal(result))
			Ω(messages).Should(BeEmpty())
		})

		It("merges config and file when file is defined in extension and extension is partially invalid", func() {
			mtaPath := getTestPath("mtaConfig.yaml")
			mtaExtPath := getTestPath("resourceConfig_withFile.mtaext")
			mtaExtPath2 := getTestPath("resourceConfig_invalid2.mtaext")
			mtaExtPath3 := getTestPath("resourceConfig_withFile2.mtaext")
			resourceConfig, messages, err := GetResourceConfig(mtaPath, []string{mtaExtPath, mtaExtPath2, mtaExtPath3}, "resourceWithConfig", "")
			Ω(err).Should(Succeed())
			Ω(resourceConfig).ShouldNot(BeNil())
			result := map[string]interface{}{
				"xsappname":     "testNameFromExt",
				"paramFromPath": "paramValueFromPath",
				"deepParam": map[string]interface{}{
					"otherValue": "deepValueFromPath",
				},
			}
			Ω(resourceConfig).Should(Equal(result))
			Ω(messages).Should(ContainElement(ContainSubstring("resourceWithConfig_notExisting")))
		})
	})

	var _ = Describe("GetMtaID", func() {
		It("Get MTA ID", func() {
			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())
			Ω(mtaPath).Should(BeAnExistingFile())

			id, messages, err := GetMtaID(mtaPath)
			Ω(err).Should(Succeed())
			oMtaInput := getMtaInput()
			Ω(id).Should(Equal(oMtaInput.ID))
			Ω(messages).Should(BeEmpty())
		})

		It("Get MTA ID from a non existing mta.yaml file", func() {
			mtaPath := getTestPath("result", "mta.yaml")
			id, messages, err := GetMtaID(mtaPath)
			Ω(err).Should(HaveOccurred())
			Ω(id).Should(Equal(""))
			Ω(messages).Should(BeEmpty())
		})
	})

	var _ = Describe("IsNameUnique", func() {
		It("Check if name exists in mta.yaml", func() {
			mtaPath := getTestPath("mta.yaml")

			//verify module name exists
			exists, messages, err := IsNameUnique(mtaPath, "backend")
			Ω(err).Should(Succeed())
			Ω(exists).Should(BeTrue())
			Ω(messages).Should(BeEmpty())

			//verify provides name exists
			exists, messages, err = IsNameUnique(mtaPath, "backend_task")
			Ω(err).Should(Succeed())
			Ω(exists).Should(BeTrue())
			Ω(messages).Should(BeEmpty())

			//verify resource name exists
			exists, messages, err = IsNameUnique(mtaPath, "database")
			Ω(err).Should(Succeed())
			Ω(exists).Should(BeTrue())
			Ω(messages).Should(BeEmpty())

			//verify random name doesn't exist
			exists, messages, err = IsNameUnique(mtaPath, "blablabla")
			Ω(err).Should(Succeed())
			Ω(exists).ShouldNot(BeTrue())
			Ω(messages).Should(BeEmpty())
		})

		It("Check if name exists in a non existing mta.yaml file ", func() {
			mtaPath := getTestPath("result", "mta.yaml")
			_, _, err := IsNameUnique(mtaPath, oModule.Name)
			Ω(err).Should(HaveOccurred())
		})
	})

	var _ = Describe("ModifyMta", func() {
		It("ModifyMta fails when it cannot create the directory", func() {
			mtaPath := getTestPath("result", "mta.yaml")
			_, _, err := ModifyMta(mtaPath, func() ([]string, error) {
				return nil, nil
			}, 0, false, true, func(s string, mode os.FileMode) error {
				return errors.New("cannot create directory")
			})
			Ω(err).Should(MatchError("cannot create directory"))
		})

		It("ModifyMta creates the directory for new MTA", func() {
			mtaPath := getTestPath("result", "mta.yaml")
			_, _, err := ModifyMta(mtaPath, func() ([]string, error) {
				return nil, nil
			}, 0, false, true, os.MkdirAll)
			Ω(err).Should(Succeed())
			Ω(getTestPath("result")).Should(BeAnExistingFile())
		})

		It("ModifyMta returns an error that the file is locked when lock file exists", func() {
			mtaPath := getTestPath("result", "mta.yaml")
			err := os.MkdirAll(mtaPath, os.ModePerm)

			Ω(err).Should(Succeed())
			lockFilePath := getTestPath("result", "mta-lock.lock")
			file, err := os.OpenFile(lockFilePath, os.O_RDONLY|os.O_CREATE|os.O_EXCL, 0666)
			Ω(err).Should(Succeed())
			_ = file.Close()
			_, _, err = ModifyMta(mtaPath, func() ([]string, error) {
				return nil, nil
			}, 0, false, true, os.MkdirAll)
			Ω(err).Should(MatchError(ContainSubstring("it is locked by another process")))
		})

		It("ModifyMta returns an error that the file cannot be locked when it cannot create the lock file", func() {
			mtaPath := getTestPath("result", "mta.yaml")
			_, _, err := ModifyMta(mtaPath, func() ([]string, error) {
				return nil, nil
			}, 0, false, true, func(s string, mode os.FileMode) error {
				return nil
			})
			Ω(err).Should(MatchError(ContainSubstring("could not lock")))
		})
	})

	var _ = Describe("GetBuildParameters", func() {
		It("Get build parameters", func() {
			myBuilder := ProjectBuilder{
				Builder: "mybuilder",
			}
			otherBuilder := ProjectBuilder{
				Builder: "otherbuilder",
			}
			oBuildParameters := ProjectBuild{
				BeforeAll: []ProjectBuilder{myBuilder},
				AfterAll:  []ProjectBuilder{otherBuilder},
			}

			mtaPath := getTestPath("mta.yaml")

			buildParameters, messages, err := GetBuildParameters(mtaPath, nil)
			Ω(err).Should(Succeed())
			Ω(*buildParameters).Should(Equal(oBuildParameters))
			Ω(messages).Should(BeEmpty())
		})

		It("Get build parameters using extension file", func() {
			myBuilder := ProjectBuilder{
				Builder: "mybuilder",
			}
			otherBuilder := ProjectBuilder{
				Builder: "otherbuilder",
			}
			oBuildParameters := ProjectBuild{
				BeforeAll: []ProjectBuilder{myBuilder},
				AfterAll:  []ProjectBuilder{otherBuilder},
			}

			mtaPath := getTestPath("mta.yaml")
			extPath := getTestPath("mta.mtaext")

			// build parameters are invalid in the mtaext and we don't merge them
			buildParameters, messages, err := GetBuildParameters(mtaPath, []string{extPath})
			Ω(err).Should(Succeed())
			Ω(*buildParameters).Should(Equal(oBuildParameters))
			Ω(messages).Should(BeEmpty())
		})

		It("Get build parameters in a non existing mta.yaml file", func() {
			mtaPath := getTestPath("result", "mta.yaml")
			_, _, err := GetBuildParameters(mtaPath, nil)
			Ω(err).Should(HaveOccurred())
		})
	})

	var _ = Describe("GetParameters", func() {
		It("Get parameters", func() {
			oParameters := map[string]interface{}{"deployer-version": ">=1.2.0"}
			mtaPath := getTestPath("mta.yaml")

			parameters, messages, err := GetParameters(mtaPath, nil)
			Ω(err).Should(Succeed())
			Ω(*parameters).Should(Equal(oParameters))
			Ω(messages).Should(BeEmpty())
		})

		It("Get parameters with extensions", func() {
			oParameters := map[string]interface{}{"deployer-version": "1.2.0", "param1": "ext_param"}
			mtaPath := getTestPath("mta.yaml")
			extPath := getTestPath("mta.mtaext")

			parameters, messages, err := GetParameters(mtaPath, []string{extPath})
			Ω(err).Should(Succeed())
			Ω(*parameters).Should(Equal(oParameters))
			Ω(messages).Should(BeEmpty())
		})

		It("Get parameters in a non existing mta.yaml file", func() {
			mtaPath := getTestPath("result", "mta.yaml")
			_, _, err := GetParameters(mtaPath, nil)
			Ω(err).Should(HaveOccurred())
		})
	})

	var _ = Describe("UpdateBuildParameters", func() {
		mybuilder := ProjectBuilder{
			Builder:  "mybuilder",
			Commands: []string{"abc"},
		}
		builders := []ProjectBuilder{mybuilder}
		projectBuild := ProjectBuild{
			BeforeAll: builders,
		}

		It("Add new build parameters", func() {
			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonBuildParametersData, err := json.Marshal(&projectBuild)
			Ω(err).Should(Succeed())
			messages, err := UpdateBuildParameters(mtaPath, string(jsonBuildParametersData))
			Ω(err).Should(Succeed())
			Ω(messages).Should(BeEmpty())

			oMtaInput := getMtaInput()
			oMtaInput.BuildParams = &projectBuild
			Ω(mtaPath).Should(BeAnExistingFile())
			yamlData, err := ioutil.ReadFile(mtaPath)
			Ω(err).Should(Succeed())
			oMtaOutput, err := Unmarshal(yamlData)
			Ω(err).Should(Succeed())
			Ω(reflect.DeepEqual(oMtaInput, *oMtaOutput)).Should(BeTrue())
		})

		It("Update existing build parameters", func() {
			otherbuilder := ProjectBuilder{
				Builder:  "otherbuilder",
				Commands: []string{"def"},
			}
			updateBuilders := []ProjectBuilder{otherbuilder}
			updateProjectBuild := ProjectBuild{
				BeforeAll: updateBuilders,
			}

			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonBuildParametersData, err := json.Marshal(&projectBuild)
			Ω(err).Should(Succeed())
			messages, err := UpdateBuildParameters(mtaPath, string(jsonBuildParametersData))
			Ω(err).Should(Succeed())
			Ω(messages).Should(BeEmpty())

			jsonUpdateBuildParametersData, err := json.Marshal(&updateProjectBuild)
			Ω(err).Should(Succeed())
			messages, err = UpdateBuildParameters(mtaPath, string(jsonUpdateBuildParametersData))
			Ω(err).Should(Succeed())
			Ω(messages).Should(BeEmpty())

			oMtaInput := getMtaInput()
			oMtaInput.BuildParams = &updateProjectBuild
			Ω(mtaPath).Should(BeAnExistingFile())
			yamlData, err := ioutil.ReadFile(mtaPath)
			Ω(err).Should(Succeed())
			oMtaOutput, err := Unmarshal(yamlData)
			Ω(err).Should(Succeed())
			Ω(reflect.DeepEqual(oMtaInput, *oMtaOutput)).Should(BeTrue())
		})

		It("Update build parameters in a non existing mta.yaml file", func() {
			json := "{name:fff}"
			mtaPath := getTestPath("result", "mta.yaml")
			_, err := UpdateBuildParameters(mtaPath, json)
			Ω(err).Should(HaveOccurred())
		})

		It("Update build parameters with bad json format", func() {
			wrongJSON := "{name:fff"

			mtaPath := getTestPath("result", "temp.mta.yaml")
			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())
			_, err = UpdateBuildParameters(mtaPath, wrongJSON)
			Ω(err).Should(HaveOccurred())
		})
	})

	var _ = Describe("UpdateParameters", func() {

		parameters := map[string]interface{}{"param1": "value1", "param2": "value2"}

		It("Add new parameters", func() {
			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonParametersData, err := json.Marshal(&parameters)
			Ω(err).Should(Succeed())
			messages, err := UpdateParameters(mtaPath, string(jsonParametersData))
			Ω(err).Should(Succeed())
			Ω(messages).Should(BeEmpty())

			oMtaInput := getMtaInput()
			oMtaInput.Parameters = parameters
			Ω(mtaPath).Should(BeAnExistingFile())
			yamlData, err := ioutil.ReadFile(mtaPath)
			Ω(err).Should(Succeed())
			oMtaOutput, err := Unmarshal(yamlData)
			Ω(err).Should(Succeed())
			Ω(oMtaInput).Should(Equal(*oMtaOutput))
		})

		It("Update existing parameters", func() {
			updatedParameters := map[string]interface{}{"param1": "value1", "param3": "value3"}

			mtaPath := getTestPath("result", "temp.mta.yaml")

			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())

			jsonUpdateParametersData, err := json.Marshal(&updatedParameters)
			Ω(err).Should(Succeed())
			messages, err := UpdateParameters(mtaPath, string(jsonUpdateParametersData))
			Ω(err).Should(Succeed())
			Ω(messages).Should(BeEmpty())

			oMtaInput := getMtaInput()
			oMtaInput.Parameters = updatedParameters
			Ω(mtaPath).Should(BeAnExistingFile())
			yamlData, err := ioutil.ReadFile(mtaPath)
			Ω(err).Should(Succeed())
			oMtaOutput, err := Unmarshal(yamlData)
			Ω(err).Should(Succeed())
			Ω(oMtaInput).Should(Equal(*oMtaOutput))
		})

		It("Update parameters in a non existing mta.yaml file", func() {
			json := "{param:value}"
			mtaPath := getTestPath("result", "mta.yaml")
			_, err := UpdateParameters(mtaPath, json)
			Ω(err).Should(HaveOccurred())
		})

		It("Update parameters with bad json format", func() {
			wrongJSON := "{param:value"

			mtaPath := getTestPath("result", "temp.mta.yaml")
			jsonRootData, err := json.Marshal(getMtaInput())
			Ω(err).Should(Succeed())
			Ω(CreateMta(mtaPath, string(jsonRootData), os.MkdirAll)).Should(Succeed())
			_, err = UpdateParameters(mtaPath, wrongJSON)
			Ω(err).Should(HaveOccurred())
		})
	})
})

var _ = Describe("Locking", func() {
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
		err := os.MkdirAll(getTestPath("result"), os.ModePerm)
		Ω(err).Should(Succeed())
		mtaPath := getTestPath("result", "mta.yaml")
		Ω(CopyFile(getTestPath("mta.yaml"), mtaPath, os.Create)).Should(Succeed())

		mtaHashCode, exists, err := GetMtaHash(mtaPath)
		Ω(err).Should(Succeed())
		Ω(exists).Should(BeTrue())

		jsonData, err := json.Marshal(oModule)
		Ω(err).Should(Succeed())
		moduleJSON := string(jsonData)
		mtaHashCodeResult, messages, err := ModifyMta(mtaPath, func() ([]string, error) {
			return AddModule(mtaPath, moduleJSON, Marshal)
		}, mtaHashCode, false, false, os.MkdirAll)
		Ω(err).Should(Succeed())
		Ω(messages).Should(BeEmpty())
		Ω(mtaHashCodeResult).ShouldNot(Equal(mtaHashCode))
		mtaHashCodeAfterModify, _, err := GetMtaHash(mtaPath)
		Ω(err).Should(Succeed())
		Ω(mtaHashCodeResult).Should(Equal(mtaHashCodeAfterModify))
		// wrong yaml
		_, _, err = ModifyMta(getTestPath("result", "mtaX.yaml"), func() ([]string, error) {
			return AddModule(getTestPath("result", "mtaX.yaml"), moduleJSON, Marshal)
		}, mtaHashCode, false, false, os.MkdirAll)
		Ω(err).Should(HaveOccurred())
		Ω(err.Error()).Should(ContainSubstring("file does not exist"))
		oModule.Name = "test1"
		jsonData, err = json.Marshal(oModule)
		Ω(err).Should(Succeed())
		moduleJSON = string(jsonData)
		// hashcode of the mta.yaml is wrong now
		_, _, err = ModifyMta(mtaPath, func() ([]string, error) {
			return AddModule(mtaPath, moduleJSON, Marshal)
		}, mtaHashCode, false, false, os.MkdirAll)
		Ω(err).Should(HaveOccurred())
	})

	It("Modify mta.yaml with force even when the hashcode is wrong", func() {
		err := os.MkdirAll(getTestPath("result"), os.ModePerm)
		Ω(err).Should(Succeed())
		mtaPath := getTestPath("result", "mta.yaml")
		Ω(CopyFile(getTestPath("mta.yaml"), mtaPath, os.Create)).Should(Succeed())

		mtaHashCode, exists, err := GetMtaHash(mtaPath)
		Ω(err).Should(Succeed())
		Ω(exists).Should(BeTrue())

		jsonData, err := json.Marshal(oModule)
		Ω(err).Should(Succeed())
		moduleJSON := string(jsonData)
		mtaHashCodeResult, messages, err := ModifyMta(mtaPath, func() ([]string, error) {
			return AddModule(mtaPath, moduleJSON, Marshal)
		}, mtaHashCode, false, false, os.MkdirAll)
		Ω(err).Should(Succeed())
		Ω(messages).Should(BeEmpty())
		Ω(mtaHashCodeResult).ShouldNot(Equal(mtaHashCode))
		mtaHashCodeAfterModify, _, err := GetMtaHash(mtaPath)
		Ω(err).Should(Succeed())
		Ω(mtaHashCodeResult).Should(Equal(mtaHashCodeAfterModify))
		// hashcode of the mta.yaml is wrong now but force is true
		_, _, err = ModifyMta(mtaPath, func() ([]string, error) {
			return AddModule(mtaPath, moduleJSON, Marshal)
		}, mtaHashCode, true, false, os.MkdirAll)
		Ω(err).Should(Succeed())
	})

	It("2 parallel processes, second fails to make locking", func() {
		err := os.MkdirAll(getTestPath("result"), os.ModePerm)
		Ω(err).Should(Succeed())
		mtaPath := getTestPath("result", "mta.yaml")
		Ω(CopyFile(getTestPath("mta.yaml"), mtaPath, os.Create)).Should(Succeed())
		mtaHashCode, _, err := GetMtaHash(mtaPath)
		Ω(err).Should(Succeed())
		var wg sync.WaitGroup
		wg.Add(1)
		var err1 error
		go func() {
			_, _, err1 = ModifyMta(mtaPath, func() ([]string, error) {
				time.Sleep(time.Second)
				return nil, nil
			}, mtaHashCode, false, false, os.MkdirAll)
			defer wg.Done()
		}()
		time.Sleep(time.Millisecond * 200)
		wg.Add(1)
		var err2 error
		go func() {
			_, _, err2 = ModifyMta(mtaPath, func() ([]string, error) {
				time.Sleep(time.Second)
				return nil, nil
			}, mtaHashCode, false, false, os.MkdirAll)
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

	It("RunModifyAndWriteHash performs the action and writes the messages and hashcode to the output when there is no error", func() {
		err := os.MkdirAll(getTestPath("result"), os.ModePerm)
		Ω(err).Should(Succeed())
		mtaPath := getTestPath("result", "temp.mta.yaml")
		json, err := json.Marshal(getMtaInput())
		Ω(err).Should(Succeed())
		output := executeAndProvideOutput(func() {
			err = RunModifyAndWriteHash("info message", mtaPath, false, func() ([]string, error) {
				return []string{"some message"}, CreateMta(mtaPath, string(json), os.MkdirAll)
			}, 0, true)
			Ω(err).Should(Succeed())
		})
		// Check the last line of the result is a json with messages and hashcode and that the hashcode is not 0
		Ω(output).Should(MatchRegexp(`{"messages":\["some message"\],"hashcode":[1-9][0-9]*}$`))
		// Note: the info message is written to the logger but we don't test it because the logger is initialized
		// with stdout before it's replaced in the test
	})

	It("RunModifyAndWriteHash performs the action and writes the hashcode to the output when there is no error and messages", func() {
		err := os.MkdirAll(getTestPath("result"), os.ModePerm)
		Ω(err).Should(Succeed())
		mtaPath := getTestPath("result", "temp.mta.yaml")
		json, err := json.Marshal(getMtaInput())
		Ω(err).Should(Succeed())
		output := executeAndProvideOutput(func() {
			err = RunModifyAndWriteHash("info message", mtaPath, false, func() ([]string, error) {
				return nil, CreateMta(mtaPath, string(json), os.MkdirAll)
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
			err := RunModifyAndWriteHash("info message", mtaPath, false, func() ([]string, error) {
				return []string{"some warning"}, errors.New("some error")
			}, 0, true)
			Ω(err).Should(MatchError("some error"))
		})
		// Check the last line of the result is a json with hashcode and that it's is not 0
		Ω(output).Should(MatchRegexp(`{"message":"some error"}$`))
		// Note: the info message is written to the logger but we don't test it because the logger is initialized
		// with stdout before it's replaced in the test
	})

	It("RunAndWriteResultAndHash performs the action and writes the hashcode, messages and result to the output when there is no error", func() {
		err := os.MkdirAll(getTestPath("result"), os.ModePerm)
		Ω(err).Should(Succeed())
		mtaPath := getTestPath("result", "temp.mta.yaml")
		output := executeAndProvideOutput(func() {
			err = RunAndWriteResultAndHash("info message", mtaPath, nil, func() (interface{}, []string, error) {
				err1 := CopyFile(getTestPath("mta.yaml"), mtaPath, os.Create)
				Ω(err1).Should(Succeed())
				return 1, []string{"some message"}, nil
			})
			Ω(err).Should(Succeed())
		})
		// Check the last line of the result is a json with messages and hashcode and that the hashcode is not 0
		Ω(output).Should(MatchRegexp(`{"result":1,"messages":\["some message"\],"hashcode":[1-9][0-9]*}$`))
		// Note: the info message is written to the logger but we don't test it because the logger is initialized
		// with stdout before it's replaced in the test
	})

	It("RunAndWriteResultAndHash performs the action and writes the hashcode and result to the output when there is no error and messages", func() {
		err := os.MkdirAll(getTestPath("result"), os.ModePerm)
		Ω(err).Should(Succeed())
		mtaPath := getTestPath("result", "temp.mta.yaml")
		output := executeAndProvideOutput(func() {
			err = RunAndWriteResultAndHash("info message", mtaPath, nil, func() (interface{}, []string, error) {
				err1 := CopyFile(getTestPath("mta.yaml"), mtaPath, os.Create)
				Ω(err1).Should(Succeed())
				return 1, nil, nil
			})
			Ω(err).Should(Succeed())
		})
		// Check the last line of the result is a json with hashcode and that it's not 0
		Ω(output).Should(MatchRegexp(`{"result":1,"hashcode":[1-9][0-9]*}$`))
		// Note: the info message is written to the logger but we don't test it because the logger is initialized
		// with stdout before it's replaced in the test
	})

	It("RunAndWriteResultAndHash performs the action and writes result with hashcode 0 to the output when there are extensions", func() {
		err := os.MkdirAll(getTestPath("result"), os.ModePerm)
		Ω(err).Should(Succeed())
		mtaPath := getTestPath("result", "temp.mta.yaml")
		output := executeAndProvideOutput(func() {
			err = RunAndWriteResultAndHash("info message", mtaPath, []string{"some ext"}, func() (interface{}, []string, error) {
				err1 := CopyFile(getTestPath("mta.yaml"), mtaPath, os.Create)
				Ω(err1).Should(Succeed())
				return 1, []string{"some message"}, nil
			})
			Ω(err).Should(Succeed())
		})
		// Check the last line of the result is a json with hashcode 0
		Ω(output).Should(Equal(`{"result":1,"messages":["some message"],"hashcode":0}`))
	})

	It("RunAndWriteResultAndHash writes the error when the action fails", func() {
		err := os.MkdirAll(getTestPath("result"), os.ModePerm)
		Ω(err).Should(Succeed())
		mtaPath := getTestPath("result", "temp.mta.yaml")
		output := executeAndProvideOutput(func() {
			err := RunAndWriteResultAndHash("info message", mtaPath, nil, func() (interface{}, []string, error) {
				return nil, []string{"some message"}, errors.New("some error")
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

func jsonMarshalErr(o interface{}) ([]byte, error) {
	return nil, errors.New("could not marshal to json")
}

type UnmarshalableString string

func (s UnmarshalableString) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal value %s", string(s))
}

// executeAndProvideOutput runs the execute function in a goroutine and returns the output written to os.Stdout
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
