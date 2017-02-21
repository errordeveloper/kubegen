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
		{"-s", "-D", ".examples/modules/sockshop", "-v", "image_registry=gcr.io/sockshop"},
		{"-s", "-D", ".examples/modules/sockshop", "-v", "image_registry=quay.io/sockshop"},
		{"-s", "-D", ".examples/modules/weavecloud", "-v", "service_token=abc123"},
	}

	for _, command := range commands {
		hash := sha1.New()
		hash.Write([]byte(strings.Join(command, " ")))
		Commands[fmt.Sprintf(".%x", hash.Sum(nil))] = command
	}
}
