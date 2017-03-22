package main

import (
	"bytes"
	"io/ioutil"
	"os/exec"
	"strings"
	"testing"
)

func TestStartupScript(t *testing.T) {

	b, err := ioutil.ReadFile("./startup_script.centos.temp")
	if err != nil {
		t.Error(err)
		t.Skip()
	}

	script := strings.Replace(string(b), "{{.Callback}}", "127.0.0.1", -1)
	bash := strings.Replace(script, "{{.StartupScript}}", `#!/usr/bin/env bash
echo -n 'xxx_once '
if xxx_once /tmp/xx.lock; then echo ok; else echo fail; fi
test -f /tmp/xx.lock && echo locked || echo lock fail
echo -n 'xxx_once_check '
if xxx_once_check /tmp/xx.lock; then echo fail; else echo ok; fi
rm -f /tmp/xx.lock
if xxx_once_check /tmp/xx.lock; then echo ok; else echo fail; fi

	`, 1)

	t.Log(bash)
	buf := bytes.NewBuffer(nil)

	cmd := exec.Command("bash", "-c", bash)
	cmd.Stdout = buf
	cmd.Stderr = buf
	buf.WriteString("bash ----------\n")
	if err := cmd.Run(); err != nil {
		t.Error(err)
	}

	t.Log(buf.String())
}
