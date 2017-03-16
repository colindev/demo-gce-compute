package main

import (
	"net/url"
	"testing"
)

func TestConfig_Read(t *testing.T) {
	v := url.Values{}

	v.Set("a", "1")
	v.Set("b", "2")

	q := Config{
		"a": "",
		"b": "",
		"c": "x",
	}

	q.Read(v)

	if s := q["a"]; s != "1" {
		t.Errorf("a expect 1, but %s", s)
	}
	if s := q["b"]; s != "2" {
		t.Errorf("b expect 2, but %s", s)
	}
	if s := q["c"]; s != "x" {
		t.Errorf("c expect x, but %s", s)
	}
}
