//go:build !unix

package generator

import "os/exec"

func confineCommand(cmd *exec.Cmd) {}
