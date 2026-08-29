package main

import (
	"bytes"
	"cmp"
	"errors"
	"os/exec"
)

func runTmuxCmd(cmds ...string) (string, error) {
	cmd := exec.Command("tmux", cmds...)
	var out bytes.Buffer
	var serr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &serr
	err := cmd.Run()
	var prettyErr error
	if err != nil {
		prettyErr = errors.New(cmp.Or(serr.String(), err.Error()))
	}

	return out.String(), prettyErr
}
