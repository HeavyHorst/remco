package memkv

import (
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/derekparker/trie"
)

// A Store represents an in-memory key-value store safe for
// concurrent access.
type Store struct {
	FuncMap map[string]interface{}
	*sync.RWMutex
	t *trie.Trie
}

func New() Store {
	s := Store{
		t:       trie.New(),
		RWMutex: &sync.RWMutex{},
	}
	s.FuncMap = map[string]interface{}{
		"exists":    s.Exists,
		"ls":        s.List,
		"lsdir":     s.ListDir,
		"get":       s.Get,
		"gets":      s.GetAll,
		"getallkvs": s.GetAllKVs,
		"getv":      s.GetValue,
		"getvs":     s.GetAllValues,
	}
	return s
}

// Delete deletes the KVPair associated with key.
func (s Store) Del(key string) {
	s.Lock()
	s.t.Remove(key)
	s.Unlock()
}

// Exists checks for the existence of key in the store.
func (s Store) Exists(key string) bool {
	s.RLock()
	_, ok := s.t.Find(key)
	s.RUnlock()
	return ok
}

// Get gets the KVPair associated with key. If there is no KVPair
// associated with key, Get returns KVPair{}.
func (s Store) Get(key string) KVPair {
	s.RLock()
	defer s.RUnlock()
	node, ok := s.t.Find(key)
	if !ok {
		return KVPair{}
	}
	return node.Meta().(KVPair)
}

// GetAll returns a KVPair for all nodes with keys matching pattern.
// The syntax of patterns is the same as in path.Match.
func (s Store) GetAll(pattern string) KVPairs {
	ks := make(KVPairs, 0)
	s.RLock()
	defer s.RUnlock()
	for _, k := range s.t.Keys() {
		m, err := path.Match(pattern, k)
		if err != nil {
			return nil
		}
		kv := s.Get(k)
		if m {
			ks = append(ks, kv)
		}
	}
	if len(ks) == 0 {
		return nil
	}
	sort.Sort(ks)
	return ks
}

func (s Store) GetAllValues(pattern string) []string {
	vs := make([]string, 0)
	for _, kv := range s.GetAll(pattern) {
		vs = append(vs, kv.Value)
	}
	sort.Strings(vs)
	return vs
}

// GetAllKVs returns all KV-Pairs
func (s Store) GetAllKVs() KVPairs {
	ks := make(KVPairs, 0)
	s.RLock()
	defer s.RUnlock()
	for _, k := range s.t.Keys() {
		ks = append(ks, s.Get(k))
	}
	sort.Sort(ks)
	return ks
}

// GetValue gets the value associated with key. If there are no values
// associated with key, GetValue returns "".
func (s Store) GetValue(key string, v ...string) string {
	defaultValue := ""
	if len(v) > 0 {
		defaultValue = v[0]
	}
	kv := s.Get(key)
	if kv.Key == "" {
		return defaultValue
	}
	return kv.Value
}

func (s Store) list(filePath string, dir bool) []string {
	vs := make([]string, 0)
	m := make(map[string]bool)
	// The prefix search should only return dirs
	filePath = path.Clean(filePath) + "/"
	s.RLock()
	defer s.RUnlock()
	for _, k := range s.t.PrefixSearch(filePath) {
		items := strings.Split(stripKey(k, filePath), "/")
		if dir {
			if len(items) < 2 {
				continue
			}
		}
		m[items[0]] = true
	}
	for k := range m {
		vs = append(vs, k)
	}
	sort.Strings(vs)
	return vs
}

func (s Store) List(filePath string) []string {
	return s.list(filePath, false)
}

func (s Store) ListDir(filePath string) []string {
	return s.list(filePath, true)
}

// Set sets the KVPair entry associated with key to value.
func (s Store) Set(key string, value string) {
	s.Lock()
	s.t.Add(key, KVPair{key, value})
	s.Unlock()
}

func (s Store) Purge() {
	s.Lock()
	for _, k := range s.t.Keys() {
		s.t.Remove(k)
	}
	s.Unlock()
}

func stripKey(key, prefix string) string {
	return strings.TrimPrefix(strings.TrimPrefix(key, prefix), "/")
}
