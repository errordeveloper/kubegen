package converter

import (
	"testing"

	_ "github.com/stretchr/testify/assert"
)

func TestCoverter(t *testing.T) {
	//tobj := map[string]map[string][]map[string]bool{
	tobj := []byte(`{
		"Kind": "Some",		
		"this":  true,
		"that":  false,
		"things": [
			{ "a": 1, "b": 2, "c": 3 }
		],
		"nothing": { "empty1": [], "empty2": [] },
		"other": {
			"moreThings": [
				{ "a": 1, "b": 2, "c": 3 },
				{ "a": 1, "b": 2, "c": 3 },
				{ "a": 1, "b": 2, "c": 3 }
			],
			"number": 1.0,
			"string": "foobar"
		}
	}`)
	conv := New()

	if err := conv.LoadStrict(tobj); err != nil {
		t.Fatalf("failed to laod – %v", err)
	}

	if err := conv.Run(); err != nil {
		t.Fatalf("failed to covert – %v", err)
	}

	for _, v := range conv.Dumps() {
		t.Log(v)
	}
}
