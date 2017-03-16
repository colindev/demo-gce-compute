package main

import "net/url"

type Config map[string]string

func (c Config) Read(val url.Values) {
	readTo(c, val)
}

func (c Config) ReadAndCopy(val url.Values) Config {
	o := Config{}
	for k, v := range c {
		o[k] = v
	}
	readTo(o, val)
	return o
}

func readTo(c Config, val url.Values) {
	for k := range c {
		if v := val.Get(k); v != "" {
			c[k] = v
		}
	}
}
