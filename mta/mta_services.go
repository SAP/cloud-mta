package mta

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"

	"github.com/SAP/cloud-mta/internal/fs"
)

func createMtaYamlFile(path string) (err error) {
	folder := filepath.Dir(path)
	err = os.MkdirAll(folder, os.ModePerm)
	if err != nil {
		return err
	}
	file, err := fs.CreateFile(path)
	defer func() {
		err = file.Close()
	}()

	return
}

// CreateMta - create MTA project
func CreateMta(path string, mtaDataJson string) error {
	mtaDataYaml, err := yaml.JSONToYAML([]byte(mtaDataJson))
	if err != nil {
		return err
	}
	err = createMtaYamlFile(filepath.Join(path))
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, mtaDataYaml, 0644)
}
