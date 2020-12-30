package fs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const PathNotFoundMsg = `could not read the "%s" file`

// CreateFile - creates a new file.
func CreateFile(path string) (file *os.File, err error) {
	file, err = os.Create(path) // Truncates the path if the file already exists.
	if err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("creation of the \"%s\" file failed", path))
	}
	// The caller needs to use the \"defer.close\" command.
	return file, err
}

// ReadFile reads the file and replaces Windows line breaks with \r
func ReadFile(path string) ([]byte, error) {
	fileContent, err := ioutil.ReadFile(filepath.Join(path))
	if err != nil {
		return nil, errors.Wrapf(err, PathNotFoundMsg, path)
	}
	s := string(fileContent)
	s = strings.Replace(s, "\r\n", "\r", -1)
	fileContent = []byte(s)
	return fileContent, nil
}

// DeleteFile - deletes the file.
func DeleteFile(path string) (err error) {
	return os.Remove(path)
}

// DeleteDir - deletes the directory and all sub-directories and files.
func DeleteDir(path string) (err error) {
	return os.RemoveAll(path)
}

// GetJSONContent reads the file, parses it as JSON and returns the result.
func GetJSONContent(path string) (map[string]interface{}, error) {
	// Get the file content
	fileConfigBuffer, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	fileConfig := make(map[string]interface{})
	err = json.Unmarshal(fileConfigBuffer, &fileConfig)
	if err != nil {
		return nil, err
	}
	return fileConfig, nil
}
