package main

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/rendon/testcli"

	"github.com/errordeveloper/kubegen/cmd/kubegen/assets/commands"
)

func TestKubegen(t *testing.T) {
	for filename, command := range commands.Commands {
		c := testcli.GoRunMain(command...)
		c.Run()
		if !c.Success() {
			t.Fatalf("Command %v was expected to succeed, but failed with error: %s\n%s\n", c.Error(), c.StdoutAndStderr())
		}
		knownOutputFilePath := path.Join("assets", filename)
		knownOutput, err := ioutil.ReadFile(knownOutputFilePath)
		if err != nil {
			t.Fatalf("failed to read from %q for command %v – %v", knownOutputFilePath, command, err)
		}
		if c.Stdout() != string(knownOutput) {
			t.Fatalf("Expected %q to contain %q", c.Stdout(), string(knownOutput))
		}
	}
}
