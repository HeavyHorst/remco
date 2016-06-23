// +build !windows

package fileutil

import (
	hash "crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
)

// Stat return a fileInfo describing the named file.
func Stat(name string) (fi fileInfo, err error) {
	if IsFileExist(name) {
		f, err := os.Open(name)
		if err != nil {
			return fi, err
		}
		defer f.Close()
		stats, _ := f.Stat()
		fi.Uid = stats.Sys().(*syscall.Stat_t).Uid
		fi.Gid = stats.Sys().(*syscall.Stat_t).Gid
		fi.Mode = stats.Mode()
		h := hash.New()
		io.Copy(h, f)
		fi.Hash = fmt.Sprintf("%x", h.Sum(nil))
		return fi, nil
	}
	return fi, errors.New("File not found")
}
