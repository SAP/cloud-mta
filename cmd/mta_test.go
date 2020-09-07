package commands

import (
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/SAP/cloud-mta/mta"
)

var _ = Describe("Resource", func() {

	AfterEach(func() {
		os.RemoveAll(getTestPath("result"))
	})

	It("Sanity", func() {
		err := os.MkdirAll(getTestPath("result"), os.ModePerm)
		Ω(err).Should(Succeed())
		createMtaCmdPath = getTestPath("result", "mta.yaml")

		hash, exists, err := mta.GetMtaHash(createMtaCmdPath)
		Ω(hash).Should(Equal(0))
		Ω(err).Should(Succeed())
		Ω(exists).Should(BeFalse())

		schemaVersion := "1.0"
		oMta := mta.MTA{
			ID:            "abc",
			SchemaVersion: &schemaVersion,
			Version:       "1.1",
		}

		jsonData, err := json.Marshal(&oMta)
		Ω(err).Should(Succeed())
		createMtaCmdData = string(jsonData)
		Ω(createMtaCmd.RunE(nil, []string{})).Should(Succeed())
		// already exists
		Ω(createMtaCmd.RunE(nil, []string{})).Should(HaveOccurred())
		deleteMtaCmdPath = getTestPath("result")
		Ω(deleteMtaCmd.RunE(nil, []string{})).Should(Succeed())
	})
})
