package iodatafmt

import (
	// Base packages.
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	// Local packages.
	yaml "github.com/mickep76/iodatafmt/yaml_mapstr"

	// Third party packages.
	"github.com/BurntSushi/toml"
)

// DataFmt represents which data serialization is used YAML, JSON or TOML.
type DataFmt int

// Constants for data format.
const (
	YAML DataFmt = iota
	TOML
	JSON
	UNKNOWN
)

// Unmarshal YAML/JSON/TOML serialized data.
func Unmarshal(b []byte, f DataFmt) (interface{}, error) {
	var d interface{}

	switch f {
	case YAML:
		if err := yaml.Unmarshal(b, &d); err != nil {
			return nil, err
		}
	case TOML:
		if err := toml.Unmarshal(b, &d); err != nil {
			return nil, err
		}
	case JSON:
		if err := json.Unmarshal(b, &d); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported data format")
	}

	return d, nil
}

// UnmarshalPtr YAML/JSON/TOML serialized data.
func UnmarshalPtr(ptr interface{}, b []byte, f DataFmt) error {
	switch f {
	case YAML:
		if err := yaml.Unmarshal(b, ptr); err != nil {
			return err
		}
	case TOML:
		if err := toml.Unmarshal(b, ptr); err != nil {
			return err
		}
	case JSON:
		if err := json.Unmarshal(b, ptr); err != nil {
			return err
		}
	default:
		return errors.New("unsupported data format")
	}

	return nil
}

// Marshal YAML/JSON/TOML serialized data.
func Marshal(d interface{}, f DataFmt) ([]byte, error) {
	switch f {
	case YAML:
		b, err := yaml.Marshal(&d)
		if err != nil {
			return nil, err
		}
		return b, nil
	case TOML:
		b := new(bytes.Buffer)
		if err := toml.NewEncoder(b).Encode(&d); err != nil {
			return nil, err
		}
		return b.Bytes(), nil
	case JSON:
		b, err := json.MarshalIndent(&d, "", "  ")
		if err != nil {
			return nil, err
		}
		return b, nil
	default:
		return nil, errors.New("unsupported data format")
	}
}

// Format returns DataFmt constant based on a string.
func Format(s string) (DataFmt, error) {
	switch strings.ToUpper(s) {
	case "YAML":
		return YAML, nil
	case "TOML":
		return TOML, nil
	case "JSON":
		return JSON, nil
	default:
		return UNKNOWN, errors.New("unsupported data format")
	}
}

// FileFormat returns DataFmt constant based on file extension.
func FileFormat(fn string) (DataFmt, error) {
	switch filepath.Ext(fn) {
	case ".yaml":
		return YAML, nil
	case ".yml":
		return YAML, nil
	case ".json":
		return JSON, nil
	case ".toml":
		return TOML, nil
	case ".tml":
		return TOML, nil
	default:
		return UNKNOWN, errors.New("unsupported data format")
	}
}

// Load a file with serialized data.
func Load(fn string, f DataFmt) (interface{}, error) {
	if _, err := os.Stat(fn); os.IsNotExist(err) {
		return nil, errors.New("file doesn't exist")
	}

	b, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}

	d, err := Unmarshal(b, f)
	if err != nil {
		return nil, err
	}

	return d, nil
}

// LoadPtr a file with serialized data.
func LoadPtr(ptr interface{}, fn string, f DataFmt) error {
	if _, err := os.Stat(fn); os.IsNotExist(err) {
		return errors.New("file doesn't exist")
	}

	b, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}

	if err := UnmarshalPtr(ptr, b, f); err != nil {
		return err
	}

	return nil
}

// Write a file with serialized data.
func Write(fn string, d map[string]interface{}, f DataFmt) error {
	b, err := Marshal(d, f)
	if err != nil {
		return err
	}

	w, err := os.Create(fn)
	if err != nil {
		return err
	}

	if _, err = w.Write(b); err != nil {
		return err
	}

	w.Close()
	return nil
}

// Print serialized data.
func Print(d interface{}, f DataFmt) error {
	b, err := Marshal(d, f)
	if err != nil {
		return err
	}

	fmt.Println(string(b))
	return nil
}

// Sprint return serialized data.
func Sprint(d interface{}, f DataFmt) (string, error) {
	b, err := Marshal(d, f)
	if err != nil {
		return "", err
	}

	return fmt.Sprintln(string(b)), nil
}
