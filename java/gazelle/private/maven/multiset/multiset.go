package multiset

import (
	"sort"
	"sync"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/cmd/parsejars/manifest"
)

type stringSet = map[string]struct{}

func newStringSet() stringSet {
	return make(map[string]struct{})
}

type StringMultiSet struct {
	lock sync.RWMutex
	data map[string]stringSet
}

func NewStringMultiSet() *StringMultiSet {
	return &StringMultiSet{data: make(map[string]stringSet)}
}

func (m *StringMultiSet) Add(key, value string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, found := m.data[key]; !found {
		m.data[key] = newStringSet()
	}
	m.data[key][value] = struct{}{}
}

func (m *StringMultiSet) Get(key string) (map[string]struct{}, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	v, ok := m.data[key]
	return v, ok
}

func (m *StringMultiSet) DumpManifest() *manifest.Manifest {
	m.lock.RLock()
	defer m.lock.RUnlock()

	out := manifest.Manifest{
		ArtifactsMapping: make(map[string][]string),
	}

	for k, v := range m.data {
		if k == "" {
			continue
		}
		var values []string
		for vv := range v {
			values = append(values, vv)
		}
		sort.Strings(values)
		out.ArtifactsMapping[k] = values
	}

	return &out
}
