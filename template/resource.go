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
	"sync"
	"syscall"
	"time"

	"github.com/HeavyHorst/memkv"
	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/template/fileutil"
	"github.com/cloudflare/cfssl/log"
	"github.com/flosch/pongo2"
	flag "github.com/spf13/pflag"
)

type SrcDst struct {
	Src       string
	Dst       string
	Mode      string
	UID       int
	GID       int
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

type StoreConfig struct {
	backends.StoreClient
	Name     string
	Onetime  bool
	Watch    bool
	Prefix   string
	Interval int
	Keys     []string
	store    *memkv.Store
}

// Resource is the representation of a parsed template resource.
type Resource struct {
	storeClients []StoreConfig
	ReloadCmd    string
	CheckCmd     string
	funcMap      map[string]interface{}
	store        memkv.Store
	syncOnly     bool
	sources      []*SrcDst
}

var ErrEmptySrc = errors.New("empty src template")

// NewResource creates a Resource.
func NewResource(storeClients []StoreConfig, sources []*SrcDst, reloadCmd, checkCmd string) (*Resource, error) {
	if len(storeClients) == 0 {
		return nil, errors.New("A valid StoreClient is required.")
	}

	for _, v := range sources {
		if v.Src == "" {
			return nil, ErrEmptySrc
		}
	}

	// TODO implement flags for these values
	syncOnly := false

	tr := &Resource{
		storeClients: storeClients,
		ReloadCmd:    reloadCmd,
		CheckCmd:     checkCmd,
		store:        memkv.New(),
		funcMap:      newFuncMap(),
		syncOnly:     syncOnly,
		sources:      sources,
	}

	// initialize the inidividual backend memkv Stores
	for i := range tr.storeClients {
		store := memkv.New()
		tr.storeClients[i].store = &store
	}

	addFuncs(tr.funcMap, tr.store.FuncMap)

	return tr, nil
}

// NewResourceFromFlags creates a Resource from the provided commandline flags.
func NewResourceFromFlags(s backends.Store, flags *flag.FlagSet, watch bool) (*Resource, error) {
	src, _ := flags.GetString("src")
	dst, _ := flags.GetString("dst")
	keys, _ := flags.GetStringSlice("keys")
	fileMode, _ := flags.GetString("fileMode")
	prefix, _ := flags.GetString("prefix")
	reloadCmd, _ := flags.GetString("reload_cmd")
	checkCmd, _ := flags.GetString("check_cmd")
	onetime, _ := flags.GetBool("onetime")
	interval, _ := flags.GetInt("interval")
	UID := os.Geteuid()
	GID := os.Getegid()

	sd := &SrcDst{
		Src:  src,
		Dst:  dst,
		Mode: fileMode,
		GID:  GID,
		UID:  UID,
	}

	b := StoreConfig{
		StoreClient: s.Client,
		Name:        s.Name,
		Onetime:     onetime,
		Prefix:      prefix,
		Watch:       watch,
		Interval:    interval,
		Keys:        keys,
	}

	return NewResource([]StoreConfig{b}, []*SrcDst{sd}, reloadCmd, checkCmd)
}

// setVars sets the Vars for template resource.
func (t *Resource) setVars(storeClient StoreConfig) error {
	var err error
	log.Debug("Retrieving keys from store: " + storeClient.Name)
	log.Debug("Key prefix set to " + storeClient.Prefix)

	result, err := storeClient.GetValues(appendPrefix(storeClient.Prefix, storeClient.Keys))
	if err != nil {
		return err
	}

	storeClient.store.Purge()

	for k, res := range result {
		storeClient.store.Set(path.Join("/", strings.TrimPrefix(k, storeClient.Prefix)), res)
	}

	//merge all stores
	t.store.Purge()
	for _, v := range t.storeClients {
		for _, kv := range v.store.GetAllKVs() {
			t.store.Set(kv.Key, kv.Value)
		}
	}

	return nil
}

// createStageFile stages the src configuration file by processing the src
// template and setting the desired owner, group, and mode. It also sets the
// StageFile for the template resource.
// It returns an error if any.
func (t *Resource) createStageFileAndSync() error {
	for _, s := range t.sources {
		log.Debug("Using source template " + s.Src)

		if !fileutil.IsFileExist(s.Src) {
			return errors.New("Missing template: " + s.Src)
		}

		log.Debug("Compiling source template(pongo2) " + s.Src)
		tmpl, err := pongo2.FromFile(s.Src)
		if err != nil {
			return fmt.Errorf("Unable to process template %s, %s", s.Src, err)
		}

		// create TempFile in Dest directory to avoid cross-filesystem issues
		temp, err := ioutil.TempFile(filepath.Dir(s.Dst), "."+filepath.Base(s.Dst))
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
		os.Chmod(temp.Name(), s.fileMode)
		os.Chown(temp.Name(), s.UID, s.GID)
		s.stageFile = temp

		if err = t.sync(s); err != nil {
			return err
		}
	}
	return nil
}

// sync compares the staged and dest config files and attempts to sync them
// if they differ. sync will run a config check command if set before
// overwriting the target config file. Finally, sync will run a reload command
// if set to have the application or service pick up the changes.
// It returns an error if any.
func (t *Resource) sync(s *SrcDst) error {
	staged := s.stageFile.Name()
	defer os.Remove(staged)

	log.Debug("Comparing candidate config to " + s.Dst)
	ok, err := fileutil.SameFile(staged, s.Dst)
	if err != nil {
		log.Error(err.Error())
	}

	if !ok {
		log.Info("Target config " + s.Dst + " out of sync")
		if !t.syncOnly && t.CheckCmd != "" {
			if err := t.check(staged); err != nil {
				return errors.New("Config check failed: " + err.Error())
			}
		}
		log.Debug("Overwriting target config " + s.Dst)
		err := os.Rename(staged, s.Dst)
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
				err := ioutil.WriteFile(s.Dst, contents, s.fileMode)
				// make sure owner and group match the temp file, in case the file was created with WriteFile
				os.Chown(s.Dst, s.UID, s.GID)
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
		log.Info("Target config " + s.Dst + " has been updated")
	} else {
		log.Debug("Target config " + s.Dst + " in sync")
	}
	return nil
}

// setFileMode sets the FileMode.
func (t *Resource) setFileMode() error {
	for _, s := range t.sources {
		if err := s.setFileMode(); err != nil {
			return nil
		}
	}
	return nil
}

// Process is a convenience function that wraps calls to the three main tasks
// required to keep local configuration files in sync. First we gather vars
// from the store, then we stage a candidate configuration file, and finally sync
// things up.
// It returns an error if any.
func (t *Resource) process(storeClient StoreConfig) error {
	if err := t.setFileMode(); err != nil {
		return err
	}
	if err := t.setVars(storeClient); err != nil {
		return err
	}
	if err := t.createStageFileAndSync(); err != nil {
		return err
	}
	return nil
}

func (t *Resource) processAll() error {
	//load all KV-pairs the first time
	var err error
	if err = t.setFileMode(); err != nil {
		return err
	}
	for _, storeClient := range t.storeClients {
		if err := t.setVars(storeClient); err != nil {
			return err
		}
	}
	if err = t.createStageFileAndSync(); err != nil {
		return err
	}
	return nil
}

// check executes the check command to validate the staged config file. The
// command is modified so that any references to src template are substituted
// with a string representing the full path of the staged file. This allows the
// check to be run on the staged file before overwriting the destination config
// file.
// It returns nil if the check command returns 0 and there are no other errors.
func (t *Resource) check(stageFile string) error {
	var cmdBuffer bytes.Buffer
	data := make(map[string]string)
	data["src"] = stageFile
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
func (t *Resource) reload() error {
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

func (s StoreConfig) watch(stopChan chan bool, processChan chan StoreConfig) {
	var lastIndex uint64
	keysPrefix := appendPrefix(s.Prefix, s.Keys)

	for {
		select {
		case <-stopChan:
			return
		default:
			index, err := s.WatchPrefix(s.Prefix, keysPrefix, lastIndex, stopChan)
			if err != nil {
				log.Error(err)
				// Prevent backend errors from consuming all resources.
				time.Sleep(time.Second * 2)
				continue
			}
			processChan <- s
			lastIndex = index
		}
	}
}

func (s StoreConfig) interval(stopChan chan bool, processChan chan StoreConfig) {
	if s.Onetime {
		return
	}
	for {
		select {
		case <-stopChan:
			return
		case <-time.After(time.Duration(s.Interval) * time.Second):
			processChan <- s
		}
	}
}

func (t *Resource) Monitor() {
	wg := &sync.WaitGroup{}
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	stopChan := make(chan bool)
	processChan := make(chan StoreConfig)

	defer close(processChan)

	if err := t.processAll(); err != nil {
		log.Error(err)
	}

	for _, sc := range t.storeClients {
		wg.Add(1)
		if sc.Watch {
			go func(s StoreConfig) {
				defer wg.Done()
				//t.watch(s, stopChan, processChan)
				s.watch(stopChan, processChan)
			}(sc)
		} else {
			go func(s StoreConfig) {
				defer wg.Done()
				s.interval(stopChan, processChan)
			}(sc)
		}
	}

	go func() {
		// If there is no goroutine left - quit
		wg.Wait()
		signalChan <- syscall.SIGINT
	}()

	for {
		select {
		case storeClient := <-processChan:
			if err := t.process(storeClient); err != nil {
				log.Error(err)
			}
		case s := <-signalChan:
			log.Info(fmt.Sprintf("Captured %v. Exiting...", s))
			close(stopChan)
			// drain processChan
			go func() {
				for range processChan {
				}
			}()
			wg.Wait()
			return
		}
	}
}
