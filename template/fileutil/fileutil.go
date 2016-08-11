/*
 * This file is part of remco.
 * Based on code from confd. https://github.com/kelseyhightower/confd
 * © 2013 Kelsey Hightower
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package fileutil

import (
	"os"

	"github.com/HeavyHorst/remco/log"
	"github.com/Sirupsen/logrus"
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

// SameFile reports whether src and dest config files are equal.
// Two config files are equal when they have the same file contents and
// Unix permissions. The owner, group, and mode must match.
// It return false in other cases.
func SameFile(src, dest string) (bool, error) {
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
		log.WithFields(logrus.Fields{
			"config":  dest,
			"current": d.Uid,
			"new":     s.Uid,
		}).Info("wrong UID")
	}
	if d.Gid != s.Gid {
		log.WithFields(logrus.Fields{
			"config":  dest,
			"current": d.Gid,
			"new":     s.Gid,
		}).Info("wrong GID")
	}
	if d.Mode != s.Mode {
		log.WithFields(logrus.Fields{
			"config":  dest,
			"current": os.FileMode(d.Mode),
			"new":     os.FileMode(s.Mode),
		}).Info("wrong filemode")
	}
	if d.Hash != s.Hash {
		log.WithFields(logrus.Fields{
			"config":  dest,
			"current": d.Hash,
			"new":     s.Hash,
		}).Info("wrong hashsum")
	}
	if d.Uid != s.Uid || d.Gid != s.Gid || d.Mode != s.Mode || d.Hash != s.Hash {
		return false, nil
	}
	return true, nil
}
