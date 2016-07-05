package template

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"strconv"

	"github.com/HeavyHorst/remco/template/fileutil"
	"github.com/cloudflare/cfssl/log"
)

type SrcDst struct {
	Src       string
	Dst       string
	Mode      string
	UID       int
	GID       int
	ReloadCmd string
	CheckCmd  string
	stageFile *os.File
	fileMode  os.FileMode
}

func (s *SrcDst) setFileMode() error {
	if s.Mode == "" {
		if !fileutil.IsFileExist(s.Dst) {
			s.fileMode = 0644
		} else {
			fi, err := os.Stat(s.Dst)
			if err != nil {
				return err
			}
			s.fileMode = fi.Mode()
		}
	} else {
		mode, err := strconv.ParseUint(s.Mode, 0, 32)
		if err != nil {
			return err
		}
		s.fileMode = os.FileMode(mode)
	}
	return nil
}

// check executes the check command to validate the staged config file. The
// command is modified so that any references to src template are substituted
// with a string representing the full path of the staged file. This allows the
// check to be run on the staged file before overwriting the destination config
// file.
// It returns nil if the check command returns 0 and there are no other errors.
func (s *SrcDst) check(stageFile string) error {
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
	log.Debug("Running " + cmdBuffer.String())
	c := exec.Command("/bin/sh", "-c", cmdBuffer.String())
	output, err := c.CombinedOutput()
	if err != nil {
		log.Error(fmt.Sprintf("%q", string(output)))
		return err
	}
	log.Debug(fmt.Sprintf("%q", string(output)))
	return nil
}

// reload executes the reload command.
// It returns nil if the reload command returns 0.
func (s *SrcDst) reload() error {
	log.Debug("Running " + s.ReloadCmd)
	c := exec.Command("/bin/sh", "-c", s.ReloadCmd)
	output, err := c.CombinedOutput()
	if err != nil {
		log.Error(fmt.Sprintf("%q", string(output)))
		return err
	}
	log.Debug(fmt.Sprintf("%q", string(output)))
	return nil
}
