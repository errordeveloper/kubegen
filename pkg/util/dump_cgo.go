// +build cgo

package util

import (
	"fmt"

	"github.com/d4l3k/go-highlight"
	"github.com/docker/docker/pkg/term"
)

func Dump(outputFormat string, data []byte) error {
	var (
		output string
	)

	if term.IsTerminal(0) {
		veryPretty, err := highlight.Term(outputFormat, data)
		if err != nil {
			return fmt.Errorf("kubegen/util: error colorizing the output for %q – %v", outputFormat, err)
		}
		output = string(veryPretty)
	} else {
		output = string(data)
	}

	fmt.Println(output)

	return nil
}
