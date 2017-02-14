// +build !windows

/*
 * This file is part of remco.
 * Based on code from confd.
 * https://github.com/kelseyhightower/confd/blob/abba746a0cb7c8cb5fe135fa2d884ea3c4a5f666/resource/template/util.go
 * © 2013 Kelsey Hightower
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package fileutil

import (
	hash "crypto/sha1"

	"github.com/pkg/errors"

	"fmt"
	"io"
	"os"
	"syscall"
)

// stat return a fileInfo describing the named file.
func stat(name string) (fi fileInfo, err error) {
	if IsFileExist(name) {
		f, err := os.Open(name)
		if err != nil {
			return fi, errors.Wrap(err, "open file failed")
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
	return fi, fmt.Errorf("file not found")
}
