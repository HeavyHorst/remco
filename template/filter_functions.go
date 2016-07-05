package template

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path"
	"sort"
	"strings"

	"golang.org/x/crypto/openpgp"

	"github.com/HeavyHorst/memkv"
	"github.com/cloudflare/cfssl/log"
	"github.com/flosch/pongo2"
)

func init() {
	pongo2.RegisterFilter("reverse", filterReverse)
	pongo2.RegisterFilter("sortByLength", filterSortByLength)
	pongo2.RegisterFilter("split", filterSplit)
	pongo2.RegisterFilter("parseJson", filterUnmarshalJSONObject)
	pongo2.RegisterFilter("parseJsonArray", filterUnmarshalJSONArray)
	pongo2.RegisterFilter("toJson", filterToJson)
	pongo2.RegisterFilter("toPrettyJson", filterToPrettyJson)
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

func filterSplit(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !in.IsString() || !param.IsString() {
		return in, nil
	}
	return pongo2.AsValue(strings.Split(in.String(), param.String())), nil
}

func filterToPrettyJson(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	b, err := json.MarshalIndent(in.Interface(), "", "    ")
	if err != nil {
		return nil, &pongo2.Error{ErrorMsg: err.Error()}
	}
	return pongo2.AsSafeValue(string(b)), nil
}

func filterToJson(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	b, err := json.Marshal(in.Interface())
	if err != nil {
		return nil, &pongo2.Error{ErrorMsg: err.Error()}
	}
	return pongo2.AsSafeValue(string(b)), nil
}

func filterUnmarshalJSONObject(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !in.IsString() {
		return in, nil
	}

	var ret map[string]interface{}
	if err := json.Unmarshal([]byte(in.String()), &ret); err != nil {
		log.Warning(err)
	}
	return pongo2.AsValue(ret), nil
}

func filterUnmarshalJSONArray(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !in.IsString() {
		return in, nil
	}

	var ret []interface{}
	if err := json.Unmarshal([]byte(in.String()), &ret); err != nil {
		log.Warning(err)
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

//Reverse returns the array in reversed order
//works with []string and []KVPair
func filterReverse(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !in.CanSlice() {
		return in, nil
	}

	values := in.Interface()
	switch values.(type) {
	case []string:
		v := values.([]string)
		for left, right := 0, len(v)-1; left < right; left, right = left+1, right-1 {
			v[left], v[right] = v[right], v[left]
		}
	case memkv.KVPairs:
		v := values.(memkv.KVPairs)
		for left, right := 0, len(v)-1; left < right; left, right = left+1, right-1 {
			v[left], v[right] = v[right], v[left]
		}
	}

	return pongo2.AsValue(values), nil
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
				log.Warningf("Couldn't decrypt `%s` - %s", v.Value, err)
			}
			new = append(new, memkv.KVPair{Key: v.Key, Value: dvalue})
		}
		return pongo2.AsValue(memkv.KVPairs(new)), nil
	case memkv.KVPair:
		i := input.(memkv.KVPair)
		dvalue, err := decrypt(i.Value, entityList)
		if err != nil {
			log.Warningf("Couldn't decrypt `%s` - %s", i.Value, err)
		}
		return pongo2.AsValue(memkv.KVPair{Key: i.Key, Value: dvalue}), nil
	case []string:
		i := input.([]string)
		var new []string
		for _, v := range i {
			dvalue, err := decrypt(v, entityList)
			if err != nil {
				log.Warningf("Couldn't decrypt `%s` - %s", v, err)
			}
			new = append(new, dvalue)
		}
		return pongo2.AsValue(new), nil
	}
	return in, nil
}
