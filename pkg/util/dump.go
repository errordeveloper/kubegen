// +build !cgo

package util

import "fmt"

func Dump(_ string, data []byte) error {
	fmt.Println(string(data))
	return nil
}
