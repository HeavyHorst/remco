package memkv

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

var gettests = []struct {
	key    string
	value  string
	want   KVPair
	exists bool
}{
	{"/db/user", "admin", KVPair{"/db/user", "admin"}, true},
	{"/db/pass", "foo", KVPair{"/db/pass", "foo"}, true},
	{"/missing", "", KVPair{}, false},
}

func TestGet(t *testing.T) {
	for _, tt := range gettests {
		s := New()
		if tt.value != "" {
			s.Set(tt.key, tt.value)
		}
		got, err := s.Get(tt.key)
		if err != nil {
			if tt.exists {
				t.Error(err)
			}
		}
		if got != tt.want {
			t.Errorf("Get(%q) = %v, want %v", tt.key, got, tt.want)
		}

		e := s.Exists(tt.key)
		if e != tt.exists {
			t.Errorf("Get(%q) = %v, want %v", tt.key, e, tt.exists)
		}
	}
}

var getvtests = []struct {
	key      string
	value    string
	want     string
	existing bool
}{
	{"/db/user", "admin", "admin", true},
	{"/db/pass", "foo", "foo", true},
	{"/missing", "", "", false},
}

func TestGetValue(t *testing.T) {
	for _, tt := range getvtests {
		s := New()
		if tt.existing {
			s.Set(tt.key, tt.value)
		}

		got, err := s.GetValue(tt.key)
		if err != nil {
			if tt.existing {
				t.Error(err)
			}

			err, ok := err.(*KeyError)
			if !ok {
				t.Error(err)
			}

			if err.Err != ErrNotExist {
				t.Error(err)
			}

		}
		if got != tt.want {
			t.Errorf("Get(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestGetValueWithDefault(t *testing.T) {
	want := "defaultValue"
	s := New()
	got, err := s.GetValue("/db/user", "defaultValue")
	if err != nil {
		t.Error(err)
	}
	if got != want {
		t.Errorf("want %v, got %v", want, got)
	}
}

var getalltestinput = map[string]string{
	"/app/db/pass":               "foo",
	"/app/db/user":               "admin",
	"/app/port":                  "443",
	"/app/url":                   "app.example.com",
	"/app/vhosts/host1":          "app.example.com",
	"/app/upstream/host1":        "203.0.113.0.1:8080",
	"/app/upstream/host1/domain": "app.example.com",
	"/app/upstream/host2":        "203.0.113.0.2:8080",
	"/app/upstream/host2/domain": "app.example.com",
}

var getalltests = []struct {
	pattern string
	want    []KVPair
}{
	{"/app/db/*",
		[]KVPair{
			KVPair{"/app/db/pass", "foo"},
			KVPair{"/app/db/user", "admin"}}},
	{"/app/*/host1",
		[]KVPair{
			KVPair{"/app/upstream/host1", "203.0.113.0.1:8080"},
			KVPair{"/app/vhosts/host1", "app.example.com"}}},

	{"/app/upstream/*",
		[]KVPair{
			KVPair{"/app/upstream/host1", "203.0.113.0.1:8080"},
			KVPair{"/app/upstream/host2", "203.0.113.0.2:8080"}}},
	{"[]a]", nil},
}

func TestGetAll(t *testing.T) {
	s := New()
	for k, v := range getalltestinput {
		s.Set(k, v)
	}
	for _, tt := range getalltests {
		got, err := s.GetAll(tt.pattern)
		if err != nil && tt.want != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual([]KVPair(got), []KVPair(tt.want)) {
			t.Errorf("GetAll(%q) = %v, want %v", tt.pattern, got, tt.want)
		}
	}
}

func TestGetAllValues(t *testing.T) {
	s := New()
	for k, v := range getalltestinput {
		s.Set(k, v)
	}
	for _, tt := range getalltests {
		var want []string
		got, err := s.GetAllValues(tt.pattern)
		if err != nil && tt.want != nil {
			t.Error(err)
		}
		if tt.want != nil {
			want = []string{tt.want[0].Value, tt.want[1].Value}
			sort.Strings(want)
		}
		if !reflect.DeepEqual([]string(got), []string(want)) {
			t.Errorf("GetAll(%q) = %v, want %v", tt.pattern, got, want)
		}
	}
}

func TestGetAllKvs(t *testing.T) {
	s := New()
	want := make(KVPairs, 0)
	for k, v := range getalltestinput {
		s.Set(k, v)
		want = append(want, KVPair{k, v})
	}
	sort.Sort(want)
	got := s.GetAllKVs()
	if !reflect.DeepEqual([]KVPair(got), []KVPair(want)) {
		t.Errorf("GetAllKVs() = %v, want %v", got, want)
	}
}

func TestDel(t *testing.T) {
	s := New()
	s.Set("/app/port", "8080")
	want := KVPair{"/app/port", "8080"}
	got, err := s.Get("/app/port")
	if err != nil {
		t.Error(err)
	}
	if got != want {
		t.Errorf("Get(%q) = %v, want %v", "/app/port", got, want)
	}
	s.Del("/app/port")
	want = KVPair{}
	got, err = s.Get("/app/port")
	if err != nil {
		err, ok := err.(*KeyError)
		if !ok {
			t.Error(err)
		}

		if err.Err != ErrNotExist {
			t.Error(err)
		}
	}
	if got != want {
		t.Errorf("Get(%q) = %v, want %v", "/app/port", got, want)
	}
}

func TestPurge(t *testing.T) {
	s := New()
	s.Set("/app/port", "8080")
	want := KVPair{"/app/port", "8080"}
	got, err := s.Get("/app/port")
	if err != nil {
		t.Error(err)
	}
	if got != want {
		t.Errorf("Get(%q) = %v, want %v", "/app/port", got, want)
	}
	s.Purge()
	want = KVPair{}
	got, err = s.Get("/app/port")
	if err != nil {
		err, ok := err.(*KeyError)
		if !ok {
			t.Error(err)
		}

		if err.Err != ErrNotExist {
			t.Error(err)
		}
	}
	if got != want {
		t.Errorf("Get(%q) = %v, want %v", "/app/port", got, want)
	}
	s.Set("/app/port", "8080")
	want = KVPair{"/app/port", "8080"}
	got, err = s.Get("/app/port")
	if err != nil {
		t.Error(err)
	}
	if got != want {
		t.Errorf("Get(%q) = %v, want %v", "/app/port", got, want)
	}
}

var listTestMap = map[string]string{
	"/deis/database/user":             "user",
	"/deis/database/pass":             "pass",
	"/deis/services/key":              "value",
	"/deis/services/notaservice/foo":  "bar",
	"/deis/services/srv1/node1":       "10.244.1.1:80",
	"/deis/services/srv1/node2":       "10.244.1.2:80",
	"/deis/services/srv1/node3":       "10.244.1.3:80",
	"/deis/services/srv2/node1":       "10.244.2.1:80",
	"/deis/services/srv2/node2":       "10.244.2.2:80",
	"/deis/prefix/node1":              "prefix_node1",
	"/deis/prefix/node2/leafnode":     "prefix_node2",
	"/deis/prefix/node3/leafnode":     "prefix_node3",
	"/deis/prefix_a/node4":            "prefix_a_node4",
	"/deis/prefixb/node5/leafnode":    "prefixb_node5",
	"/deis/dirprefix/node1":           "prefix_node1",
	"/deis/dirprefix/node2/leafnode":  "prefix_node2",
	"/deis/dirprefix/node3/leafnode":  "prefix_node3",
	"/deis/dirprefix_a/node4":         "prefix_a_node4",
	"/deis/dirprefixb/node5/leafnode": "prefixb_node5",
	"/deis/prefix/node2/sub1/leaf1":   "prefix_node2_sub1_leaf1",
	"/deis/prefix/node2/sub1/leaf2":   "prefix_node2_sub1_leaf2",
}

func testList(t *testing.T, want, paths []string, dir bool) {
	var got []string
	s := New()
	for k, v := range listTestMap {
		s.Set(k, v)
	}
	for _, filePath := range paths {
		if dir {
			got = s.ListDir(filePath)
		} else {
			got = s.List(filePath)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("List(%s) = %v, want %v", filePath, got, want)
		}
	}
}

func TestList(t *testing.T) {
	want := []string{"key", "notaservice", "srv1", "srv2"}
	paths := []string{
		"/deis/services",
		"/deis/services/",
	}
	testList(t, want, paths, false)
}

func TestListForFile(t *testing.T) {
	var want []string
	paths := []string{"/deis/services/key"}
	testList(t, want, paths, false)
}

func TestListDir(t *testing.T) {
	want := []string{"notaservice", "srv1", "srv2"}
	paths := []string{
		"/deis/services",
		"/deis/services/",
	}
	testList(t, want, paths, true)
}

func TestListForSamePrefix(t *testing.T) {
	want := []string{"node1", "node2", "node3"}
	paths := []string{
		"/deis/prefix",
		"/deis/prefix/",
	}
	testList(t, want, paths, false)
}

func TestListDirForSamePrefix(t *testing.T) {
	want := []string{"node2", "node3"}
	paths := []string{
		"/deis/dirprefix",
		"/deis/dirprefix/",
	}
	testList(t, want, paths, true)
}

func TestListForMixedLeafSubnodes(t *testing.T) {
	want := []string{"leaf1", "leaf2"}
	paths := []string{"/deis/prefix/node2/sub1"}
	testList(t, want, paths, false)
}

func BenchmarkSet(b *testing.B) {
	s := New()
	for n := 0; n < b.N; n++ {
		st := fmt.Sprintf("%d", n)
		s.Set(st, st)
	}
}

var GetResult KVPair

func BenchmarkGet(b *testing.B) {
	var err error
	s := New()
	s.Set("hallomoin", "hallomoin")
	for k, v := range listTestMap {
		s.Set(k, v)
	}

	var kv KVPair
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		kv, err = s.Get("hallomoin")
		if err != nil {
			b.Error(err)
		}
	}

	if kv.Value != "hallomoin" {
		b.Error("unexpected result: " + kv.Value)
	}

	GetResult = kv
}

var GetValueResult string

func BenchmarkGetValue(b *testing.B) {
	var err error
	s := New()
	s.Set("hallomoin", "hallomoin")
	for k, v := range listTestMap {
		s.Set(k, v)
	}

	var v string
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		v, err = s.GetValue("hallomoin")
		if err != nil {
			b.Error(err)
		}
	}

	if v != "hallomoin" {
		b.Error("unexpected result: " + v)
	}

	GetValueResult = v
}

var ListResult []string

func benchmarkList(b *testing.B, dir bool) {
	s := New()

	for k, v := range listTestMap {
		s.Set(k, v)
	}

	var v []string
	b.ResetTimer()
	if !dir {
		for n := 0; n < b.N; n++ {
			v = s.List("/deis/services")
		}
	} else {
		for n := 0; n < b.N; n++ {
			v = s.ListDir("/deis/services")
		}
	}

	ListResult = v
}

func BenchmarkList(b *testing.B) {
	benchmarkList(b, false)
}

func BenchmarkListDir(b *testing.B) {
	benchmarkList(b, true)
}
