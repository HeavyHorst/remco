/*
 * This file is part of remco.
 * Based on code from confd. https://github.com/kelseyhightower/confd
 * © 2013 Kelsey Hightower
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package template

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"

	"github.com/HeavyHorst/pongo2"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template/fileutil"
	"github.com/Sirupsen/logrus"
)

// Processor contains all data needed for the template processing
type Processor struct {
	Src       string `json:"src"`
	Dst       string `json:"dst"`
	Mode      string `json:"mode"`
	UID       int    `json:"uid"`
	GID       int    `json:"gid"`
	ReloadCmd string `toml:"reload_cmd" json:"reload_cmd"`
	CheckCmd  string `toml:"check_cmd" json:"check_cmd"`
	stageFile *os.File
	logger    *logrus.Entry
}

// createStageFile stages the src configuration file by processing the src
// template and setting the desired owner, group, and mode. It also sets the
// StageFile for the template resource.
// It returns an error if any.
func (s *Processor) createStageFile(funcMap map[string]interface{}) error {
	if !fileutil.IsFileExist(s.Src) {
		return errors.New("Missing template: " + s.Src)
	}

	s.logger.WithFields(logrus.Fields{
		"template": s.Src,
	}).Debug("Compiling source template")

	set := pongo2.NewSet("local", &pongo2.LocalFilesystemLoader{})
	tmpl, err := set.FromFile(s.Src)
	if err != nil {
		return fmt.Errorf("Unable to process template %s, %s", s.Src, err)
	}

	// create TempFile in Dest directory to avoid cross-filesystem issues
	temp, err := ioutil.TempFile(filepath.Dir(s.Dst), "."+filepath.Base(s.Dst))
	if err != nil {
		return err
	}

	if err = tmpl.ExecuteWriter(funcMap, temp); err != nil {
		temp.Close()
		os.Remove(temp.Name())
		return err
	}

	temp.Close()

	fileMode, err := s.getFileMode()
	if err != nil {
		return err
	}
	// Set the owner, group, and mode on the stage file now to make it easier to
	// compare against the destination configuration file later.
	os.Chmod(temp.Name(), fileMode)
	os.Chown(temp.Name(), s.UID, s.GID)
	s.stageFile = temp

	return nil
}

// syncFiles compares the staged and dest config files and attempts to sync them
// if they differ. syncFiles will run a config check command if set before
// overwriting the target config file. Finally, syncFile will run a reload command
// if set to have the application or service pick up the changes.
// It returns an error if any.
func (s *Processor) syncFiles() (bool, error) {
	var changed bool
	staged := s.stageFile.Name()
	defer os.Remove(staged)

	s.logger.WithFields(logrus.Fields{
		"staged": path.Base(staged),
		"dest":   s.Dst,
	}).Debug("Comparing staged and dest config files")

	ok, err := fileutil.SameFile(staged, s.Dst, s.logger)
	if err != nil {
		log.Error(err.Error())
	}

	if !ok {
		s.logger.WithFields(logrus.Fields{
			"config": s.Dst,
		}).Info("Target config out of sync")

		if err := s.check(staged); err != nil {
			return changed, errors.New("Config check failed: " + err.Error())
		}

		s.logger.WithFields(logrus.Fields{
			"config": s.Dst,
		}).Debug("Overwriting target config")

		fileMode, err := s.getFileMode()
		if err != nil {
			return changed, err
		}
		if err := fileutil.ReplaceFile(staged, s.Dst, fileMode, s.logger); err != nil {
			return changed, err
		}

		// make sure owner and group match the temp file, in case the file was created with WriteFile
		os.Chown(s.Dst, s.UID, s.GID)
		changed = true

		if err := s.reload(); err != nil {
			return changed, err
		}

		s.logger.WithFields(logrus.Fields{
			"config": s.Dst,
		}).Info("Target config has been updated")

	} else {
		s.logger.WithFields(logrus.Fields{
			"config": s.Dst,
		}).Debug("Target config in sync")

	}
	return changed, nil
}

func (s *Processor) getFileMode() (os.FileMode, error) {
	if s.Mode == "" {
		if !fileutil.IsFileExist(s.Dst) {
			return 0644, nil
		} else {
			fi, err := os.Stat(s.Dst)
			if err != nil {
				return 0, err
			}
			return fi.Mode(), nil
		}
	} else {
		mode, err := strconv.ParseUint(s.Mode, 0, 32)
		if err != nil {
			return 0, err
		}
		return os.FileMode(mode), nil
	}
}

// check executes the check command to validate the staged config file. The
// command is modified so that any references to src template are substituted
// with a string representing the full path of the staged file. This allows the
// check to be run on the staged file before overwriting the destination config file.
// It returns nil if the check command returns 0 and there are no other errors.
func (s *Processor) check(stageFile string) error {
	if s.CheckCmd == "" {
		return nil
	}
	var cmdBuffer bytes.Buffer
	data := make(map[string]string)
	data["src"] = stageFile
	tmpl, err := template.New("checkcmd").Parse(s.CheckCmd)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(&cmdBuffer, data); err != nil {
		return err
	}
	s.logger.Debug("Running " + cmdBuffer.String())
	c := exec.Command("/bin/sh", "-c", cmdBuffer.String())
	output, err := c.CombinedOutput()
	if err != nil {
		s.logger.Error(fmt.Sprintf("%q", string(output)))
		return err
	}
	s.logger.Debug(fmt.Sprintf("%q", string(output)))
	return nil
}

// reload executes the reload command.
// It returns nil if the reload command returns 0.
func (s *Processor) reload() error {
	if s.ReloadCmd == "" {
		return nil
	}
	s.logger.Debug("Running " + s.ReloadCmd)
	c := exec.Command("/bin/sh", "-c", s.ReloadCmd)
	output, err := c.CombinedOutput()
	if err != nil {
		s.logger.Error(fmt.Sprintf("%q", string(output)))
		return err
	}
	s.logger.Debug(fmt.Sprintf("%q", string(output)))
	return nil
}
