/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package template

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/HeavyHorst/memkv"
	"github.com/HeavyHorst/pongo2"
	"github.com/dop251/goja"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	gyml "sigs.k8s.io/yaml"
)

func init() {
	pongo2.RegisterFilter("sortByLength", filterSortByLength)
	pongo2.RegisterFilter("parseInt", filterParseInt)
	pongo2.RegisterFilter("parseFloat", filterParseFloat)
	pongo2.RegisterFilter("parseYAML", filterUnmarshalYAML)
	pongo2.RegisterFilter("parseJSON", filterUnmarshalYAML)      // just an alias
	pongo2.RegisterFilter("parseYAMLArray", filterUnmarshalYAML) // deprecated
	pongo2.RegisterFilter("toJSON", filterToJSON)
	pongo2.RegisterFilter("toPrettyJSON", filterToPrettyJSON)
	pongo2.RegisterFilter("toYAML", filterToYAML)
	pongo2.RegisterFilter("dir", filterDir)
	pongo2.RegisterFilter("base", filterBase)
	pongo2.RegisterFilter("base64", filterBase64)
	pongo2.RegisterFilter("index", filterIndex)
	pongo2.RegisterFilter("mapValue", filterMapValue)
}

// RegisterCustomJsFilters loads all filters from the given directory.
// It returns an error if any.
func RegisterCustomJsFilters(folder string) error {
	files, _ := ioutil.ReadDir(folder)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".js") {
			fp := filepath.Join(folder, file.Name())
			buf, err := ioutil.ReadFile(fp)
			if err != nil {
				return errors.Errorf("couldn't load custom filter %s", fp)
			}
			name := file.Name()
			name = name[0 : len(name)-3]

			filterFunc := pongoJSFilter(name, string(buf))

			if err := pongo2.RegisterFilter(name, filterFunc); err != nil {
				if err := pongo2.ReplaceFilter(name, filterFunc); err != nil {
					return errors.Errorf("couldn't replace existing filter %s", name)
				}
			}
		}
	}
	return nil
}

func pongoJSFilter(name string, js string) func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		vm := goja.New()

		vm.Set("In", in.Interface())
		vm.Set("Param", param.Interface())

		v, err := vm.RunString(js)
		if err != nil {
			return nil, &pongo2.Error{
				Sender:    "javascript-filter:" + name,
				OrigError: err,
			}
		}

		return pongo2.AsValue(v.Export()), nil
	}
}

func parseParamMap(in string) (url.Values, error) {
	in = strings.ReplaceAll(in, ", ", ",")
	in = strings.ReplaceAll(in, ",", "&")
	return url.ParseQuery(in)
}

func filterBase64(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !in.IsString() {
		return in, nil
	}
	sEnc := base64.StdEncoding.EncodeToString([]byte(in.String()))
	return pongo2.AsValue(sEnc), nil
}

func filterBase(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !in.IsString() {
		return in, nil
	}
	return pongo2.AsValue(path.Base(in.String())), nil
}

func filterDir(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !in.IsString() {
		return in, nil
	}
	return pongo2.AsValue(path.Dir(in.String())), nil
}

func filterToPrettyJSON(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	b, err := json.MarshalIndent(in.Interface(), "", "    ")
	if err != nil {
		return nil, &pongo2.Error{
			Sender:    "filter:filterToPrettyJSON",
			OrigError: err,
		}
	}
	return pongo2.AsValue(string(b)), nil
}

func filterToJSON(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	b, err := json.Marshal(in.Interface())
	if err != nil {
		return nil, &pongo2.Error{
			Sender:    "filter:filterToJSON",
			OrigError: err,
		}
	}
	return pongo2.AsValue(string(b)), nil
}

func filterToYAML(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	b := bytes.Buffer{}
	yamlEncoder := yaml.NewEncoder(&b)

	if param != nil && param.String() != "" {
		pm, err := parseParamMap(param.String())
		if err != nil {
			return nil, &pongo2.Error{
				Sender:    "filter:filterToYAML",
				OrigError: fmt.Errorf("could't parese parameter list: %w", err),
			}
		}

		indent, err := strconv.Atoi(pm.Get("indent"))
		if err != nil {
			return nil, &pongo2.Error{
				Sender:    "filter:filterToYAML",
				OrigError: fmt.Errorf("couldn't parse integer: %w", err),
			}
		}

		yamlEncoder.SetIndent(indent)
	}

	err := yamlEncoder.Encode(in.Interface())
	if err != nil {
		return nil, &pongo2.Error{
			Sender:    "filter:filterToYAML",
			OrigError: err,
		}
	}
	return pongo2.AsValue(string(b.Bytes())), nil
}

func filterParseInt(in, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !in.IsString() {
		return in, nil
	}

	ins := in.String()
	if ins == "" {
		return pongo2.AsValue(0), nil
	}

	result, err := strconv.ParseInt(ins, 10, 64)
	if err != nil {
		return nil, &pongo2.Error{
			Sender:    "filter:filterParseInt",
			OrigError: err,
		}
	}

	return pongo2.AsValue(result), nil
}

func filterParseFloat(in, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !in.IsString() {
		return in, nil
	}

	ins := in.String()
	if ins == "" {
		return pongo2.AsValue(0.0), nil
	}

	result, err := strconv.ParseFloat(ins, 10)
	if err != nil {
		return nil, &pongo2.Error{
			Sender:    "filter:filterParseFloat",
			OrigError: err,
		}
	}

	return pongo2.AsValue(result), nil
}

func filterUnmarshalYAML(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !in.IsString() {
		return in, nil
	}

	var ret interface{}
	if err := gyml.Unmarshal([]byte(in.String()), &ret); err != nil {
		return nil, &pongo2.Error{
			Sender:    "filter:filterUnmarshalYAML",
			OrigError: err,
		}
	}

	return pongo2.AsValue(ret), nil
}

func filterIndex(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !in.CanSlice() {
		return in, nil
	}

	index := param.Integer()
	if index < 0 {
		index = in.Len() + index
	}

	return pongo2.AsValue(in.Index(index).Interface()), nil
}

func filterMapValue(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if in == nil || in.IsNil() {
		return pongo2.AsValue(nil), nil
	}

	val := reflect.ValueOf(in.Interface())
	if val.Kind() == reflect.Map {
		valueType := val.Type().Key().Kind()
		paramValue := reflect.ValueOf(param.Interface())

		if paramValue.Kind() != valueType {
			return pongo2.AsValue(nil), nil
		}

		mv := val.MapIndex(paramValue)
		if !mv.IsValid() {
			return pongo2.AsValue(nil), nil
		}

		return pongo2.AsValue(mv.Interface()), nil
	}
	return pongo2.AsValue(nil), nil
}

func filterSortByLength(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !in.CanSlice() {
		return in, nil
	}

	values := in.Interface()
	switch v := values.(type) {
	case []string:
		sort.Slice(v, func(i, j int) bool {
			return len(v[i]) < len(v[j])
		})
		return pongo2.AsValue(v), nil
	case memkv.KVPairs:
		sort.Slice(v, func(i, j int) bool {
			return len(v[i].Key) < len(v[j].Key)
		})
		return pongo2.AsValue(v), nil
	}

	return in, nil
}
