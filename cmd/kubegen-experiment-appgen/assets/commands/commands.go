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
		{"--image", "errordeveloper/foo:latest", "--port=8080"},
		{"--image", "errordeveloper/foo:latest", "--flavor=minimal"},
		{"--image", "nginx:latest", "--flavor=minimal"},
		{"--image", "nginx", "--output=json-stdout"},
		{"--image", "errordeveloper/test:latest", "--flavor=default"},
		{"--image", "errordeveloper/test:latest", "--output=json-stdout"},
		{"--image", "errordeveloper/test:latest", "--flavor=minimal", "--port=4040", "--env", "FOO=bar,BAR=foo"},
		{"--image", "errordeveloper/test:latest", "--env=FOO=bar", "--env=BAR=foo", "--flavor=minimal"},
		{"--image", "foo/bar/test", "--replicas=100", "-o", "json-stdout", "-F=minimal"},
	}

	for _, command := range commands {
		hash := sha1.New()
		hash.Write([]byte(strings.Join(command, " ")))
		Commands[fmt.Sprintf(".%x", hash.Sum(nil))] = command
	}
}
