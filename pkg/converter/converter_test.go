package converter

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertPathKeys(pks [][]interface{}, conv *Converter, t *testing.T) {
	assert := assert.New(t)

	for _, pk := range pks {
		x := conv.get(pk...)
		if assert.NotNil(x, fmt.Sprintf("path %v should exist", pk)) {
			assert.Equal(len(pk)+1, len(x.path),
				fmt.Sprintf("lenght of path to %v should be len(pk)+1", pk))
			assert.Equal(fmt.Sprintf("%#v", pk), fmt.Sprintf("%#v", x.path[1:]),
				fmt.Sprintf("formatted path to %v should be the same", pk))
		}
	}
}

func TestConverterBasic(t *testing.T) {

	assert := assert.New(t)

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
		},
		"and more": {
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
		}
	}`)

	conv := New()

	if err := conv.loadStrict(tobj); err != nil {
		t.Fatalf("failed to laod – %v", err)
	}

	if err := conv.run(KeywordEvalPhaseB); err != nil {
		t.Fatalf("failed to covert – %v", err)
	}

	assert.NotNil(conv.locator.self)

	assert.Equal(7, len(conv.locator.self),
		"converter should have a tree of length 7")

	assert.NotNil(conv.locator.self["and more"])
	assert.NotNil(conv.locator.self["and more"].self)
	assert.Equal(6, len(conv.locator.self["and more"].self),
		"converter should have `and more` subtree of length 6")

	assert.Equal(Object, conv.get("other", "moreThings", 0).kind)
	assert.Equal(Object, conv.get("other", "moreThings", 1).kind)
	assert.Equal(Object, conv.get("other", "moreThings", 2).kind)
	assert.Nil(conv.get("other", "moreThings", 9))
}

func TestConverterOnlyObjects(t *testing.T) {

	assert := assert.New(t)

	tobj := []byte(`{
			"Kind": "Some",
			"this":  true,
			"that":  false,
			"things": { "a": 1, "b": 2, "c": 3 },
			"other": {
				"moreThings": {
					"1": { "a": 1, "b": 2, "c": 3 },
					"2": { "a": 1, "b": 2, "c": 3 },
					"3": { "a": 1, "b": 2, "c": 3 }
				},
				"number": 1.0,
				"string": "foobar"
			},
			"and more": {
				"Kind": "Some",
				"this":  true,
				"that":  false,
				"things": {
					"0": { "a": 1, "b": 2, "c": 3 }
				},
				"other": {
					"moreThings": {
						"1": { "a": 1, "b": 2, "c": 3 },
						"2": { "a": 1, "b": 2, "c": 3 },
						"3": { "a": 1, "b": 2, "c": 3 }
					},
					"number": 1.0,
					"string": "foobar"
				}
			}
		}`)

	conv := New()

	if err := conv.loadStrict(tobj); err != nil {
		t.Fatalf("failed to laod – %v", err)
	}

	if err := conv.run(KeywordEvalPhaseB); err != nil {
		t.Fatalf("failed to covert – %v", err)
	}

	assert.Equal(6, len(conv.locator.self),
		"converter should have a tree of length 7")
	assert.Equal(5, len(conv.locator.self["and more"].self),
		"converter should have `and more` subtree of length 6")

	assert.Equal(Object, conv.get("other", "moreThings", "1").kind)
	assert.Equal(Object, conv.get("other", "moreThings", "2").kind)
	assert.Equal(Object, conv.get("other", "moreThings", "3").kind)
	assert.Nil(conv.get("other", "moreThings", "0"))

	pathKeys := [][]interface{}{
		{"things"},
		{"things", "a"},
		{"things", "b"},
		{"things", "c"},
		{"other", "number"},
		{"other", "string"},
		{"other", "moreThings"},
		{"other", "moreThings", "1"},
		{"other", "moreThings", "1", "a"},
		{"other", "moreThings", "1", "b"},
		{"other", "moreThings", "1", "c"},
		{"other", "moreThings", "2"},
		{"other", "moreThings", "2", "a"},
		{"other", "moreThings", "2", "b"},
		{"other", "moreThings", "2", "c"},
		{"other", "moreThings", "3"},
		{"other", "moreThings", "3", "a"},
		{"other", "moreThings", "3", "b"},
		{"other", "moreThings", "3", "c"},
		{"and more", "Kind"},
		{"and more", "this"},
		{"and more", "that"},
		{"and more", "other", "number"},
		{"and more", "other", "string"},
	}

	assertPathKeys(pathKeys, conv, t)

}

func TestBasicKubegenAsset(t *testing.T) {

	assert := assert.New(t)

	tobj := []byte(`{
		"Kind": "kubegen.k8s.io/Module.v1alpha2",
		"Deployments": [
			{
				"name": "cart",
				"replicas": 1,
				"containers": [
					{
						"name": "cart",
						"image": "<image_registry>/cart:0.4.0",
						"ports": [
							{
								"name": "http",
								"containerPort": 80
							}
						],
						"securityContext": {
							"runAsNonRoot": true,
							"runAsUser": 10001,
							"capabilities": {
								"drop": [
									"all"
								],
								"add": [
									"NET_BIND_SERVICE"
								]
							},
							"readOnlyRootFilesystem": true
						},
						"volumeMounts": [
							{
								"mountPath": "/tmp",
								"name": "tmp-volume"
							}
						],
						"livenessProbe": {
							"httpGet": {
								"path": "/health"
							},
							"initialDelaySeconds": 300,
							"periodSeconds": 3
						},
						"readinessProbe": {
							"httpGet": {
								"path": "/health"
							},
							"initialDelaySeconds": 180,
							"periodSeconds": 3
						}
					}
				],
				"volumes": [
					{
						"name": "tmp-volume",
						"emptyDir": {
							"medium": "Memory"
						}
					}
				]
			},
			{
				"name": "cart-db",
				"kubegen.MapMerge.Lookup": "mongo",
				"replicas": 2
			}
		],
		"Services": [
			{
				"name": "cart",
				"annotations": {
					"prometheus.io/path": "/prometheus"
				},
				"ports": [
					{
						"name": "http"
					}
				]
			},
			{
				"name": "cart-db",
				"ports": [
					{
						"name": "mongo"
					}
				]
			}
		]
	}`)

	conv := New()

	if err := conv.loadStrict(tobj); err != nil {
		t.Fatalf("failed to laod – %v", err)
	}

	if err := conv.run(KeywordEvalPhaseB); err != nil {
		t.Fatalf("failed to covert – %v", err)
	}

	assert.Equal(2, len(conv.locator.self["Deployments"].self),
		"there are two Deployments")
	assert.Equal(String, conv.locator.self["Deployments"].self[0].self["name"].kind,
		"there should be name in a Deployments")
	assert.Equal(String, conv.locator.self["Deployments"].self[1].self["name"].kind,
		"there should be name in a Deployments")

	assert.Equal(2, len(conv.locator.self["Services"].self),
		"there are two Services")
	assert.Equal(String, conv.locator.self["Services"].self[0].self["name"].kind,
		"there should be cart in Services")
	assert.Equal(String, conv.locator.self["Services"].self[1].self["name"].kind,
		"there should be cart in Services")

	pathKeys := [][]interface{}{
		{"Deployments"},

		{"Deployments", 0},
		{"Deployments", 0, "name"},
		{"Deployments", 0, "replicas"},

		{"Deployments", 0, "containers"},
		{"Deployments", 0, "containers", 0},
		{"Deployments", 0, "volumes"},
		{"Deployments", 0, "volumes", 0},

		{"Deployments", 1},
		{"Deployments", 1, "name"},
		{"Deployments", 1, "replicas"},

		{"Services"},

		{"Services", 0, "ports"},
		{"Services", 0, "ports", 0, "name"},
		{"Services", 0, "annotations"},
		{"Services", 0, "annotations", "prometheus.io/path"},

		{"Services", 1, "ports"},
		{"Services", 1, "ports", 0, "name"},
	}

	assertPathKeys(pathKeys, conv, t)
}

func TestConverterGet(t *testing.T) {
	conv := New()
	tobj := []byte(`{ "Kind": "some" }`)

	tobj2 := new(map[string]interface{})
	tobj3 := new(map[string]interface{})

	if err := conv.loadStrict(tobj); err != nil {
		t.Fatalf("failed to laod – %v", err)
	}

	json.Unmarshal(tobj, tobj2)

	json.Unmarshal([]byte(`
		{ "mash": { "count": 1 }, "chips": { "count": 2 }, "sausages": true, "gravy": { "beef": 1, "chicken": 0 }  }
	`), tobj3)

	if err := conv.run(KeywordEvalPhaseB); err != nil {
		t.Fatalf("failed to covert – %v", err)
	}

	errors := make(chan error)
	go func() {
		for _, x := range []string{"1", "2", "3"} {
			conv.doIterate(&conv.locator, x, *tobj2, Object, errors)
		}

		conv.doIterate(&conv.locator, "order", make(map[string]interface{}), Object, errors)

		conv.doIterate(conv.get("order"), "potatoe",
			*tobj3,
			Object, errors)

		errors <- nil
	}()

	if err := <-errors; err != nil {
		t.Fatalf("%v", err)
	}

	pathKeys := [][]interface{}{
		{"1", "Kind"},
		{"2", "Kind"},
		{"3", "Kind"},
		{"order", "potatoe"},
		{"order", "potatoe", "mash"},
		{"order", "potatoe", "chips"},
	}

	assertPathKeys(pathKeys, conv, t)
}
