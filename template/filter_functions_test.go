package template

import (
	"testing"

	"github.com/flosch/pongo2"
)

func TestFilterBase64(t *testing.T) {
	in := pongo2.AsValue("foo")
	res, err := filterBase64(in, nil)
	if err != nil {
		t.Error(err.ErrorMsg)
	}

	if res.String() != "Zm9v" {
		t.Errorf("Expected filterBase64 to be %v, got %v", "Zm9v", res.String())
	}
}

func TestFilterBase(t *testing.T) {
	in := pongo2.AsValue("/etc/foo/bar")
	res, err := filterBase(in, nil)
	if err != nil {
		t.Error(err.ErrorMsg)
	}

	if res.String() != "bar" {
		t.Errorf("Expected filterBase to be %v, got %v", "bar", res.String())
	}
}

func TestFilterDir(t *testing.T) {
	in := pongo2.AsValue("/etc/foo/bar")
	res, err := filterDir(in, nil)
	if err != nil {
		t.Error(err.ErrorMsg)
	}

	if res.String() != "/etc/foo" {
		t.Errorf("Expected filterDir to be %v, got %v", "/etc/foo", res.String())
	}
}

func TestFilterSplit(t *testing.T) {
	in := pongo2.AsValue("foo/bar")
	res, err := filterSplit(in, pongo2.AsValue("/"))
	if err != nil {
		t.Error(err.ErrorMsg)
	}

	if res.Interface().([]string)[0] != "foo" || res.Interface().([]string)[1] != "bar" {
		t.Errorf("Expected filterSplit to be %v, got %v", []string{"foo", "bar"}, res.Interface())
	}
}

func TestFilterToPrettyJson(t *testing.T) {
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

	if res.String() != expected {
		t.Errorf("Expected filterToPrettyJson to be %v, got %v", expected, res.String())
	}
}

func TestFilterToJson(t *testing.T) {
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
	if res.String() != expected {
		t.Errorf("Expected filterToJson to be %v, got %v", expected, res.String())
	}
}
