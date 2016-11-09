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
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"strconv"

	"github.com/HeavyHorst/remco/template/fileutil"
	"github.com/Sirupsen/logrus"
)

// ProcessConfig contains all data needed for the template processing
type ProcessConfig struct {
	Src       string
	Dst       string
	Mode      string
	UID       int
	GID       int
	ReloadCmd string
	CheckCmd  string
	stageFile *os.File
	logger    *logrus.Entry
}

func (s *ProcessConfig) getFileMode() (os.FileMode, error) {
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
func (s *ProcessConfig) check(stageFile string) error {
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
func (s *ProcessConfig) reload() error {
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
