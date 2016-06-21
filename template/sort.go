package template

import (
	"fmt"
	"net"

	"github.com/kelseyhightower/memkv"
)

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
