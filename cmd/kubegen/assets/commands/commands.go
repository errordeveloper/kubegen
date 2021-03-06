package commands

import (
	"crypto/sha1"
	"fmt"
	"strings"
)

var Commands map[string][]string

// TODO `--output=json`
// TODO `--stdout=false`
// TODO `bundle -m`
func init() {
	Commands = make(map[string][]string)
	commands := [][]string{
		{"module", "--stdout=true", ".examples/modules/sockshop"},
		{"module", "-s", ".examples/modules/sockshop", "-p", "image_registry=gcr.io/sockshop"},
		{"module", "-s", ".examples/modules/sockshop", "-p", "image_registry=quay.io/sockshop"},
		{"module", "-s", ".examples/modules/weavecloud", "-p", "service_token=abc123"},
		{"bundle", "--stdout", ".examples/sockshop.yml"},
		{"bundle", "--stdout", ".examples/weavecloud.yml"},
		{"bundle", "--stdout", ".examples/weavecloud.yml", ".examples/sockshop.yml"},
		{"module", "--output=json", "--stdout=true", ".examples/modules/sockshop"},
		{"module", "--output=json", "-s", ".examples/modules/sockshop", "-p", "image_registry=gcr.io/sockshop"},
		{"module", "--output=json", "-s", ".examples/modules/sockshop", "-p", "image_registry=quay.io/sockshop"},
		{"module", "--output=json", "-s", ".examples/modules/weavecloud", "-p", "service_token=abc123"},
		{"bundle", "--output=json", "--stdout", ".examples/sockshop.yml"},
		{"bundle", "--output=json", "--stdout", ".examples/weavecloud.yml"},
		{"bundle", "--output=json", "--stdout", ".examples/weavecloud.yml", ".examples/sockshop.yml"},
	}

	for _, command := range commands {
		hash := sha1.New()
		hash.Write([]byte(strings.Join(command, " ")))
		Commands[fmt.Sprintf(".generated/%x", hash.Sum(nil))] = command
	}
}
