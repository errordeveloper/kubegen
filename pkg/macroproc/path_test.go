package macroproc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPathExp(t *testing.T) {
	assert := assert.New(t)

	testPathExpressions := []string{
		`["something"][0]`,
		`["something"]`,
		`[0]`,
		`[0]["something"]`,
		`["foo"][0]["something"][109][1]["bar"]["do it"]`,
		`["foo"][0]["something"][109][1]["bar"]["do ]it"]`,
		`""`,
	}

	for _, e := range testPathExpressions {
		p, err := getPathExp(e)
		if assert.Nil(err) {
			t.Logf("matched for %s => %v", e, p)
		} else {
			t.Logf("matched for %s", e)
		}
	}
}
