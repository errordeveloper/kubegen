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
	}

	for _, command := range commands {
		hash := sha1.New()
		hash.Write([]byte(strings.Join(command, " ")))
		Commands[fmt.Sprintf("%x", hash.Sum(nil))] = command
	}
}
