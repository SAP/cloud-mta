package mta

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"

	"github.com/SAP/cloud-mta/internal/fs"
)

func createMtaYamlFile(path string, mkDirs func(string, os.FileMode) error) (err error) {
	folder := filepath.Dir(path)
	err = mkDirs(folder, os.ModePerm)
	if err != nil {
		return err
	}
	file, err := fs.CreateFile(path)
	defer func() {
		err = file.Close()
	}()

	return
}

func getMtaFromFile(path string) (*MTA, error) {
	mtaContent, err := ioutil.ReadFile(filepath.Join(path))
	if err != nil {
		return nil, errors.Wrapf(err, `addition failed when reading %s file`, path)
	}
	return Unmarshal(mtaContent)
}

func unmarshalData(dataJSON string, mta *MTA, o interface{}) error {
	dataYaml, err := yaml.JSONToYAML([]byte(dataJSON))
	if err != nil {
		return err
	}
	return yaml.Unmarshal(dataYaml, &o)
}

func saveMTA(path string, mta *MTA, marshal func(interface{}) ([]byte, error)) error {
	mtaBytes, err := marshal(mta)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, mtaBytes, 0644)
}

// CreateMta - create MTA project
func CreateMta(path string, mtaDataJSON string, mkDirs func(string, os.FileMode) error) error {
	mtaDataYaml, err := yaml.JSONToYAML([]byte(mtaDataJSON))
	if err != nil {
		return err
	}
	err = createMtaYamlFile(filepath.Join(path), mkDirs)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, mtaDataYaml, 0644)
}

//AddModule - add new module
func AddModule(path string, moduleDataJSON string, marshal func(interface{}) ([]byte, error)) error {
	mta, err := getMtaFromFile(path)
	if err != nil {
		return err
	}

	module := Module{}
	err = unmarshalData(moduleDataJSON, mta, &module)
	if err != nil {
		return err
	}

	mta.Modules = append(mta.Modules, &module)
	return saveMTA(path, mta, marshal)
}

//AddResource - add new resource
func AddResource(path string, resourceDataJSON string, marshal func(interface{}) ([]byte, error)) error {
	mta, err := getMtaFromFile(path)
	if err != nil {
		return err
	}

	resource := Resource{}
	err = unmarshalData(resourceDataJSON, mta, &resource)
	if err != nil {
		return err
	}

	mta.Resources = append(mta.Resources, &resource)
	return saveMTA(path, mta, marshal)
}

// CopyFile - copy from source path to target path
func CopyFile(src, dst string, create func(string) (*os.File, error)) (rerr error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		e := in.Close()
		if rerr == nil {
			rerr = e
		}
	}()

	out, err := create(dst)
	if err != nil {
		return err
	}
	defer func() {
		e := out.Close()
		if rerr == nil {
			rerr = e
		}
	}()

	_, err = io.Copy(out, in)
	return err
}

// DeleteFile - delete file
func DeleteFile(path string) error {
	return fs.DeleteFile(path)
}
