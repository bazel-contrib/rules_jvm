package multiset

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestStringMultiSet_Add(t *testing.T) {
	tests := map[string]struct {
		add  func(string, *StringMultiSet)
		want map[string]struct{}
	}{
		"simple add": {
			add: func(key string, ms *StringMultiSet) {
				ms.Add(key, "bar")
			},
			want: map[string]struct{}{
				"bar": {},
			},
		},
		"double add": {
			add: func(key string, ms *StringMultiSet) {
				ms.Add(key, "bar")
				ms.Add(key, "bar")
			},
			want: map[string]struct{}{
				"bar": {},
			},
		},
		"add 2 things": {
			add: func(key string, ms *StringMultiSet) {
				ms.Add(key, "bar")
				ms.Add(key, "baz")
			},
			want: map[string]struct{}{
				"bar": {},
				"baz": {},
			},
		},
		"add on 2 keys": {
			add: func(key string, ms *StringMultiSet) {
				ms.Add(key, "bar")
				ms.Add("not"+key, "bar")
			},
			want: map[string]struct{}{
				"bar": {},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ms := NewStringMultiSet()
			const key = "foo"
			tt.add(key, ms)
			got, ok := ms.Get(key)
			if !ok {
				t.Fatal("expected to find key")
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func BenchmarkAdd(b *testing.B) {
	ms := NewStringMultiSet()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ms.Add("foo", "bar")
	}
}

func BenchmarkGet(b *testing.B) {
	ms := NewStringMultiSet()
	ms.Add("foo", "bar")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ms.Get("foo")
	}
}
