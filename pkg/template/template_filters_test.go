/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package template

import (
	"testing"

	"github.com/HeavyHorst/memkv"
	"github.com/HeavyHorst/pongo2"
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
		t.Error(err.OrigError)
	}

	t.Check(res.String(), Equals, "Zm9v")
}

func (s *FilterSuite) TestFilterBase64Decode(t *C) {
	in := pongo2.AsValue("Zm9v")
	res, err := filterBase64Decode(in, nil)
	if err != nil {
		t.Error(err.OrigError)
	}

	t.Check(res.String(), Equals, "foo")
}

func (s *FilterSuite) TestFilterBase(t *C) {
	in := pongo2.AsValue("/etc/foo/bar")
	res, err := filterBase(in, nil)
	if err != nil {
		t.Error(err.OrigError)
	}

	t.Check(res.String(), Equals, "bar")
}

func (s *FilterSuite) TestFilterParseInt(t *C) {
	in := pongo2.AsValue("100")
	res, err := filterParseInt(in, nil)
	if err != nil {
		t.Error(err.OrigError)
	}

	t.Check(res.Integer(), Equals, 100)
}

func (s *FilterSuite) TestFilterParseFloat(t *C) {
	in := pongo2.AsValue("22.45")
	res, err := filterParseFloat(in, nil)
	if err != nil {
		t.Error(err.OrigError)
	}

	t.Check(res.Float(), Equals, 22.45)
}

func (s *FilterSuite) TestFilterDir(t *C) {
	in := pongo2.AsValue("/etc/foo/bar")
	res, err := filterDir(in, nil)
	if err != nil {
		t.Error(err.OrigError)
	}

	t.Check(res.String(), Equals, "/etc/foo")
}

func (s *FilterSuite) TestFilterToPrettyJSON(t *C) {
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
	res, err := filterToPrettyJSON(in, nil)
	if err != nil {
		t.Error(err.OrigError)
	}

	t.Check(res.String(), Equals, expected)
}

func (s *FilterSuite) TestFilterToJSON(t *C) {
	expected := `{"test":"bla","test2":1,"test3":2.5}`
	in := pongo2.AsValue(map[string]interface{}{
		"test":  "bla",
		"test2": 1,
		"test3": 2.5,
	})
	res, err := filterToJSON(in, nil)
	if err != nil {
		t.Error(err.OrigError)
	}

	t.Check(res.String(), Equals, expected)
}

func (s *FilterSuite) TestFilterToYAML(t *C) {
	expected := `test: bla
test2: 1
test3: 2.5
`
	in := pongo2.AsValue(map[string]interface{}{
		"test":  "bla",
		"test2": 1,
		"test3": 2.5,
	})
	res, err := filterToYAML(in, nil)
	if err != nil {
		t.Error(err.OrigError)
	}

	t.Check(res.String(), Equals, expected)
}

func (s *FilterSuite) TestFilterUnmarshalYAMLObject(t *C) {
	in := pongo2.AsValue(`{"test":"bla","test2":"1","test3":"2.5"}`)
	expected := map[string]interface{}{
		"test":  "bla",
		"test2": "1",
		"test3": "2.5",
	}
	res, err := filterUnmarshalYAML(in, nil)
	if err != nil {
		t.Error(err.OrigError)
	}
	m1 := res.Interface().(map[string]interface{})
	t.Check(m1, DeepEquals, expected)
}

func (s *FilterSuite) TestFilterUnmarshalYAMLArray(t *C) {
	in := pongo2.AsValue(`["a", "b", "c"]`)
	expected := []interface{}{"a", "b", "c"}
	res, err := filterUnmarshalYAML(in, nil)
	if err != nil {
		t.Error(err.OrigError)
	}
	m1 := res.Interface().([]interface{})
	t.Check(m1, DeepEquals, expected)
}

func (s *FilterSuite) TestFilterSortByLengthString(t *C) {
	in := pongo2.AsValue([]string{"123", "foobar", "1234"})
	expected := []string{"123", "1234", "foobar"}
	res, err := filterSortByLength(in, nil)
	if err != nil {
		t.Error(err.OrigError)
	}
	m1 := res.Interface().([]string)
	t.Check(m1, DeepEquals, expected)
}

func (s *FilterSuite) TestFilterSortByLengthKVPair(t *C) {
	a := memkv.KVPair{Key: "123", Value: "Test"}
	b := memkv.KVPair{Key: "1234", Value: "Test"}
	c := memkv.KVPair{Key: "foobar", Value: "Test"}
	in := pongo2.AsValue(memkv.KVPairs{a, c, b})
	expected := memkv.KVPairs{a, b, c}
	res, err := filterSortByLength(in, nil)
	if err != nil {
		t.Error(err.OrigError)
	}
	m1 := res.Interface().(memkv.KVPairs)
	t.Check(m1, DeepEquals, expected)
}

func (s *FilterSuite) TestFilterMapValue(t *C) {
	a := map[string]int{
		"hallo": 1,
		"moin":  2,
	}
	fm, err := filterMapValue(pongo2.AsValue(a), pongo2.AsValue("moin"))
	if err != nil {
		t.Error(err.OrigError)
	}
	t.Check(fm.Interface(), DeepEquals, 2)

	b := map[string]string{
		"servus": "one",
		"hi":     "oneone",
	}
	fm, err = filterMapValue(pongo2.AsValue(b), pongo2.AsValue("servus"))
	if err != nil {
		t.Error(err.OrigError)
	}
	t.Check(fm.Interface(), DeepEquals, "one")

	c := map[string]interface{}{
		"servus": "one",
		"hi":     100,
	}
	fm, err = filterMapValue(pongo2.AsValue(c), pongo2.AsValue("hi"))
	if err != nil {
		t.Error(err.OrigError)
	}
	t.Check(fm.Interface(), DeepEquals, 100)

	d := map[int]string{
		1: "one",
		2: "oneone",
	}
	fm, err = filterMapValue(pongo2.AsValue(d), pongo2.AsValue(2))
	if err != nil {
		t.Error(err.OrigError)
	}
	t.Check(fm.Interface(), DeepEquals, "oneone")
}

func (s *FilterSuite) TestFilterIndex(t *C) {
	in := pongo2.AsValue([]string{"Hallo", "Test", "123", "Moin"})
	expected := "123"

	res, err := filterIndex(in, pongo2.AsValue(2))
	if err != nil {
		t.Error(err.OrigError)
	}

	// test negative index
	res2, err := filterIndex(in, pongo2.AsValue(-2))
	if err != nil {
		t.Error(err.OrigError)
	}

	m1 := res.String()
	m2 := res2.String()
	t.Check(m1, DeepEquals, expected)
	t.Check(m2, DeepEquals, expected)
}
