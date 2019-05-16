package mta

import (
	"crypto/sha1"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	ghodss "github.com/ghodss/yaml"
	"github.com/pkg/errors"

	"github.com/SAP/cloud-mta/internal/fs"
)

func createMtaYamlFile(path string, mkDirs func(string, os.FileMode) error) (rerr error) {
	folder := filepath.Dir(path)
	rerr = mkDirs(folder, os.ModePerm)
	if rerr != nil {
		return
	}
	file, rerr := fs.CreateFile(path)
	defer func() {
		if file != nil {
			e := file.Close()
			if rerr == nil {
				rerr = e
			}
		}
	}()

	return
}

func getMtaFromFile(path string) (*MTA, error) {
	mtaContent, err := ioutil.ReadFile(filepath.Join(path))
	if err != nil {
		return nil, errors.Wrapf(err, `failed when reading %s file`, path)
	}
	s := string(mtaContent)
	s = strings.Replace(s, "\r\n", "\r", -1)
	mtaContent = []byte(s)
	return Unmarshal(mtaContent)
}

func unmarshalData(dataJSON string, o interface{}) error {
	dataYaml, err := ghodss.JSONToYAML([]byte(dataJSON))
	if err != nil {
		return err
	}
	return yaml.Unmarshal(dataYaml, o)
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
	mtaDataYaml, err := ghodss.JSONToYAML([]byte(mtaDataJSON))
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
	mta, err := getMtaFromFile(filepath.Join(path))
	if err != nil {
		return err
	}

	module := Module{}
	err = unmarshalData(moduleDataJSON, &module)
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
	err = unmarshalData(resourceDataJSON, &resource)
	if err != nil {
		return err
	}

	mta.Resources = append(mta.Resources, &resource)
	return saveMTA(path, mta, marshal)
}

//GetModules - get all modules
func GetModules(path string) ([]*Module, error) {
	mta, err := getMtaFromFile(path)
	if err != nil {
		return nil, err
	}
	return mta.Modules, nil
}

//GetResources - get all resources
func GetResources(path string) ([]*Resource, error) {
	mta, err := getMtaFromFile(path)
	if err != nil {
		return nil, err
	}
	return mta.Resources, nil
}

// UpdateModule updates an existing module according to the module name. In case more than one module with this
// name exists, one of the modules is updated to the existing structure.
func UpdateModule(path string, moduleDataJSON string, marshal func(interface{}) ([]byte, error)) error {
	mtaContent, err := ioutil.ReadFile(filepath.Join(path))
	if err != nil {
		return err
	}

	mtaObj, err := Unmarshal(mtaContent)
	if err != nil {
		return err
	}

	moduleDataYaml, err := ghodss.JSONToYAML([]byte(moduleDataJSON))
	if err != nil {
		return err
	}

	module := Module{}
	err = yaml.Unmarshal(moduleDataYaml, &module)
	if err != nil {
		return err
	}

	// Replace the first existing module with the same name
	found := false
	for index, existingModule := range mtaObj.Modules {
		if existingModule.Name == module.Name {
			mtaObj.Modules[index] = &module
			found = true
			break
		}
	}
	if !found {
		err = fmt.Errorf("module with name %s does not exist", module.Name)
		return err
	}

	mtaBytes, err := marshal(mtaObj)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, mtaBytes, 0644)
}

// UpdateResource updates an existing resource according to the resource name. In case more than one resource with this
// name exists, one of the resources is updated to the existing structure.
func UpdateResource(path string, resourceDataJSON string, marshal func(interface{}) ([]byte, error)) error {
	mtaContent, err := ioutil.ReadFile(filepath.Join(path))
	if err != nil {
		return err
	}

	mtaObj, err := Unmarshal(mtaContent)
	if err != nil {
		return err
	}

	resourceDataYaml, err := ghodss.JSONToYAML([]byte(resourceDataJSON))
	if err != nil {
		return err
	}

	resource := Resource{}
	err = yaml.Unmarshal(resourceDataYaml, &resource)
	if err != nil {
		return err
	}

	// Replace the first existing resource with the same name
	found := false
	for index, existingResource := range mtaObj.Resources {
		if existingResource.Name == resource.Name {
			mtaObj.Resources[index] = &resource
			found = true
			break
		}
	}
	if !found {
		err = fmt.Errorf("resource with name %s does not exist", resource.Name)
		return err
	}

	mtaBytes, err := marshal(mtaObj)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, mtaBytes, 0644)
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

// GetMtaHash - get hashcode of the mta file
func GetMtaHash(path string) (int, bool, error) {
	mtaContent, err := ioutil.ReadFile(filepath.Join(path))
	if err != nil {
		// file not exists
		return 0, false, nil
	}
	h := sha1.New()
	code, err := h.Write(mtaContent)
	return code, true, err
}

// ModifyMta - lock and modify mta.yaml file
func ModifyMta(path string, modify func() error, hashcode int, isNew bool) (rerr error) {
	// create lock file
	lockFilePath := filepath.Join(filepath.Dir(path), "mta-lock.lock")
	file, err := os.OpenFile(lockFilePath, os.O_RDONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return fmt.Errorf(`could not modify the "%s" file; it is locked by another process`, path)
	}
	// unlock and remove lock file at the end of modification
	defer func() {
		if file == nil {
			return
		}
		e := file.Close()
		if e == nil {
			e = os.Remove(lockFilePath)
		}
		if rerr == nil {
			rerr = e
		}
	}()

	currentHash, exists, err := GetMtaHash(path)

	if err == nil {
		err = ifFileChangeable(path, isNew, exists, currentHash == hashcode)
	}
	if err == nil {
		err = modify()
	}

	return err
}

func ifFileChangeable(path string, isNew, exists, sameHash bool) error {
	if isNew && exists {
		return fmt.Errorf(`could not create the "%s" file; another file with this name already exists`, path)
	} else if !isNew && !exists {
		return fmt.Errorf(`the "%s" file does not exist`, path)
	} else if !sameHash {
		return fmt.Errorf(`could not update the "%s" file; it was modified by another process`, path)
	}
	return nil
}
