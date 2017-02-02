package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/errordeveloper/kubegen/pkg/appmaker/assets/sockshop"
)

func main() {
	generated := []string{}
	app := sockshop.MakeSockShop()
	for _, resources := range app.MakeAll() {
		deployment, err := json.Marshal(resources.Deployment())
		if err != nil {
			panic(err)
		}

		generated = append(generated, string(deployment))

		service, err := json.Marshal(resources.Service())
		if err != nil {
			panic(err)
		}

		generated = append(generated, string(service))
	}

	manifest, err := json.Marshal(app)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n%s\n", strings.Join(generated, "\n"), string(manifest))
}
