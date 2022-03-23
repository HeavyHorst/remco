/*
 * This file is part of remco.
 * Based on code from confd.
 * https://github.com/kelseyhightower/confd/blob/6bb3c21a63459c3be340d53c4d3463397c8324c6/resource/template/template_funcs.go
 * © 2013 Kelsey Hightower
 * © 2015 Justin Burnham
 * © 2016 odedlaz
 *
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package template

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/HeavyHorst/remco/pkg/template/fileutil"
)

type interfaceSet map[string]struct{}

func (s interfaceSet) Append(value interface{}) string {
	v := fmt.Sprintf("%v", value)
	s[v] = struct{}{}
	return ""
}

func (s interfaceSet) Remove(value string) string {
	v := fmt.Sprintf("%v", value)
	delete(s, v)
	return ""
}

func (s interfaceSet) Contains(value interface{}) bool {
	v := fmt.Sprintf("%v", value)
	_, c := s[v]
	return c
}

func (s interfaceSet) SortedSet() []string {
	var i []string
	for k := range s {
		i = append(i, k)
	}
	sort.Strings(i)
	return i
}

func (s interfaceSet) MarshalYAML() (interface{}, error) {
	return s.SortedSet(), nil
}

func (s interfaceSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.SortedSet())
}

type templateMap map[string]interface{}

func (t templateMap) Set(key string, value interface{}) string {
	t[key] = value
	return ""
}

func (t templateMap) Remove(key string) string {
	delete(t, key)
	return ""
}

func (t templateMap) Get(key string) interface{} {
	return t[key]
}

func newFuncMap() map[string]interface{} {
	m := map[string]interface{}{
		"getenv":      getenv,
		"contains":    strings.Contains,
		"replace":     strings.Replace,
		"lookupIP":    lookupIP,
		"lookupSRV":   lookupSRV,
		"fileExists":  fileutil.IsFileExist,
		"printf":      fmt.Sprintf,
		"unixTS":      unixTimestampNow,
		"dateRFC3339": dateRFC3339Now,
		"createMap":   createMap,
		"createSet":   createSet,
	}

	return m
}

func addFuncs(out, in map[string]interface{}) {
	for name, fn := range in {
		out[name] = fn
	}
}

// Getenv retrieves the value of the environment variable named by the key.
// It returns the value, which will the default value if the variable is not present.
// If no default value was given - returns "".
func getenv(key string, v ...string) string {
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

func lookupIP(data string) ([]string, error) {
	ips, err := net.LookupIP(data)
	if err != nil {
		return nil, err
	}
	// "Cast" IPs into strings and sort the array
	ipStrings := make([]string, len(ips))

	for i, ip := range ips {
		ipStrings[i] = ip.String()
	}
	sort.Strings(ipStrings)
	return ipStrings, nil
}

func createMap() templateMap {
	tm := make(map[string]interface{})
	return tm
}

func createSet() interfaceSet {
	return make(map[string]struct{})
}

func lookupSRV(service, proto, name string) ([]*net.SRV, error) {
	_, addrs, err := net.LookupSRV(service, proto, name)
	if err != nil {
		return nil, err
	}
	sort.Slice(addrs, func(i, j int) bool {
		str1 := fmt.Sprintf("%s%d%d%d", addrs[i].Target, addrs[i].Port, addrs[i].Priority, addrs[i].Weight)
		str2 := fmt.Sprintf("%s%d%d%d", addrs[j].Target, addrs[j].Port, addrs[j].Priority, addrs[j].Weight)
		return str1 < str2
	})
	return addrs, nil
}

func unixTimestampNow() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func dateRFC3339Now() string {
	return time.Now().Format(time.RFC3339)
}
