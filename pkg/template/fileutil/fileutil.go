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
	"github.com/hashicorp/go-hclog"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
)

// FileInfo describes a configuration file and is returned by filestat.
type fileInfo struct {
	Uid  uint32
	Gid  uint32
	Mode os.FileMode
	Hash string
}

// IsFileExist reports whether path exits.
func IsFileExist(fpath string) bool {
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		return false
	}
	return true
}

// ReplaceFile replaces dest with src.
//
// ReplaceFile just renames (move) the file if possible.
// If that fails it will read the src file and write the content to the destination file.
// It returns an error if any.
func ReplaceFile(src, dest string, mode os.FileMode, logger hclog.Logger) error {
	err := os.Rename(src, dest)
	if err != nil {
		if strings.Contains(err.Error(), "device or resource busy") {
			logger.Debug("Rename failed - target is likely a mount. Trying to write instead")
			// try to open the file and write to it
			var contents []byte
			var rerr error
			contents, rerr = ioutil.ReadFile(src)
			if rerr != nil {
				return errors.Wrap(rerr, "couldn't read source file")
			}
			err := ioutil.WriteFile(dest, contents, mode)
			if err != nil {
				return errors.Wrap(rerr, "couldn't write destination file")
			}
		} else {
			return errors.Wrap(err, "couldn't rename src -> dst")
		}
	}
	return nil
}

// SameFile reports whether src and dest config files are equal.
// Two config files are equal when they have the same file contents and
// Unix permissions. The owner, group, and mode must match.
// It return false in other cases.
func SameFile(src, dest string, logger hclog.Logger) (bool, error) {
	if !IsFileExist(dest) {
		return false, nil
	}
	d, err := stat(dest)
	if err != nil {
		return false, err
	}
	s, err := stat(src)
	if err != nil {
		return false, err
	}
	if d.Uid != s.Uid {
		logger.With(
			"config", dest,
			"current", d.Uid,
			"new", s.Uid,
		).Info("wrong UID")
	}
	if d.Gid != s.Gid {
		logger.With(
			"config", dest,
			"current", d.Gid,
			"new", s.Gid,
		).Info("wrong GID")
	}
	if d.Mode != s.Mode {
		logger.With(
			"config", dest,
			"current", d.Mode,
			"new", s.Mode,
		).Info("wrong filemode")
	}
	if d.Hash != s.Hash {
		logger.With(
			"config", dest,
			"current", d.Hash,
			"new", s.Hash,
		).Info("wrong hashsum")
	}
	if d.Uid != s.Uid || d.Gid != s.Gid || d.Mode != s.Mode || d.Hash != s.Hash {
		return false, nil
	}
	return true, nil
}
