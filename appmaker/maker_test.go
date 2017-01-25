package appmaker

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func TestMakeAll(t *testing.T) {
	generated := make([]interface{}, 0)
	temp := new(interface{})
	data := make([]byte, 0)

	for _, resources := range SockShop.MakeAll() {
		data, err := json.Marshal(resources.Deployment)
		if err != nil {
			t.Fatal(err)
		}

		if err := json.Unmarshal(data, temp); err != nil {
			t.Fatal(err)
		}

		generated = append(generated, temp)

		data, err = json.Marshal(resources.Service)
		if err != nil {
			t.Fatal(err)
		}

		if err := json.Unmarshal(data, temp); err != nil {
			t.Fatal(err)
		}

		generated = append(generated, temp)
	}

	data, err := ioutil.ReadFile("assets/sockshop.json")
	if err != nil {
		t.Fatal(err)
	}

	for i, input := range strings.Split(string(data), "\n") {
		if err := json.Unmarshal([]byte(input), temp); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(generated[i], temp) {
			t.FailNow()
		}
	}
}
