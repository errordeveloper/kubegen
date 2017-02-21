package commands

import (
	"crypto/sha1"
	"fmt"
	"strings"
)

var Commands map[string][]string

func init() {
	Commands = make(map[string][]string)
	commands := [][]string{
		{"--stdout=true", "--source-dir=.examples/modules/sockshop"},
	}

	for _, command := range commands {
		hash := sha1.New()
		hash.Write([]byte(strings.Join(command, " ")))
		Commands[fmt.Sprintf(".%x", hash.Sum(nil))] = command
	}
}
