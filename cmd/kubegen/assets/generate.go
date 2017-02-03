package main

import (
	"fmt"
	"os"

	"github.com/rendon/testcli"

	"github.com/errordeveloper/kubegen/cmd/kubegen/assets/commands"
)

func main() {
	for _, command := range commands.Commands {
		c := testcli.GoRun("../main.go", command...)
		c.Run()
		if !c.Success() {
			fmt.Fprintf(os.Stderr, "Command %v expected to succeed, but failed with error: %s\n%s\n", command, c.Error(), c.StdoutAndStderr())
		}
	}
}
