package commands

import (
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/SAP/cloud-mta/mta"
)

var _ = Describe("Module", func() {
	It("Sanity", func() {
		os.MkdirAll(getTestPath("result"), os.ModePerm)
		addModuleMtaCmdPath = getTestPath("result", "mta.yaml")
		Ω(mta.CopyFile(getTestPath("mta.yaml"), addModuleMtaCmdPath, os.Create)).Should(Succeed())

		var err error
		addModuleCmdHashcode, err = mta.GetMtaHash(addModuleMtaCmdPath)
		Ω(err).Should(Succeed())
		oModule := mta.Module{
			Name: "testModule",
			Type: "testType",
			Path: "test",
		}

		jsonData, err := json.Marshal(oModule)
		addModuleCmdData = string(jsonData)
		Ω(addModuleCmd.RunE(nil, []string{})).Should(Succeed())
		oModule.Name = "test1"
		jsonData, err = json.Marshal(oModule)
		addModuleCmdData = string(jsonData)
		// hashcode of the mta.yaml is wrong now
		Ω(addModuleCmd.RunE(nil, []string{})).Should(HaveOccurred())
	})
})
