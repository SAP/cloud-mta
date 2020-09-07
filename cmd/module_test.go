package commands

import (
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/SAP/cloud-mta/mta"
)

var _ = Describe("Module", func() {

	AfterEach(func() {
		os.RemoveAll(getTestPath("result"))
	})

	It("Sanity", func() {
		err := os.MkdirAll(getTestPath("result"), os.ModePerm)
		Ω(err).Should(Succeed())
		addModuleMtaCmdPath = getTestPath("result", "mta.yaml")
		Ω(mta.CopyFile(getTestPath("mta.yaml"), addModuleMtaCmdPath, os.Create)).Should(Succeed())

		hash, exists, err := mta.GetMtaHash(addModuleMtaCmdPath)
		addModuleCmdHashcode = hash
		Ω(err).Should(Succeed())
		Ω(exists).Should(BeTrue())
		oModule := mta.Module{
			Name: "testModule",
			Type: "testType",
			Path: "test",
		}

		jsonData, err := json.Marshal(oModule)
		Ω(err).Should(Succeed())
		addModuleCmdData = string(jsonData)
		Ω(addModuleCmd.RunE(nil, []string{})).Should(Succeed())
		oModule.Name = "test1"
		jsonData, err = json.Marshal(oModule)
		Ω(err).Should(Succeed())
		addModuleCmdData = string(jsonData)
		// hashcode of the mta.yaml is wrong now
		Ω(addModuleCmd.RunE(nil, []string{})).Should(HaveOccurred())
	})
})
