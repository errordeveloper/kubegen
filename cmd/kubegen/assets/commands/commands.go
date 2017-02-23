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
		{"module", "--stdout=true", ".examples/modules/sockshop"},
		{"module", "-s", ".examples/modules/sockshop", "-v", "image_registry=gcr.io/sockshop"},
		{"module", "-s", ".examples/modules/sockshop", "-v", "image_registry=quay.io/sockshop"},
		{"module", "-s", ".examples/modules/weavecloud", "-v", "service_token=abc123"},
	}

	for _, command := range commands {
		hash := sha1.New()
		hash.Write([]byte(strings.Join(command, " ")))
		Commands[fmt.Sprintf(".%x", hash.Sum(nil))] = command
	}
}
