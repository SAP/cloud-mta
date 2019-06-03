package mta

import (
	"crypto/sha1"
	"encoding/json"
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
	"github.com/SAP/cloud-mta/internal/logs"
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
		return nil, errors.Wrapf(err, "failed when reading the '%s' file", path)
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

// CreateMta - creates an MTA project.
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

//AddModule - adds a new module.
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

//AddResource - adds a new resource.
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

//GetModules - gets all modules.
func GetModules(path string) ([]*Module, error) {
	mta, err := getMtaFromFile(path)
	if err != nil {
		return nil, err
	}
	return mta.Modules, nil
}

//GetResources - gets all resources.
func GetResources(path string) ([]*Resource, error) {
	mta, err := getMtaFromFile(path)
	if err != nil {
		return nil, err
	}
	return mta.Resources, nil
}

// UpdateModule updates an existing module according to the module name. If more than one module with this
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

	// Replaces the first existing module with the same name.
	for index, existingModule := range mtaObj.Modules {
		if existingModule.Name == module.Name {
			mtaObj.Modules[index] = &module
			return saveMTA(path, mtaObj, marshal)
		}
	}

	return fmt.Errorf("the '%s' module does not exist", module.Name)
}

// UpdateResource updates an existing resource according to the resource name. If more than one resource with this
// name exists, one of the resources is updated in the existing structure.
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

	// Replaces the first existing resource with the same name.
	for index, existingResource := range mtaObj.Resources {
		if existingResource.Name == resource.Name {
			mtaObj.Resources[index] = &resource
			return saveMTA(path, mtaObj, marshal)
		}
	}

	return fmt.Errorf("the '%s' resource does not exist", resource.Name)
}

//IsNameUnique - checks if the name already exists as a `module`/`resource`/`provide` name.
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

//UpdateBuildParameters - updates the MTA build parameters.
func UpdateBuildParameters(path string, buildParamsDataJSON string) error {
	mta, err := getMtaFromFile(path)
	if err != nil {
		return err
	}

	buildParams := ProjectBuild{}
	err = unmarshalData(buildParamsDataJSON, &buildParams)
	if err != nil {
		return err
	}

	mta.BuildParams = &buildParams
	return saveMTA(path, mta, Marshal)
}

// CopyFile - copies a file from the source path to the target path.
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

// DeleteFile - deletes the file.
func DeleteFile(path string) error {
	return fs.DeleteFile(path)
}

// GetMtaHash - gets the hashcode of the MTA file.
func GetMtaHash(path string) (int, bool, error) {
	mtaContent, err := ioutil.ReadFile(filepath.Join(path))
	if err != nil {
		// the file does not exist.
		return 0, false, nil
	}
	h := sha1.New()
	code, err := h.Write(mtaContent)
	return code, true, err
}

// ModifyMta - locks and modifies the "mta.yaml" file.
func ModifyMta(path string, modify func() error, hashcode int, force bool, isNew bool, mkDirs func(string, os.FileMode) error) (newHashcode int, rerr error) {
	// Creates the lock file.
	// Makes sure the directory of the lock file exists (it might not exist if it is a new MTA).
	folder := filepath.Dir(path)
	rerr = mkDirs(folder, os.ModePerm)
	if rerr != nil {
		return 0, rerr
	}
	lockFilePath := filepath.Join(filepath.Dir(path), "mta-lock.lock")
	file, err := os.OpenFile(lockFilePath, os.O_RDONLY|os.O_CREATE|os.O_EXCL, 0666)
	if os.IsExist(err) {
		return 0, fmt.Errorf("could not modify the \"%s\" file; it is locked by another process", path)
	} else if err != nil {
		return 0, fmt.Errorf("could not lock the \"%s\" file for modification; %s", path, err)
	}
	// Unlocks and removes the lock file at the end of modification.
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
		err = ifFileChangeable(path, isNew, exists, currentHash == hashcode, force)
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

func ifFileChangeable(path string, isNew, exists, sameHash bool, force bool) error {
	if isNew && exists {
		return fmt.Errorf("could not create the \"%s\" file; another file with this name already exists", path)
	} else if !isNew && !exists {
		return fmt.Errorf("the \"%s\" file does not exist", path)
	} else if !sameHash && !force {
		return fmt.Errorf("could not update the \"%s\" file; it was modified by another process", path)
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

// WriteResult - writes the result of an operation to the output in JSON format. If successful, the hashcode and results are written; otherwise an error is displayed.
func WriteResult(result interface{}, hashcode int, err error) error {
	return printResult(result, hashcode, err, fmt.Print, json.Marshal)
}

func printResult(result interface{}, hashcode int, err error, print func(...interface{}) (n int, err error), jsonMarshal func(v interface{}) ([]byte, error)) error {
	if err != nil {
		outputErr := outputError{err.Error()}
		bytes, err1 := jsonMarshal(outputErr)
		if err1 != nil {
			_, _ = print("could not marshal error with message " + err.Error() + "; " + err1.Error())
			return err1
		}
		_, err1 = print(string(bytes))
		return err1
	}
	output := outputResult{result, hashcode}
	bytes, err := jsonMarshal(output)
	if err != nil {
		_, _ = print(err.Error())
		return err
	}
	_, err = print(string(bytes))
	return err
}

// RunModifyAndWriteHash - logs the info, executes the action while locking the MTA file in the path, and writes the
// result and hashcode (or error, if needed) to the output.
func RunModifyAndWriteHash(info string, path string, force bool, action func() error, hashcode int, isNew bool) error {
	logs.Logger.Info(info)
	newHashcode, err := ModifyMta(path, action, hashcode, force, isNew, os.MkdirAll)
	writeErr := WriteResult(nil, newHashcode, err)
	if err != nil {
		// The original error has greater importance and will be displayed first.
		return err
	}
	return writeErr
}

// RunAndWriteResultAndHash - logs the info, executes the action, and writes the result and hashcode of the MTA in the
// path (or an error, if needed) to the output
func RunAndWriteResultAndHash(info string, path string, action func() (interface{}, error)) error {
	logs.Logger.Info(info)
	result, err := action()
	hashcode := 0
	if err == nil {
		hashcode, _, err = GetMtaHash(path)
	}
	writeErr := WriteResult(result, hashcode, err)
	if err != nil {
		// The original error has greater importance and will be displayed first.
		return err
	}
	return writeErr
}
