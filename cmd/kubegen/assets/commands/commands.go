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
		{"module", "-s", ".examples/modules/sockshop", "-v", "image_registry=gcr.io/sockshop"},
		{"module", "-s", ".examples/modules/sockshop", "-v", "image_registry=quay.io/sockshop"},
		{"module", "-s", ".examples/modules/weavecloud", "-v", "service_token=abc123"},
		{"bundle", "--stdout", ".examples/sockshop-prod.yml"},
		{"bundle", "--stdout", ".examples/weavecloud.yml"},
		{"bundle", "--stdout", ".examples/weavecloud.yml", ".examples/sockshop-prod.yml"},
		{"module", "--output=json", "--stdout=true", ".examples/modules/sockshop"},
		{"module", "--output=json", "-s", ".examples/modules/sockshop", "-v", "image_registry=gcr.io/sockshop"},
		{"module", "--output=json", "-s", ".examples/modules/sockshop", "-v", "image_registry=quay.io/sockshop"},
		{"module", "--output=json", "-s", ".examples/modules/weavecloud", "-v", "service_token=abc123"},
		{"bundle", "--output=json", "--stdout", ".examples/sockshop-prod.yml"},
		{"bundle", "--output=json", "--stdout", ".examples/weavecloud.yml"},
		{"bundle", "--output=json", "--stdout", ".examples/weavecloud.yml", ".examples/sockshop-prod.yml"},
	}

	for _, command := range commands {
		hash := sha1.New()
		hash.Write([]byte(strings.Join(command, " ")))
		Commands[fmt.Sprintf(".%x", hash.Sum(nil))] = command
	}
}
