/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package template

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/crypto/openpgp"

	"github.com/HeavyHorst/memkv"
	"github.com/HeavyHorst/pongo2"
	"github.com/HeavyHorst/remco/pkg/log"
	"github.com/dop251/goja"
	"github.com/mickep76/iodatafmt/yaml_mapstr"
	"github.com/pkg/errors"
)

func init() {
	pongo2.RegisterFilter("sortByLength", filterSortByLength)
	pongo2.RegisterFilter("parseYAML", filterUnmarshalYAML)
	pongo2.RegisterFilter("parseJSON", filterUnmarshalYAML)      // just an alias
	pongo2.RegisterFilter("parseYAMLArray", filterUnmarshalYAML) // deprecated
	pongo2.RegisterFilter("toJSON", filterToJSON)
	pongo2.RegisterFilter("toPrettyJSON", filterToPrettyJSON)
	pongo2.RegisterFilter("toYAML", filterToYAML)
	pongo2.RegisterFilter("dir", filterDir)
	pongo2.RegisterFilter("base", filterBase)
	pongo2.RegisterFilter("base64", filterBase64)
	pongo2.RegisterFilter("decrypt", filterDecrypt)
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

			filterFunc := pongoJSFilter(string(buf))

			if err := pongo2.RegisterFilter(name, filterFunc); err != nil {
				if err := pongo2.ReplaceFilter(name, filterFunc); err != nil {
					return errors.Errorf("couldn't replace existing filter %s", name)
				}
			}
		}
	}
	return nil
}

func pongoJSFilter(js string) func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		vm := goja.New()

		vm.Set("In", in.Interface())
		vm.Set("Param", param.Interface())

		v, err := vm.RunString(js)
		if err != nil {
			return nil, &pongo2.Error{
				Sender:    "filterToEnv",
				OrigError: err,
			}
		}

		return pongo2.AsValue(v.Export()), nil
	}
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
			Sender:    "filterToJSON",
			OrigError: err,
		}
	}
	return pongo2.AsValue(string(b)), nil
}

func filterToYAML(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	b, err := yaml_mapstr.Marshal(in.Interface())
	if err != nil {
		return nil, &pongo2.Error{
			Sender:    "filter:filterToYAML",
			OrigError: err,
		}
	}
	return pongo2.AsValue(string(b)), nil
}

func filterUnmarshalYAML(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !in.IsString() {
		return in, nil
	}

	var ret interface{}
	if err := yaml_mapstr.Unmarshal([]byte(in.String()), &ret); err != nil {
		return nil, &pongo2.Error{
			Sender:    "filterUnmarshalYAML",
			OrigError: err,
		}
	}

	return pongo2.AsValue(ret), nil
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

func filterDecrypt(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !param.IsString() {
		return in, nil
	}

	secretKeyring, err := os.Open(param.String())
	if err != nil {
		return nil, &pongo2.Error{
			Sender:    "filter:filterDecrypt",
			OrigError: err,
		}
	}

	defer secretKeyring.Close()
	entityList, err := openpgp.ReadArmoredKeyRing(secretKeyring)
	if err != nil {
		return nil, &pongo2.Error{
			Sender:    "filter:filterDecrypt",
			OrigError: err,
		}
	}

	input := in.Interface()
	switch i := input.(type) {
	case string:
		data, err := decrypt(i, entityList)
		if err != nil {
			return nil, &pongo2.Error{
				Sender:    "filter:filterDecrypt",
				OrigError: err,
			}
		}
		return pongo2.AsValue(data), nil
	case memkv.KVPairs:
		var new []memkv.KVPair
		for _, v := range i {
			dvalue, err := decrypt(v.Value, entityList)
			if err != nil {
				log.Warning(fmt.Sprintf("Couldn't decrypt `%s` - %s", v.Value, err))
			}
			new = append(new, memkv.KVPair{Key: v.Key, Value: dvalue})
		}
		return pongo2.AsValue(memkv.KVPairs(new)), nil
	case memkv.KVPair:
		dvalue, err := decrypt(i.Value, entityList)
		if err != nil {
			log.Warning(fmt.Sprintf("Couldn't decrypt `%s` - %s", i.Value, err))
		}
		return pongo2.AsValue(memkv.KVPair{Key: i.Key, Value: dvalue}), nil
	case []string:
		var new []string
		for _, v := range i {
			dvalue, err := decrypt(v, entityList)
			if err != nil {
				log.Warning(fmt.Sprintf("Couldn't decrypt `%s` - %s", v, err))
			}
			new = append(new, dvalue)
		}
		return pongo2.AsValue(new), nil
	}
	return in, nil
}
