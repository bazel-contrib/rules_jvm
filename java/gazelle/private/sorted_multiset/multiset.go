package sorted_multiset

import (
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"
	"github.com/google/btree"
)

type SortedMultiSet[K btree.Ordered, V any] struct {
	ms              map[K]*sorted_set.SortedSet[V]
	keys            *sorted_set.SortedSet[K]
	valueSetCreator func() *sorted_set.SortedSet[V]
}

func NewSortedMultiSet[K btree.Ordered, V btree.Ordered]() *SortedMultiSet[K, V] {
	return &SortedMultiSet[K, V]{
		ms:   make(map[K]*sorted_set.SortedSet[V]),
		keys: sorted_set.NewSortedSet([]K{}),
		valueSetCreator: func() *sorted_set.SortedSet[V] {
			return sorted_set.NewSortedSet([]V{})
		},
	}
}

func NewSortedMultiSetFn[K btree.Ordered, V any](less btree.LessFunc[V]) *SortedMultiSet[K, V] {
	return &SortedMultiSet[K, V]{
		ms:   make(map[K]*sorted_set.SortedSet[V]),
		keys: sorted_set.NewSortedSet([]K{}),
		valueSetCreator: func() *sorted_set.SortedSet[V] {
			return sorted_set.NewSortedSetFn([]V{}, less)
		},
	}
}

func (s *SortedMultiSet[K, V]) Add(key K, value V) {
	if !s.keys.Contains(key) {
		s.keys.Add(key)
		s.ms[key] = s.valueSetCreator()
	}
	s.ms[key].Add(value)
}

func (s *SortedMultiSet[K, V]) Keys() []K {
	if s == nil {
		return nil
	}

	return s.keys.SortedSlice()
}

func (s *SortedMultiSet[K, V]) Values(key K) *sorted_set.SortedSet[V] {
	if s == nil {
		return nil
	}

	return s.ms[key]
}

func (s *SortedMultiSet[K, V]) SortedValues(key K) []V {
	if s == nil {
		return nil
	}

	return s.ms[key].SortedSlice()
}
