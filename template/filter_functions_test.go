package template

import (
	"testing"

	"github.com/HeavyHorst/memkv"
	"github.com/flosch/pongo2"
	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type FilterSuite struct{}

var _ = Suite(&FilterSuite{})

func (s *FilterSuite) TestFilterBase64(t *C) {
	in := pongo2.AsValue("foo")
	res, err := filterBase64(in, nil)
	if err != nil {
		t.Error(err.ErrorMsg)
	}

	t.Check(res.String(), Equals, "Zm9v")
}

func (s *FilterSuite) TestFilterBase(t *C) {
	in := pongo2.AsValue("/etc/foo/bar")
	res, err := filterBase(in, nil)
	if err != nil {
		t.Error(err.ErrorMsg)
	}

	t.Check(res.String(), Equals, "bar")
}

func (s *FilterSuite) TestFilterDir(t *C) {
	in := pongo2.AsValue("/etc/foo/bar")
	res, err := filterDir(in, nil)
	if err != nil {
		t.Error(err.ErrorMsg)
	}

	t.Check(res.String(), Equals, "/etc/foo")
}

func (s *FilterSuite) TestFilterSplit(t *C) {
	in := pongo2.AsValue("foo/bar")
	res, err := filterSplit(in, pongo2.AsValue("/"))
	if err != nil {
		t.Error(err.ErrorMsg)
	}

	t.Check(res.Interface().([]string), DeepEquals, []string{"foo", "bar"})
}

func (s *FilterSuite) TestFilterToPrettyJson(t *C) {
	expected := `{
    "test": "bla",
    "test2": 1,
    "test3": 2.5
}`
	in := pongo2.AsValue(map[string]interface{}{
		"test":  "bla",
		"test2": 1,
		"test3": 2.5,
	})
	res, err := filterToPrettyJson(in, nil)
	if err != nil {
		t.Error(err.ErrorMsg)
	}

	t.Check(res.String(), Equals, expected)
}

func (s *FilterSuite) TestFilterToJson(t *C) {
	expected := `{"test":"bla","test2":1,"test3":2.5}`
	in := pongo2.AsValue(map[string]interface{}{
		"test":  "bla",
		"test2": 1,
		"test3": 2.5,
	})
	res, err := filterToJson(in, nil)
	if err != nil {
		t.Error(err.ErrorMsg)
	}

	t.Check(res.String(), Equals, expected)
}

func (s *FilterSuite) TestFilterUnmarshalJSONObject(t *C) {
	in := pongo2.AsValue(`{"test":"bla","test2":"1","test3":"2.5"}`)
	expected := map[string]interface{}{
		"test":  "bla",
		"test2": "1",
		"test3": "2.5",
	}
	res, err := filterUnmarshalJSONObject(in, nil)
	if err != nil {
		t.Error(err.ErrorMsg)
	}
	m1 := res.Interface().(map[string]interface{})
	t.Check(m1, DeepEquals, expected)
}

func (s *FilterSuite) TestFilterUnmarshalJSONArray(t *C) {
	in := pongo2.AsValue(`["a", "b", "c"]`)
	expected := []interface{}{"a", "b", "c"}
	res, err := filterUnmarshalJSONArray(in, nil)
	if err != nil {
		t.Error(err.ErrorMsg)
	}
	m1 := res.Interface().([]interface{})
	t.Check(m1, DeepEquals, expected)
}

func (s *FilterSuite) TestFilterSortByLengthString(t *C) {
	in := pongo2.AsValue([]string{"123", "foobar", "1234"})
	expected := []string{"123", "1234", "foobar"}
	expectedRev := []string{"foobar", "1234", "123"}
	res, err := filterSortByLength(in, nil)
	if err != nil {
		t.Error(err.ErrorMsg)
	}
	m1 := res.Interface().([]string)
	t.Check(m1, DeepEquals, expected)

	rev, err := filterReverse(res, nil)
	if err != nil {
		t.Error(err.ErrorMsg)
	}
	m2 := rev.Interface().([]string)
	t.Check(m2, DeepEquals, expectedRev)
}

func (s *FilterSuite) TestFilterSortByLengthKVPair(t *C) {
	a := memkv.KVPair{Key: "123", Value: "Test"}
	b := memkv.KVPair{Key: "1234", Value: "Test"}
	c := memkv.KVPair{Key: "foobar", Value: "Test"}
	in := pongo2.AsValue(memkv.KVPairs{a, c, b})
	expected := memkv.KVPairs{a, b, c}
	expectedRev := memkv.KVPairs{c, b, a}
	res, err := filterSortByLength(in, nil)
	if err != nil {
		t.Error(err.ErrorMsg)
	}
	m1 := res.Interface().(memkv.KVPairs)
	t.Check(m1, DeepEquals, expected)

	rev, err := filterReverse(res, nil)
	if err != nil {
		t.Error(err.ErrorMsg)
	}
	m2 := rev.Interface().(memkv.KVPairs)
	t.Check(m2, DeepEquals, expectedRev)
}
