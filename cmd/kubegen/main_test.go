package main

import (
	"fmt"

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
		t.Run(fmt.Sprintf("args=[%v]", command), func(t *testing.T) {
			t.Parallel()
			c := testcli.GoRunMain(append([]string{"bundle.go", "module.go", "self_upgrade.go"}, command...)...)
			c.Run()
			if !c.Success() {
				t.Fatalf("Command %v was expected to succeed, but failed with error: %s\n%s\n", command, c.Error(), c.StdoutAndStderr())
			}

			knownOutputFilePath := path.Join("assets", filename)
			knownOutput, err := ioutil.ReadFile(knownOutputFilePath)
			if err != nil {
				t.Fatalf("failed to read from %q for command %v – %v", knownOutputFilePath, command, err)
			}

			testOutput := c.Stdout()
			assert.True((len(testOutput) > 0))
			assert.Equal(string(knownOutput), testOutput)
		})
	}
}
