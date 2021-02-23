package mta

import (
	"crypto/sha256"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	ghodss "github.com/ghodss/yaml"
	"github.com/json-iterator/go"

	"github.com/SAP/cloud-mta/internal/fs"
	"github.com/SAP/cloud-mta/internal/logs"
)

const UnmarshalFailsMsg = `the "%s" file is not a valid MTA descriptor`

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

func GetMtaFromFile(path string, extensions []string, returnMergeError bool) (mta *MTA, messages []string, err error) {
	mtaContent, err := fs.ReadFile(filepath.Join(path))
	if err != nil {
		return nil, nil, err
	}
	mta, err = Unmarshal(mtaContent)
	if err != nil {
		return nil, nil, errors.Wrapf(err, UnmarshalFailsMsg, path)
	}

	// If there is an error during the merge return the result so far and return the error as a message (or error if required).
	extErr := mergeWithExtensionFiles(mta, extensions, path)
	if extErr != nil {
		if returnMergeError {
			return mta, nil, extErr
		}
		messages = []string{extErr.Error()}
	}
	return mta, messages, nil
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

// DeleteMta - deletes the MTA
func DeleteMta(path string) error {
	return fs.DeleteDir(path)
}

//AddModule - adds a new module.
func AddModule(path string, moduleDataJSON string, marshal func(*MTA) ([]byte, error)) ([]string, error) {
	mta, messages, err := GetMtaFromFile(filepath.Join(path), nil, false)
	if err != nil {
		return messages, err
	}

	module := Module{}
	err = unmarshalData(moduleDataJSON, &module)
	if err != nil {
		return messages, err
	}

	mta.Modules = append(mta.Modules, &module)
	return messages, saveMTA(path, mta, marshal)
}

//AddResource - adds a new resource.
func AddResource(path string, resourceDataJSON string, marshal func(*MTA) ([]byte, error)) ([]string, error) {
	mta, messages, err := GetMtaFromFile(path, nil, false)
	if err != nil {
		return messages, err
	}

	resource := Resource{}
	err = unmarshalData(resourceDataJSON, &resource)
	if err != nil {
		return messages, err
	}

	mta.Resources = append(mta.Resources, &resource)
	return messages, saveMTA(path, mta, marshal)
}

//GetModules - gets all modules.
func GetModules(path string, extensions []string) ([]*Module, []string, error) {
	mta, messages, err := GetMtaFromFile(path, extensions, false)
	if err != nil {
		return nil, messages, err
	}
	return mta.Modules, messages, nil
}

//GetResources - gets all resources.
func GetResources(path string, extensions []string) ([]*Resource, []string, error) {
	mta, messages, err := GetMtaFromFile(path, extensions, false)
	if err != nil {
		return nil, messages, err
	}
	return mta.Resources, messages, nil
}

// GetResourceConfig returns the configuration for a resource (its service creation parameters).
// If both the config and path parameters are defined, the result is merged.
func GetResourceConfig(path string, extensions []string, resourceName string, workspaceDir string) (map[string]interface{}, []string, error) {
	mta, messages, err := GetMtaFromFile(path, extensions, false)
	if err != nil {
		return nil, messages, err
	}
	if len(workspaceDir) == 0 {
		workspaceDir = filepath.Dir(path)
	}

	resource := mta.GetResourceByName(resourceName)
	if resource == nil {
		return nil, messages, fmt.Errorf("the '%s' resource does not exist", resourceName)
	}

	// Get the resource config from its parameters
	configParam := resource.Parameters["config"]
	var config map[string]interface{}
	if configParam != nil {
		config, _, _ = getMapValue(configParam)
	}

	// Get the resource service creation parameters file path
	filePath := resource.Parameters["path"]
	var fileConfig map[string]interface{}

	if filePath != nil {
		fileConfig, err = fs.GetJSONContent(filepath.Join(workspaceDir, filePath.(string)))
		if err != nil {
			return nil, messages, err
		}
	}

	return mergeMaps(config, fileConfig), messages, nil
}

// Shallow merge the maps. If the first map is not nil the merge result is inlined in it.
// The first map keys override the second map keys.
func mergeMaps(first map[string]interface{}, second map[string]interface{}) map[string]interface{} {
	var result map[string]interface{}
	if first == nil && second == nil {
		// Both maps are nil - return empty map
		result = make(map[string]interface{})
	} else if first == nil {
		// Only second exists
		result = second
	} else if second == nil {
		// Only first exists
		result = first
	} else {
		// Both maps are not nil at this point.
		// Shallow merge the maps (same as the deployer).
		result = first
		for key, value := range second {
			// The first map keys override the second map keys.
			_, ok := result[key]
			if !ok {
				result[key] = value
			}
		}
	}

	return result
}

// UpdateModule updates an existing module according to the module name. If more than one module with this
// name exists, one of the modules is updated to the existing structure.
func UpdateModule(path string, moduleDataJSON string, marshal func(*MTA) ([]byte, error)) ([]string, error) {
	mtaObj, messages, err := GetMtaFromFile(path, nil, false)
	if err != nil {
		return messages, err
	}

	module := Module{}
	err = unmarshalData(moduleDataJSON, &module)
	if err != nil {
		return messages, err
	}

	// Replaces the first existing module with the same name.
	for index, existingModule := range mtaObj.Modules {
		if existingModule.Name == module.Name {
			mtaObj.Modules[index] = &module
			return messages, saveMTA(path, mtaObj, marshal)
		}
	}

	return messages, fmt.Errorf("the '%s' module does not exist", module.Name)
}

// UpdateResource updates an existing resource according to the resource name. If more than one resource with this
// name exists, one of the resources is updated in the existing structure.
func UpdateResource(path string, resourceDataJSON string, marshal func(*MTA) ([]byte, error)) ([]string, error) {
	mtaObj, messages, err := GetMtaFromFile(path, nil, false)
	if err != nil {
		return messages, err
	}

	resource := Resource{}
	err = unmarshalData(resourceDataJSON, &resource)
	if err != nil {
		return messages, err
	}

	// Replaces the first existing resource with the same name.
	for index, existingResource := range mtaObj.Resources {
		if existingResource.Name == resource.Name {
			mtaObj.Resources[index] = &resource
			return messages, saveMTA(path, mtaObj, marshal)
		}
	}

	return messages, fmt.Errorf("the '%s' resource does not exist", resource.Name)
}

//GetMtaID - gets MTA ID.
func GetMtaID(path string) (string, []string, error) {
	mta, messages, err := GetMtaFromFile(path, nil, false)
	if err != nil {
		return "", messages, err
	}
	return mta.ID, messages, nil
}

//IsNameUnique - checks if the name already exists as a `module`/`resource`/`provide` name.
func IsNameUnique(path string, name string) (bool, []string, error) {
	mta, messages, err := GetMtaFromFile(path, nil, false)
	if err != nil {
		return true, messages, err
	}

	for _, module := range mta.Modules {
		if name == module.Name {
			return true, messages, nil
		}
		for _, provide := range module.Provides {
			if name == provide.Name {
				return true, messages, nil
			}
		}
	}
	for _, resource := range mta.Resources {
		if name == resource.Name {
			return true, messages, nil
		}
	}
	return false, messages, nil
}

//getBuildParameters - gets the MTA build parameters.
func GetBuildParameters(path string, extensions []string) (*ProjectBuild, []string, error) {
	mta, messages, err := GetMtaFromFile(path, extensions, false)
	if err != nil {
		return nil, messages, err
	}
	return mta.BuildParams, messages, nil
}

//getParameters - gets the MTA parameters.
func GetParameters(path string, extensions []string) (*map[string]interface{}, []string, error) {
	mta, messages, err := GetMtaFromFile(path, extensions, false)
	if err != nil {
		return nil, messages, err
	}
	return &mta.Parameters, messages, nil
}

//UpdateBuildParameters - updates the MTA build parameters.
func UpdateBuildParameters(path string, buildParamsDataJSON string) ([]string, error) {
	mta, messages, err := GetMtaFromFile(path, nil, false)
	if err != nil {
		return messages, err
	}

	buildParams := ProjectBuild{}
	err = unmarshalData(buildParamsDataJSON, &buildParams)
	if err != nil {
		return messages, err
	}

	mta.BuildParams = &buildParams
	return messages, saveMTA(path, mta, Marshal)
}

//UpdateParameters - updates the MTA parameters.
func UpdateParameters(path string, paramsDataJSON string) ([]string, error) {
	mta, messages, err := GetMtaFromFile(path, nil, false)
	if err != nil {
		return messages, err
	}

	params := map[string]interface{}{}
	err = unmarshalData(paramsDataJSON, &params)
	if err != nil {
		return messages, err
	}

	mta.Parameters = params
	return messages, saveMTA(path, mta, Marshal)
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

	folder := filepath.Dir(dst)
	rerr = os.MkdirAll(folder, os.ModePerm)
	if rerr != nil {
		return
	}
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
	h := sha256.New()
	code, err := h.Write(mtaContent)
	return code, true, err
}

// ModifyMta - locks and modifies the "mta.yaml" file.
func ModifyMta(path string, modify func() ([]string, error), hashcode int, force bool, isNew bool, mkDirs func(string, os.FileMode) error) (newHashcode int, messages []string, rerr error) {
	// Creates the lock file.
	// Makes sure the directory of the lock file exists (it might not exist if it is a new MTA).
	folder := filepath.Dir(path)
	rerr = mkDirs(folder, os.ModePerm)
	if rerr != nil {
		return 0, nil, rerr
	}
	lockFilePath := filepath.Join(filepath.Dir(path), "mta-lock.lock")
	file, err := os.OpenFile(lockFilePath, os.O_RDONLY|os.O_CREATE|os.O_EXCL, 0666)
	if os.IsExist(err) {
		return 0, nil, fmt.Errorf("could not modify the \"%s\" file; it is locked by another process", path)
	} else if err != nil {
		return 0, nil, fmt.Errorf("could not lock the \"%s\" file for modification; %s", path, err)
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
		messages, err = modify()
	}
	if err != nil {
		return 0, messages, err
	}
	newHashcode, _, err = GetMtaHash(path)
	return newHashcode, messages, err
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
	Messages []string    `json:"messages,omitempty"`
	Hashcode int         `json:"hashcode"`
}
type outputError struct {
	Message string `json:"message"`
}

// WriteResult - writes the result of an operation to the output in JSON format. If successful, the hashcode and results are written; otherwise an error is displayed.
func WriteResult(result interface{}, messages []string, hashcode int, err error) error {
	return printResult(result, messages, hashcode, err, fmt.Print, jsoniter.Marshal)
}

func printResult(result interface{}, messages []string, hashcode int, err error, print func(...interface{}) (n int, err error), jsonMarshal func(v interface{}) ([]byte, error)) error {
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
	output := outputResult{result, messages, hashcode}
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
func RunModifyAndWriteHash(info string, path string, force bool, action func() ([]string, error), hashcode int, isNew bool) error {
	logs.Logger.Info(info)
	newHashcode, messages, err := ModifyMta(path, action, hashcode, force, isNew, os.MkdirAll)
	writeErr := WriteResult(nil, messages, newHashcode, err)
	if err != nil {
		// If there is an error in both the “ModifyMta” function and the “WriteResult” function, only the “ModifyMta”
		// function returns the error.
		return err
	}
	return writeErr
}

// RunAndWriteResultAndHash - logs the info, executes the action, and writes the result and hashcode of the MTA in the
// path (or an error, if needed) to the output
func RunAndWriteResultAndHash(info string, path string, extensions []string, action func() (interface{}, []string, error)) error {
	logs.Logger.Info(info)
	result, messages, err := action()
	hashcode := 0
	if err == nil && len(extensions) == 0 {
		hashcode, _, err = GetMtaHash(path)
	}
	writeErr := WriteResult(result, messages, hashcode, err)
	if err != nil {
		// If there is an error in both the “GetMtaHash” function and the “WriteResult” function, only the “GetMtaHash”
		// function returns the error.
		return err
	}
	return writeErr
}
