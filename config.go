package main

import (
	"context"
	"net/url"
)

// Config read value from url.Values and allow set default
type Config map[string]string

func (c Config) Read(val url.Values) Config {
	readTo(c, val)

	return c
}

// ReadAndCopy read value from url.Value then copy self
func (c Config) ReadAndCopy(val url.Values) Config {
	o := Config{}
	for k, v := range c {
		o[k] = v
	}
	readTo(o, val)
	return o
}

// WithContext support context design
func (c Config) WithContext(ctx context.Context) context.Context {

	for k, v := range c {
		ctx = context.WithValue(ctx, k, v)
	}

	return ctx
}

func readTo(c Config, val url.Values) {
	for k := range c {
		if v := val.Get(k); v != "" {
			c[k] = v
		}
	}
}
