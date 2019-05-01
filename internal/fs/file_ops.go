package fs

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
)

// CreateFile - create new file
func CreateFile(path string) (file *os.File, err error) {
	file, err = os.Create(path) // Truncates if file already exists
	if err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("creation of the %s file failed", path))
	}
	// The caller needs to use defer.close
	return file, err
}
