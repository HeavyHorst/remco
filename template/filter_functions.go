package template

import (
	"encoding/base64"
	"encoding/json"
	"path"
	"sort"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/kelseyhightower/memkv"
)

func initFilters() {
	pongo2.RegisterFilter("reverse", filterReverse)
	pongo2.RegisterFilter("sortByLength", filterSortByLength)
	pongo2.RegisterFilter("split", filterSplit)
	pongo2.RegisterFilter("json", filterUnmarshalJSONObject)
	pongo2.RegisterFilter("jsonArray", filterUnmarshalJSONArray)
	pongo2.RegisterFilter("dir", filterDir)
	pongo2.RegisterFilter("base", filterBase)
	pongo2.RegisterFilter("base64", filterBase64)
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

func filterUnmarshalJSONObject(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !in.IsString() {
		return in, nil
	}

	var ret map[string]interface{}
	json.Unmarshal([]byte(in.String()), &ret)
	return pongo2.AsValue(ret), nil
}

func filterUnmarshalJSONArray(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !in.IsString() {
		return in, nil
	}

	var ret []interface{}
	json.Unmarshal([]byte(in.String()), &ret)
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
	case []memkv.KVPair:
		v := values.([]memkv.KVPair)
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
	case []memkv.KVPair:
		v := values.([]memkv.KVPair)
		for left, right := 0, len(v)-1; left < right; left, right = left+1, right-1 {
			v[left], v[right] = v[right], v[left]
		}
	}

	return pongo2.AsValue(values), nil
}
