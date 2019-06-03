package fs

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
)

// CreateFile - creates a new file.
func CreateFile(path string) (file *os.File, err error) {
	file, err = os.Create(path) // Truncates the path if the file already exists.
	if err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("creation of the \"%s\" file failed", path))
	}
	// The caller needs to use the \"defer.close\" command.
	return file, err
}

// DeleteFile - deletes the file.
func DeleteFile(path string) (err error) {
	return os.Remove(path)
}
