package main

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/errordeveloper/testcli"
	"github.com/stretchr/testify/assert"

	"github.com/errordeveloper/kubegen/cmd/kubegen/assets/commands"
)

func TestKubegenCmd(t *testing.T) {
	assert := assert.New(t)
	for filename, command := range commands.Commands {
		c := testcli.GoRunMain(append([]string{"bundle.go", "module.go", "update.go"}, command...)...)
		c.Run()
		if !c.Success() {
			t.Fatalf("Command %v was expected to succeed, but failed with error: %s\n%s\n", c.Error(), c.StdoutAndStderr())
		}

		knownOutputFilePath := path.Join("assets", filename)
		knownOutput, err := ioutil.ReadFile(knownOutputFilePath)
		if err != nil {
			t.Fatalf("failed to read from %q for command %v – %v", knownOutputFilePath, command, err)
		}

		assert.Equal(c.Stdout(), string(knownOutput))
	}
}
