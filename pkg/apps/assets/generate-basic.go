package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/errordeveloper/kubegen/pkg/apps/assets/sockshop"
)

func main() {
	generated := []string{}
	app := sockshop.MakeSockShop()
	appManifest, err := app.MakeAll()
	if err != nil {
		panic(err)
	}
	for _, resources := range appManifest {
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
