package mta

import (
	"os"

	"github.com/SAP/cloud-mta/internal/fs"
)

func createMtaYamlFile(path string) (*os.File, error) {
	return fs.CreateFile(path)
}

// CreateMta - create MTA project
func (mta *MTA) CreateMta(path string, data []byte) (*MTA, error) {
	_, err := createMtaYamlFile(path)

	return nil, err
}
