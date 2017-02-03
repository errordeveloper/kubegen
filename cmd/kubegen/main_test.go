package main

import (
	"testing"

	"github.com/rendon/testcli"
)

func TestKubegen(t *testing.T) {
	c := testcli.GoRunMain("--image", "errordeveloper/foo:latest")
	c.Run()
	if !c.Success() {
		t.Fatalf("Expected to succeed, but failed with error: %s\n%s\n", c.Error(), c.StdoutAndStderr())
	}

	//if !c.StdoutContains("Hello John!") {
	//	t.Fatalf("Expected %q to contain %q", c.Stdout(), "Hello John!")
	//}
}
