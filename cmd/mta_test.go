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
		os.MkdirAll(getTestPath("result"), os.ModePerm)
		createMtaCmdPath = getTestPath("result", "mta.yaml")

		var err error

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
		createMtaCmdData = string(jsonData)
		Ω(createMtaCmd.RunE(nil, []string{})).Should(Succeed())
		// already exists
		Ω(createMtaCmd.RunE(nil, []string{})).Should(HaveOccurred())
	})
})
