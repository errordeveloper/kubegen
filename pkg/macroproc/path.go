package macroproc

import (
	"fmt"
	"regexp"
)

const pathExtractArrayIndices = `\[(?P<arrayIndex>\d+)\]`
const pathExtractObjectIndices = `\["(?P<objectIndex>[a-z|A-Z]+)"\]`

var pathexp *regexp.Regexp

func init() {
	pathexp = regexp.MustCompile(fmt.Sprintf("%s|%s",
		pathExtractArrayIndices,
		pathExtractObjectIndices,
	))
}

func getPathExp(s string) ([][]string, error) {
	if pathexp.MatchString(s) {
		//result := make([]map[string]string, 0)

		// for i, m := range pathexp.FindAllStringSubmatch(s, -1) {
		// 	result
		// }
		return pathexp.FindAllStringSubmatch(s, -1), nil
	}
	return nil, fmt.Errorf("no match")
}
