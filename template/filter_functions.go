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
	"os"
	"path"
	"sort"

	"gopkg.in/yaml.v2"

	"golang.org/x/crypto/openpgp"

	"github.com/HeavyHorst/memkv"
	"github.com/HeavyHorst/pongo2"
	"github.com/HeavyHorst/remco/log"
)

func init() {
	pongo2.RegisterFilter("sortByLength", filterSortByLength)
	pongo2.RegisterFilter("parseYAML", filterUnmarshalYAML)
	pongo2.RegisterFilter("parseYAMLArray", filterUnmarshalYAML) //deprecated
	pongo2.RegisterFilter("toJSON", filterToJSON)
	pongo2.RegisterFilter("toPrettyJSON", filterToPrettyJSON)
	pongo2.RegisterFilter("toYAML", filterToYAML)
	pongo2.RegisterFilter("dir", filterDir)
	pongo2.RegisterFilter("base", filterBase)
	pongo2.RegisterFilter("base64", filterBase64)
	pongo2.RegisterFilter("decrypt", filterDecrypt)
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
		return nil, &pongo2.Error{ErrorMsg: err.Error()}
	}
	return pongo2.AsValue(string(b)), nil
}

func filterToJSON(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	b, err := json.Marshal(in.Interface())
	if err != nil {
		return nil, &pongo2.Error{ErrorMsg: err.Error()}
	}
	return pongo2.AsValue(string(b)), nil
}

func filterToYAML(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	b, err := yaml.Marshal(in.Interface())
	if err != nil {
		return nil, &pongo2.Error{ErrorMsg: err.Error()}
	}
	return pongo2.AsValue(string(b)), nil
}

func filterUnmarshalYAML(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !in.IsString() {
		return in, nil
	}

	var ret interface{}
	if err := yaml.Unmarshal([]byte(in.String()), &ret); err != nil {
		return nil, &pongo2.Error{ErrorMsg: err.Error()}
	}
	return pongo2.AsValue(ret), nil
}

func filterSortByLength(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !in.CanSlice() {
		return in, nil
	}

	values := in.Interface()
	switch values.(type) {
	case []string:
		v := values.([]string)
		sort.Sort(byLength(v))
		return pongo2.AsValue(v), nil
	case memkv.KVPairs:
		v := values.(memkv.KVPairs)
		sort.Sort(byLengthKV(v))
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
		return nil, &pongo2.Error{ErrorMsg: err.Error()}
	}

	defer secretKeyring.Close()
	entityList, err := openpgp.ReadArmoredKeyRing(secretKeyring)
	if err != nil {
		return nil, &pongo2.Error{ErrorMsg: err.Error()}
	}

	input := in.Interface()
	switch input.(type) {
	case string:
		i := input.(string)
		data, err := decrypt(i, entityList)
		if err != nil {
			return nil, &pongo2.Error{ErrorMsg: err.Error()}
		}
		return pongo2.AsValue(data), nil
	case memkv.KVPairs:
		i := input.(memkv.KVPairs)
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
		i := input.(memkv.KVPair)
		dvalue, err := decrypt(i.Value, entityList)
		if err != nil {
			log.Warning(fmt.Sprintf("Couldn't decrypt `%s` - %s", i.Value, err))
		}
		return pongo2.AsValue(memkv.KVPair{Key: i.Key, Value: dvalue}), nil
	case []string:
		i := input.([]string)
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
