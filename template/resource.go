package template

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/template/fileutil"
	"github.com/cloudflare/cfssl/log"
	"github.com/flosch/pongo2"
	"github.com/kelseyhightower/memkv"
	flag "github.com/spf13/pflag"
)

// TemplateResource is the representation of a parsed template resource.
type TemplateResource struct {
	CheckCmd    string
	Dest        string
	FileMode    os.FileMode
	Gid         int
	Keys        []string
	Mode        string
	Prefix      string
	ReloadCmd   string
	Src         string
	StageFile   *os.File
	Uid         int
	funcMap     map[string]interface{}
	store       memkv.Store
	storeClient backends.StoreClient
	syncOnly    bool
	onetime     bool
}

var ErrEmptySrc = errors.New("empty src template")

// NewTemplateResource creates a TemplateResource.
func NewTemplateResource(storeClient backends.StoreClient, path string, flags *flag.FlagSet) (*TemplateResource, error) {
	if storeClient == nil {
		return nil, errors.New("A valid StoreClient is required.")
	}

	src, _ := flags.GetString("src")
	dst, _ := flags.GetString("dst")
	keys, _ := flags.GetStringSlice("keys")
	fileMode, _ := flags.GetString("fileMode")
	prefix, _ := flags.GetString("prefix")
	reloadCmd, _ := flags.GetString("reload_cmd")
	checkCmd, _ := flags.GetString("check_cmd")
	onetime, _ := flags.GetBool("onetime")

	if src == "" {
		return nil, ErrEmptySrc
	}

	// TODO implement flags for these values
	syncOnly := false
	uid := os.Geteuid()
	gid := os.Getegid()

	tr := &TemplateResource{
		Src:         src,
		Dest:        dst,
		Keys:        keys,
		Mode:        fileMode,
		Prefix:      prefix,
		ReloadCmd:   reloadCmd,
		CheckCmd:    checkCmd,
		storeClient: storeClient,
		store:       memkv.New(),
		funcMap:     newFuncMap(),
		Uid:         uid,
		Gid:         gid,
		syncOnly:    syncOnly,
		onetime:     onetime,
	}

	//Wrap the Memkv functions, so that they work with pongo2
	get := func(key string) memkv.KVPair {
		val, err := tr.store.Get(key)
		if err != nil {
			log.Warning(err)
		}
		return val
	}
	gets := func(pattern string) memkv.KVPairs {
		val, err := tr.store.GetAll(pattern)
		if err != nil {
			log.Warning(err)
		}
		return val
	}
	getv := func(key string, v ...string) string {
		val, err := tr.store.GetValue(key, v...)
		if err != nil {
			log.Warning(err)
		}
		return val
	}
	getvs := func(pattern string) []string {
		val, err := tr.store.GetAllValues(pattern)
		if err != nil {
			log.Warning(err)
		}
		return val
	}

	tr.funcMap["exists"] = tr.store.Exists
	tr.funcMap["ls"] = tr.store.List
	tr.funcMap["lsdir"] = tr.store.ListDir
	tr.funcMap["get"] = get
	tr.funcMap["gets"] = gets
	tr.funcMap["getv"] = getv
	tr.funcMap["getvs"] = getvs

	return tr, nil
}

// setVars sets the Vars for template resource.
func (t *TemplateResource) setVars() error {
	var err error
	log.Debug("Retrieving keys from store")
	log.Debug("Key prefix set to " + t.Prefix)

	result, err := t.storeClient.GetValues(appendPrefix(t.Prefix, t.Keys))
	if err != nil {
		return err
	}

	t.store.Purge()

	for k, v := range result {
		t.store.Set(path.Join("/", strings.TrimPrefix(k, t.Prefix)), v)
	}

	return nil
}

// createStageFile stages the src configuration file by processing the src
// template and setting the desired owner, group, and mode. It also sets the
// StageFile for the template resource.
// It returns an error if any.
func (t *TemplateResource) createStageFile() error {
	log.Debug("Using source template " + t.Src)

	if !fileutil.IsFileExist(t.Src) {
		return errors.New("Missing template: " + t.Src)
	}

	log.Debug("Compiling source template(pongo2) " + t.Src)
	tmpl, err := pongo2.FromFile(t.Src)
	if err != nil {
		return fmt.Errorf("Unable to process template %s, %s", t.Src, err)
	}

	// create TempFile in Dest directory to avoid cross-filesystem issues
	temp, err := ioutil.TempFile(filepath.Dir(t.Dest), "."+filepath.Base(t.Dest))
	if err != nil {
		return err
	}

	if err = tmpl.ExecuteWriter(t.funcMap, temp); err != nil {
		temp.Close()
		os.Remove(temp.Name())
		return err
	}

	defer temp.Close()

	// Set the owner, group, and mode on the stage file now to make it easier to
	// compare against the destination configuration file later.
	os.Chmod(temp.Name(), t.FileMode)
	os.Chown(temp.Name(), t.Uid, t.Gid)
	t.StageFile = temp
	return nil
}

// sync compares the staged and dest config files and attempts to sync them
// if they differ. sync will run a config check command if set before
// overwriting the target config file. Finally, sync will run a reload command
// if set to have the application or service pick up the changes.
// It returns an error if any.
func (t *TemplateResource) sync() error {
	staged := t.StageFile.Name()
	defer os.Remove(staged)

	log.Debug("Comparing candidate config to " + t.Dest)
	ok, err := fileutil.SameFile(staged, t.Dest)
	if err != nil {
		log.Error(err.Error())
	}

	if !ok {
		log.Info("Target config " + t.Dest + " out of sync")
		if !t.syncOnly && t.CheckCmd != "" {
			if err := t.check(); err != nil {
				return errors.New("Config check failed: " + err.Error())
			}
		}
		log.Debug("Overwriting target config " + t.Dest)
		err := os.Rename(staged, t.Dest)
		if err != nil {
			if strings.Contains(err.Error(), "device or resource busy") {
				log.Debug("Rename failed - target is likely a mount. Trying to write instead")
				// try to open the file and write to it
				var contents []byte
				var rerr error
				contents, rerr = ioutil.ReadFile(staged)
				if rerr != nil {
					return rerr
				}
				err := ioutil.WriteFile(t.Dest, contents, t.FileMode)
				// make sure owner and group match the temp file, in case the file was created with WriteFile
				os.Chown(t.Dest, t.Uid, t.Gid)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
		if !t.syncOnly && t.ReloadCmd != "" {
			if err := t.reload(); err != nil {
				return err
			}
		}
		log.Info("Target config " + t.Dest + " has been updated")
	} else {
		log.Debug("Target config " + t.Dest + " in sync")
	}
	return nil
}

// check executes the check command to validate the staged config file. The
// command is modified so that any references to src template are substituted
// with a string representing the full path of the staged file. This allows the
// check to be run on the staged file before overwriting the destination config
// file.
// It returns nil if the check command returns 0 and there are no other errors.
func (t *TemplateResource) check() error {
	var cmdBuffer bytes.Buffer
	data := make(map[string]string)
	data["src"] = t.StageFile.Name()
	tmpl, err := template.New("checkcmd").Parse(t.CheckCmd)
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
func (t *TemplateResource) reload() error {
	log.Debug("Running " + t.ReloadCmd)
	c := exec.Command("/bin/sh", "-c", t.ReloadCmd)
	output, err := c.CombinedOutput()
	if err != nil {
		log.Error(fmt.Sprintf("%q", string(output)))
		return err
	}
	log.Debug(fmt.Sprintf("%q", string(output)))
	return nil
}

// Process is a convenience function that wraps calls to the three main tasks
// required to keep local configuration files in sync. First we gather vars
// from the store, then we stage a candidate configuration file, and finally sync
// things up.
// It returns an error if any.
func (t *TemplateResource) process() error {
	if err := t.setFileMode(); err != nil {
		return err
	}
	if err := t.setVars(); err != nil {
		return err
	}
	if err := t.createStageFile(); err != nil {
		return err
	}
	if err := t.sync(); err != nil {
		return err
	}
	return nil
}

// setFileMode sets the FileMode.
func (t *TemplateResource) setFileMode() error {
	if t.Mode == "" {
		if !fileutil.IsFileExist(t.Dest) {
			t.FileMode = 0644
		} else {
			fi, err := os.Stat(t.Dest)
			if err != nil {
				return err
			}
			t.FileMode = fi.Mode()
		}
	} else {
		mode, err := strconv.ParseUint(t.Mode, 0, 32)
		if err != nil {
			return err
		}
		t.FileMode = os.FileMode(mode)
	}
	return nil
}

func (t *TemplateResource) Monitor() error {
	var lastIndex uint64
	stopChan := make(chan bool)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	defer close(stopChan)
	keys := appendPrefix(t.Prefix, t.Keys)

	go func() {
		for {
			if err := t.process(); err != nil {
				log.Error(err)
			}
			index, err := t.storeClient.WatchPrefix(t.Prefix, keys, lastIndex, stopChan)
			if err != nil {
				log.Error(err)
				// Prevent backend errors from consuming all resources.
				time.Sleep(time.Second * 2)
				continue
			}
			lastIndex = index
		}
	}()

	for {
		select {
		case s := <-signalChan:
			log.Info(fmt.Sprintf("Captured %v. Exiting...", s))
			os.Exit(0)
		}
	}
}

func (t *TemplateResource) Interval(interval int) error {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		err := t.process()
		if t.onetime {
			os.Exit(0)
		}
		select {
		case <-time.After(time.Duration(interval) * time.Second):
			if err != nil {
				log.Error(err)
			}
		case s := <-signalChan:
			log.Info(fmt.Sprintf("Captured %v. Exiting...", s))
			os.Exit(0)
		}
	}
}
