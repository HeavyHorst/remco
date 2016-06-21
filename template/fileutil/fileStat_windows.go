package fileutil

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
)

// Stat return a fileInfo describing the named file.
func Stat(name string) (fi fileInfo, err error) {
	if IsFileExist(name) {
		f, err := os.Open(name)
		defer f.Close()
		if err != nil {
			return fi, err
		}
		stats, _ := f.Stat()
		fi.Mode = stats.Mode()
		h := md5.New()
		io.Copy(h, f)
		fi.Md5 = fmt.Sprintf("%x", h.Sum(nil))
		return fi, nil
	}
	return fi, errors.New("File not found")
}
