package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoverterBasic(t *testing.T) {

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

	if err := conv.LoadStrict(tobj); err != nil {
		t.Fatalf("failed to laod – %v", err)
	}

	if err := conv.Run(); err != nil {
		t.Fatalf("failed to covert – %v", err)
	}

	assert.Equal(7, len(conv.tree.self), "converter should have a tree of length 7")
	assert.Equal(6, len(conv.tree.self["and more"].self), "converter should have `and more` subtree of length 6")
	for _, v := range conv.Dump() {
		t.Log(v)
	}

}

func TestBasicKubegenAsset(t *testing.T) {
	tobj := []byte(`{
		"Kind": "kubegen.k8s.io/Module.v1alpha1",
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
				"kubegen.fromPartial": "mongo",
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

	if err := conv.LoadStrict(tobj); err != nil {
		t.Fatalf("failed to laod – %v", err)
	}

	if err := conv.Run(); err != nil {
		t.Fatalf("failed to covert – %v", err)
	}

	for _, v := range conv.Dump() {
		t.Log(v)
	}
}
