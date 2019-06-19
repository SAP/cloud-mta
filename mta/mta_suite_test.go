package mta

import (
	"github.com/SAP/cloud-mta/internal/logs"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMta(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mta Suite")
}

var _ = BeforeSuite(func() {
	logs.Logger = logs.NewLogger()
})

func getTestPath(relPath ...string) string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "testdata", filepath.Join(relPath...))
}
