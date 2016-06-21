package template

import (
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
)

func newFuncMap() map[string]interface{} {
	m := map[string]interface{}{
		"getenv":     Getenv,
		"contains":   strings.Contains,
		"replace":    strings.Replace,
		"lookupIP":   LookupIP,
		"lookupSRV":  LookupSRV,
		"fileExists": isFileExist,
		"printf":     fmt.Sprintf,
	}

	initFilters()

	return m
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

func LookupSRV(service, proto, name string) []*net.SRV {
	_, addrs, err := net.LookupSRV(service, proto, name)
	if err != nil {
		return []*net.SRV{}
	}
	sort.Sort(sortSRV(addrs))
	return addrs
}
