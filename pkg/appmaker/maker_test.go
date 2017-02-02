package appmaker

import (
	_ "fmt"

	"encoding/json"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func parseManifest(t *testing.T) (*App, []string) {
	data, err := ioutil.ReadFile("assets/sockshop-basic.json")
	if err != nil {
		t.Fatal(err)
	}
	allManifests := strings.Split(strings.TrimSpace(string(data)), "\n")

	appManifestIndex := len(allManifests) - 1
	// Last line in our JSON fixture is an encoded version of App struct
	// this is done like this becuase we need to define the whole thing
	// somewhere, but we hit circular import errors when tests include
	// a package that also includes our package...
	appManifest := []byte(allManifests[appManifestIndex])
	appResources := allManifests[:appManifestIndex]

	app, err := NewFromJSON(appManifest)
	if err != nil {
		t.Fatal(err)
	}

	return app, appResources
}

func TestMakeAll(t *testing.T) {
	generated := make([]interface{}, 0)
	temp := new(interface{})

	app, serialized := parseManifest(t)

	for _, resources := range app.MakeAll() {
		data, err := json.Marshal(resources.Deployment())
		if err != nil {
			t.Fatal(err)
		}

		if err := json.Unmarshal(data, temp); err != nil {
			t.Fatal(err)
		}

		generated = append(generated, temp)

		if data, err = json.Marshal(resources.Service()); err != nil {
			t.Fatal(err)
		}

		if err := json.Unmarshal(data, temp); err != nil {
			t.Fatal(err)
		}

		generated = append(generated, temp)
	}

	for i, acceptable := range serialized {
		if err := json.Unmarshal([]byte(acceptable), temp); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(generated[i], temp) {
			t.FailNow()
		}
	}
}

func TestMarshalMakeAllEachToJSON(t *testing.T) {
	generated := make([]interface{}, 0)
	temp := new(interface{})

	app, serialized := parseManifest(t)

	for _, resources := range app.MakeAll() {
		dataMap, err := resources.MarshalToJSON()
		if err != nil {
			t.Fatal(err)
		}

		for _, data := range dataMap {
			if err := json.Unmarshal(data, temp); err != nil {
				t.Fatal(err)
			}

			generated = append(generated, temp)
		}
	}

	for i, acceptable := range serialized {
		if err := json.Unmarshal([]byte(acceptable), temp); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(generated[i], temp) {
			t.FailNow()
		}
	}
}

func TestMarshalAllToJSON(t *testing.T) {
	generated := make([]interface{}, 0)
	temp := new(interface{})

	app, serialized := parseManifest(t)

	all, err := app.MarshalToJSON()
	if err != nil {
		t.Fatal(err)
	}

	for _, dataMap := range all {

		for _, data := range dataMap {
			if err := json.Unmarshal(data, temp); err != nil {
				t.Fatal(err)
			}

			generated = append(generated, temp)
		}
	}

	for i, acceptable := range serialized {
		if err := json.Unmarshal([]byte(acceptable), temp); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(generated[i], temp) {
			t.FailNow()
		}
	}
}
