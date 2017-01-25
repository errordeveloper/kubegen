package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/errordeveloper/kubegen/appmaker"
)

func main() {
	generated := []string{}

	for _, resources := range appmaker.SockShop.MakeAll() {
		deployment, err := json.Marshal(resources.Deployment)
		if err != nil {
			panic(err)
		}

		generated = append(generated, string(deployment))

		service, err := json.Marshal(resources.Service)
		if err != nil {
			panic(err)
		}

		generated = append(generated, string(service))
	}

	fmt.Printf(strings.Join(generated, "\n"))
}
