package template

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/kelseyhightower/memkv"
)

func newFuncMap() map[string]interface{} {
	m := map[string]interface{}{
		"json":       UnmarshalJsonObject,
		"jsonArray":  UnmarshalJsonArray,
		"dir":        path.Dir,
		"base":       path.Base,
		"getenv":     Getenv,
		"contains":   strings.Contains,
		"replace":    strings.Replace,
		"lookupIP":   LookupIP,
		"lookupSRV":  LookupSRV,
		"fileExists": isFileExist,
		"printf":     fmt.Sprintf,
	}

	//already available in pongo2 ?
	//  m["join"] = strings.Join -- {{ value|join:" // " }}
	// 	m["toUpper"] = strings.ToUpper -- {{ value|upper }}
	//  m["toLower"] = strings.ToLower -- {{ value|lower }}
	//  m["datetime"] = time.Now  -- {% now "jS F Y H:i" %}

	pongo2.RegisterFilter("reverse", filterReverse)
	pongo2.RegisterFilter("sortByLength", filterSortByLength)
	pongo2.RegisterFilter("split", filterSplit)

	return m
}

func filterSplit(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !in.IsString() || !param.IsString() {
		return in, nil
	}
	return pongo2.AsValue(strings.Split(in.String(), param.String())), nil
}

type byLengthKV []memkv.KVPair

func (s byLengthKV) Len() int {
	return len(s)
}

func (s byLengthKV) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byLengthKV) Less(i, j int) bool {
	return len(s[i].Key) < len(s[j].Key)
}

type byLength []string

func (s byLength) Len() int {
	return len(s)
}
func (s byLength) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byLength) Less(i, j int) bool {
	return len(s[i]) < len(s[j])
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

// Getenv retrieves the value of the environment variable named by the key.
// It returns the value, which will the default value if the variable is not present.
// If no default value was given - returns "".
func Getenv(key string, v ...string) string {
	defaultValue := ""
	if len(v) > 0 {
		defaultValue = v[0]
	}

	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func UnmarshalJsonObject(data string) map[string]interface{} {
	var ret map[string]interface{}
	json.Unmarshal([]byte(data), &ret)
	return ret
}

func UnmarshalJsonArray(data string) []interface{} {
	var ret []interface{}
	json.Unmarshal([]byte(data), &ret)
	return ret
}

func LookupIP(data string) []string {
	ips, err := net.LookupIP(data)
	if err != nil {
		return nil
	}
	// "Cast" IPs into strings and sort the array
	ipStrings := make([]string, len(ips))

	for i, ip := range ips {
		ipStrings[i] = ip.String()
	}
	sort.Strings(ipStrings)
	return ipStrings
}

type sortSRV []*net.SRV

func (s sortSRV) Len() int {
	return len(s)
}

func (s sortSRV) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortSRV) Less(i, j int) bool {
	str1 := fmt.Sprintf("%s%d%d%d", s[i].Target, s[i].Port, s[i].Priority, s[i].Weight)
	str2 := fmt.Sprintf("%s%d%d%d", s[j].Target, s[j].Port, s[j].Priority, s[j].Weight)
	return str1 < str2
}

func LookupSRV(service, proto, name string) []*net.SRV {
	_, addrs, err := net.LookupSRV(service, proto, name)
	if err != nil {
		return []*net.SRV{}
	}
	sort.Sort(sortSRV(addrs))
	return addrs
}
