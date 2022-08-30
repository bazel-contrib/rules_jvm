package sorted_multiset_test

import (
	"reflect"
	"testing"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_multiset"
)

func TestNil(t *testing.T) {
	var s *sorted_multiset.SortedMultiSet[string, string]

	gotKeys := s.Keys()
	if gotKeysLen := len(gotKeys); gotKeysLen != 0 {
		t.Errorf("want len(nil.Keys()) == 0, got %d", gotKeysLen)
	}

	gotValues := s.Values("foo")
	if gotValuesLen := len(gotValues); gotValuesLen != 0 {
		t.Errorf("want len(nil.Values(\"foo\")) == 0, got %d", gotValuesLen)
	}
}

func TestMultiSet(t *testing.T) {
	s := sorted_multiset.NewSortedMultiSet[string, string]()

	{
		gotKeys := s.Keys()
		if gotKeysLen := len(gotKeys); gotKeysLen != 0 {
			t.Errorf("want no keys, got %d: %v", gotKeysLen, gotKeys)
		}

		gotValues := s.Values("tasty")
		if gotValuesLen := len(gotValues); gotValuesLen != 0 {
			t.Errorf("want no values for tasty, got %d: %v", gotValuesLen, gotValues)
		}
	}

	s.Add("tasty", "hummus")
	{
		wantKeys := []string{"tasty"}
		gotKeys := s.Keys()
		if !reflect.DeepEqual(wantKeys, gotKeys) {
			t.Errorf("want keys %v got %v", wantKeys, gotKeys)
		}

		wantTastyValues := []string{"hummus"}
		gotTastyValues := s.Values("tasty")
		if !reflect.DeepEqual(wantTastyValues, gotTastyValues) {
			t.Errorf("want tasty values %v got %v", wantTastyValues, gotTastyValues)
		}

		gotBadValues := s.Values("bad")
		if gotBadValuesLen := len(gotBadValues); gotBadValuesLen != 0 {
			t.Errorf("want no values for tasty, got %d: %v", gotBadValuesLen, gotBadValues)
		}
	}

	s.Add("tasty", "cheese")
	{
		wantKeys := []string{"tasty"}
		gotKeys := s.Keys()
		if !reflect.DeepEqual(wantKeys, gotKeys) {
			t.Errorf("want keys %v got %v", wantKeys, gotKeys)
		}

		wantTastyValues := []string{"cheese", "hummus"}
		gotTastyValues := s.Values("tasty")
		if !reflect.DeepEqual(wantTastyValues, gotTastyValues) {
			t.Errorf("want tasty values %v got %v", wantTastyValues, gotTastyValues)
		}

		gotBadValues := s.Values("bad")
		if gotBadValuesLen := len(gotBadValues); gotBadValuesLen != 0 {
			t.Errorf("want no values for tasty, got %d: %v", gotBadValuesLen, gotBadValues)
		}
	}

	s.Add("tasty", "cheese")
	{
		wantKeys := []string{"tasty"}
		gotKeys := s.Keys()
		if !reflect.DeepEqual(wantKeys, gotKeys) {
			t.Errorf("want keys %v got %v", wantKeys, gotKeys)
		}

		wantTastyValues := []string{"cheese", "hummus"}
		gotTastyValues := s.Values("tasty")
		if !reflect.DeepEqual(wantTastyValues, gotTastyValues) {
			t.Errorf("want tasty values %v got %v", wantTastyValues, gotTastyValues)
		}

		gotBadValues := s.Values("bad")
		if gotBadValuesLen := len(gotBadValues); gotBadValuesLen != 0 {
			t.Errorf("want no values for tasty, got %d: %v", gotBadValuesLen, gotBadValues)
		}
	}

	s.Add("bad", "soil")
	{
		wantKeys := []string{"bad", "tasty"}
		gotKeys := s.Keys()
		if !reflect.DeepEqual(wantKeys, gotKeys) {
			t.Errorf("want keys %v got %v", wantKeys, gotKeys)
		}

		wantTastyValues := []string{"cheese", "hummus"}
		gotTastyValues := s.Values("tasty")
		if !reflect.DeepEqual(wantTastyValues, gotTastyValues) {
			t.Errorf("want tasty values %v got %v", wantTastyValues, gotTastyValues)
		}

		wantBadValues := []string{"soil"}
		gotBadValues := s.Values("bad")
		if !reflect.DeepEqual(wantBadValues, gotBadValues) {
			t.Errorf("want bad values %v got %v", wantBadValues, gotBadValues)
		}
	}
}
