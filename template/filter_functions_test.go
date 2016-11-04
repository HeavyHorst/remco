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

const (
	secret = `wcBMA2d6jjCDCcXCAQgARmuHp37r8h9NHylLytnCfazCKTwJtG6eIcXH/BC8ACvYfx/yKqcj9vf5v7TCI+xCko3cBCZ3+Y/3b0497niK1qioqLBrWGiPhee/iwZ/fEXDjlC1w4Mzqa+GrDHsUwhKIbwtKJCeNXbA7SkJFm0tFGdA1RLKhsuWFmvtwtRbuMc9c73Lq89uzF9fWARcNF0GaGZtkk3Sui6GrtrJc7gZqLqKDhgrR1uwlpBKyZh5us6rSh9cBmfDB1BYvO9q3Ywz4pKA9yEtPFsnrqSxeFPBPrGhap7RJdzHB28Ysx6ogmhuDNNSNuwjQNyZn1JA0zDWuzg+Lo3rjsDpYiE2EK2EHNLgAeT3mYEW4p3QYSXY4PLd6cHx4ZSF4Crgz+HNq+BT4qo4svzgsOPds1LhZ6z9quAu4RKe4L7j86ZluQaud/vgneH86eDQ4hTm86DgkeO65b++2bBQOeDK5FB2EXstONJS6qnN/dyqNZzivXJHLeE2RgA=`
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
		t.Error(err.ErrorMsg)
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
		t.Error(err.ErrorMsg)
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

func (s *FilterSuite) TestFilterDecryptString(t *C) {
	in := pongo2.AsValue(secret)
	expected := "secret1\n"
	res, err := filterDecrypt(in, pongo2.AsValue("../integration/config/test.gpg"))
	if err != nil {
		t.Error(err.ErrorMsg)
		t.FailNow()
	}
	t.Check(res.String(), Equals, expected)
}

func (s *FilterSuite) TestFilterDecryptKVPairs(t *C) {
	in := pongo2.AsValue(memkv.KVPairs{
		memkv.KVPair{
			Key:   "/some/key",
			Value: secret,
		},
	})
	expected := memkv.KVPairs{memkv.KVPair{Key: "/some/key", Value: "secret1\n"}}
	res, err := filterDecrypt(in, pongo2.AsValue("../integration/config/test.gpg"))
	if err != nil {
		t.Error(err.ErrorMsg)
		t.FailNow()
	}
	m2 := res.Interface().(memkv.KVPairs)
	t.Check(m2, DeepEquals, expected)
}

func (s *FilterSuite) TestFilterDecryptKVPair(t *C) {
	in := pongo2.AsValue(memkv.KVPair{
		Key:   "/some/key",
		Value: secret,
	})
	expected := memkv.KVPair{Key: "/some/key", Value: "secret1\n"}
	res, err := filterDecrypt(in, pongo2.AsValue("../integration/config/test.gpg"))
	if err != nil {
		t.Error(err.ErrorMsg)
		t.FailNow()
	}
	m2 := res.Interface().(memkv.KVPair)
	t.Check(m2, DeepEquals, expected)
}

func (s *FilterSuite) TestFilterDecryptStringArray(t *C) {
	in := pongo2.AsValue([]string{secret})
	expected := []string{"secret1\n"}

	res, err := filterDecrypt(in, pongo2.AsValue("../integration/config/test.gpg"))
	if err != nil {
		t.Error(err.ErrorMsg)
		t.FailNow()
	}
	m2 := res.Interface().([]string)
	t.Check(m2, DeepEquals, expected)
}
