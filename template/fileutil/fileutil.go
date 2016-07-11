package fileutil

import (
	"fmt"
	"os"

	"github.com/HeavyHorst/remco/log"
)

// fileInfo describes a configuration file and is returned by fileStat.
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

// sameFile reports whether src and dest config files are equal.
// Two config files are equal when they have the same file contents and
// Unix permissions. The owner, group, and mode must match.
// It return false in other cases.
func SameFile(src, dest string) (bool, error) {
	if !IsFileExist(dest) {
		return false, nil
	}
	d, err := Stat(dest)
	if err != nil {
		return false, err
	}
	s, err := Stat(src)
	if err != nil {
		return false, err
	}
	if d.Uid != s.Uid {
		log.Info(fmt.Sprintf("%s has UID %d should be %d", dest, d.Uid, s.Uid))
	}
	if d.Gid != s.Gid {
		log.Info(fmt.Sprintf("%s has GID %d should be %d", dest, d.Gid, s.Gid))
	}
	if d.Mode != s.Mode {
		log.Info(fmt.Sprintf("%s has mode %s should be %s", dest, os.FileMode(d.Mode), os.FileMode(s.Mode)))
	}
	if d.Hash != s.Hash {
		log.Info(fmt.Sprintf("%s has hashsum %s should be %s", dest, d.Hash, s.Hash))
	}
	if d.Uid != s.Uid || d.Gid != s.Gid || d.Mode != s.Mode || d.Hash != s.Hash {
		return false, nil
	}
	return true, nil
}
