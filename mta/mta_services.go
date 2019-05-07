package mta

import (
	"io"
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
func CreateMta(path string, mtaDataJSON string) error {
	mtaDataYaml, err := yaml.JSONToYAML([]byte(mtaDataJSON))
	if err != nil {
		return err
	}
	err = createMtaYamlFile(filepath.Join(path))
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, mtaDataYaml, 0644)
}

//AddModule - add new module
func AddModule(path string, moduleDataJSON string) error {
	mtaContent, err := ioutil.ReadFile(filepath.Join(path))
	if err != nil {
		return err
	}

	mtaObj, err := Unmarshal(mtaContent)
	if err != nil {
		return err
	}

	moduleDataYaml, err := yaml.JSONToYAML([]byte(moduleDataJSON))
	if err != nil {
		return err
	}

	module := Module{}
	err = yaml.Unmarshal(moduleDataYaml, &module)
	if err != nil {
		return err
	}

	mtaObjModules := append(mtaObj.Modules, &module)
	mtaObj.Modules = mtaObjModules
	mtaBytes, _ := yaml.Marshal(mtaObj)
	return ioutil.WriteFile(path, mtaBytes, 0644)
}

//AddResource - add new resource
func AddResource(path string, resourceDataJSON string) error {
	mtaContent, err := ioutil.ReadFile(filepath.Join(path))
	if err != nil {
		return err
	}

	mtaObj, err := Unmarshal(mtaContent)
	if err != nil {
		return err
	}

	resourceDataYaml, err := yaml.JSONToYAML([]byte(resourceDataJSON))
	if err != nil {
		return err
	}

	resource := Resource{}
	err = yaml.Unmarshal(resourceDataYaml, &resource)
	if err != nil {
		return err
	}

	mtaObjResources := append(mtaObj.Resources, &resource)
	mtaObj.Resources = mtaObjResources
	mtaBytes, _ := yaml.Marshal(mtaObj)
	return ioutil.WriteFile(path, mtaBytes, 0644)
}

// CopyFile - copy from source path to target path
func CopyFile(src, dst string) (rerr error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		rerr = in.Close()
	}()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		rerr = out.Close()
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return err
}

// DeleteFile - delete file
func DeleteFile(path string) error {
	return fs.DeleteFile(path)
}
