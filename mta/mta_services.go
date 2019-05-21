package mta

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/SAP/cloud-mta/internal/logs"
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

func saveMTA(path string, mta *MTA, marshal func(*MTA) ([]byte, error)) error {
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
func AddModule(path string, moduleDataJSON string, marshal func(*MTA) ([]byte, error)) error {
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
func AddResource(path string, resourceDataJSON string, marshal func(*MTA) ([]byte, error)) error {
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
func UpdateModule(path string, moduleDataJSON string, marshal func(*MTA) ([]byte, error)) error {
	mtaObj, err := getMtaFromFile(path)
	if err != nil {
		return err
	}

	module := Module{}
	err = unmarshalData(moduleDataJSON, &module)
	if err != nil {
		return err
	}

	// Replace the first existing module with the same name
	for index, existingModule := range mtaObj.Modules {
		if existingModule.Name == module.Name {
			mtaObj.Modules[index] = &module
			return saveMTA(path, mtaObj, marshal)
		}
	}

	return fmt.Errorf("module with name %s does not exist", module.Name)
}

// UpdateResource updates an existing resource according to the resource name. In case more than one resource with this
// name exists, one of the resources is updated to the existing structure.
func UpdateResource(path string, resourceDataJSON string, marshal func(*MTA) ([]byte, error)) error {
	mtaObj, err := getMtaFromFile(path)
	if err != nil {
		return err
	}

	resource := Resource{}
	err = unmarshalData(resourceDataJSON, &resource)
	if err != nil {
		return err
	}

	// Replace the first existing resource with the same name
	for index, existingResource := range mtaObj.Resources {
		if existingResource.Name == resource.Name {
			mtaObj.Resources[index] = &resource
			return saveMTA(path, mtaObj, marshal)
		}
	}

	return fmt.Errorf("resource with name %s does not exist", resource.Name)
}

//IsNameUnique - check if name already exists as module/resource/provide name
func IsNameUnique(path string, name string) (bool, error) {
	mta, err := getMtaFromFile(path)
	if err != nil {
		return true, err
	}

	for _, module := range mta.Modules {
		if name == module.Name {
			return true, nil
		}
		for _, provide := range module.Provides {
			if name == provide.Name {
				return true, nil
			}
		}
	}
	for _, resource := range mta.Resources {
		if name == resource.Name {
			return true, nil
		}
	}
	return false, nil
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
func ModifyMta(path string, modify func() error, hashcode int, isNew bool) (newHashcode int, rerr error) {
	// create lock file
	lockFilePath := filepath.Join(filepath.Dir(path), "mta-lock.lock")
	file, err := os.OpenFile(lockFilePath, os.O_RDONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return 0, fmt.Errorf(`could not modify the "%s" file; it is locked by another process`, path)
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
	if err != nil {
		return 0, err
	}
	newHashcode, _, err = GetMtaHash(path)
	return newHashcode, err
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

type outputResult struct {
	Result   interface{} `json:"result,omitempty"`
	Hashcode int         `json:"hashcode"`
}
type outputError struct {
	Message string `json:"message"`
}

// WriteResult writes the result of an operation to the output in JSON format. In case of an error
// the message is written. In case of success the hashcode and results are written.
func WriteResult(result interface{}, hashcode int, err error) error {
	return PrintResult(result, hashcode, err, fmt.Print)
}

// PrintResult calls the sent print function on the output in JSON format. In case of an error
// the message is printed. In case of success the hashcode and results are printed.
func PrintResult(result interface{}, hashcode int, err error, print func(...interface{}) (n int, err error)) error {
	if err != nil {
		outputErr := outputError{err.Error()}
		bytes, err1 := json.Marshal(outputErr)
		if err1 != nil {
			return err1
		}
		_, err1 = print(string(bytes))
		return err1
	}
	output := outputResult{result, hashcode}
	bytes, err := json.Marshal(output)
	if err != nil {
		return err
	}
	_, err = print(string(bytes))
	return err
}

// RunModifyAndWriteHash logs the info, executes the action while locking the mta file in the path, and writes the
// result and hashcode (or error) to the output
func RunModifyAndWriteHash(info string, path string, action func() error, hashcode int, isNew bool) error {
	logs.Logger.Info(info)
	newHashcode, err := ModifyMta(path, action, hashcode, isNew)
	writeErr := WriteResult(nil, newHashcode, err)
	if err != nil {
		// The original error is more important
		return err
	}
	return writeErr
}

// RunAndWriteResultAndHash logs the info, executes the action, and writes the result and hashcode of the mta in the
// path (or error) to the output
func RunAndWriteResultAndHash(info string, path string, action func() (interface{}, error)) error {
	logs.Logger.Info(info)
	result, err := action()
	hashcode := 0
	if err != nil {
		hashcode, _, err = GetMtaHash(path)
	}
	writeErr := WriteResult(result, hashcode, err)
	if err != nil {
		// The original error is more important
		return err
	}
	return writeErr
}
