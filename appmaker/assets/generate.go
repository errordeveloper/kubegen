package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/errordeveloper/kubegen/appmaker"
)

func main() {
	generated := []string{}
	app := makeSockShop()
	for _, resources := range app.MakeAll() {
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

	manifest, err := json.Marshal(app)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n%s\n", strings.Join(generated, "\n"), string(manifest))
}

func makeSockShop() *appmaker.App {
	altPromPath := appmaker.AppComponentOpts{PrometheusPath: "/prometheus"}

	zipkinEnv := map[string]string{
		"IPKIN": "http://zipkin:9411/api/v1/spans",
	}

	mongo := appmaker.AppComponent{
		Image: "mongo",
		Port:  27017,
		Opts: appmaker.AppComponentOpts{
			PrometheusScrape:      false,
			WithoutStandardProbes: true,
		},
		//security:
		//  capabilities:
		//    drop: [ all ]
		//    add: [ CHOWN, SETGID, SETUID ]
		//  readOnlyRootFilesystem: true
		Env: zipkinEnv,
	}

	return &appmaker.App{
		Name: "sock-shop",
		Group: []appmaker.AppComponent{
			{
				Name:    "cart-db",
				BasedOn: &mongo,
			},
			{
				Image: "weaveworksdemos/cart:0.4.0",
				Opts:  altPromPath,
			},
			{
				Image: "weaveworksdemos/catalogue-db:0.3.0",
				Port:  3306,
				Opts: appmaker.AppComponentOpts{
					PrometheusScrape:               false,
					WithoutStandardProbes:          true,
					WithoutStandardSecurityContext: true,
				},
				Env: map[string]string{
					"MYSQL_ROOT_PASSWORD": "fake_password",
					"MYSQL_DATABASE":      "socksdb",
				},
			},
			{
				Image: "weaveworksdemos/catalogue:0.3.0",
				Env:   zipkinEnv,
			},
			{
				Image: "weaveworksdemos/front-end:0.3.0",
				Port:  8079,
				//service_port: 80
				//service_type: NodePort
				//service_session_affinity: ClientIP
			},
			{
				Name:    "orders-db",
				BasedOn: &mongo,
			},
			{
				Image: "weaveworksdemos/orders:0.4.2",
				Opts:  altPromPath,
			},
			{
				Image: "weaveworksdemos/payment:0.4.0",
				Env:   zipkinEnv,
			},
			{
				Image: "weaveworksdemos/queue-master:0.3.0",
				Opts: appmaker.AppComponentOpts{
					PrometheusPath:                 "/prometheus",
					WithoutStandardSecurityContext: true,
				},
			},
			{
				Image: "rabbitmq:3",
				Port:  5672,
				Opts: appmaker.AppComponentOpts{
					PrometheusScrape:      false,
					WithoutStandardProbes: true,
				},
				//security:
				//  capabilities:
				//    drop: [ all ]
				//    add: [ CHOWN, SETGID, SETUID, DAC_OVERRIDE ]
				//  readOnlyRootFilesystem: true
			},
			{
				Image: "weaveworksdemos/shipping:0.4.0",
				Opts:  altPromPath,
			},
			{
				Image:   "weaveworksdemos/user-db:0.3.0",
				BasedOn: &mongo,
			},
			{
				Image: "weaveworksdemos/user:0.4.0",
				Env: map[string]string{
					"ZIPKIN":     "http://zipkin:9411/api/v1/spans",
					"MONGO_HOST": "user-db:27017",
				},
			},
			{
				Image: "openzipkin/zipkin",
				Port:  9411,
				Opts: appmaker.AppComponentOpts{
					PrometheusScrape:               false,
					WithoutStandardProbes:          true,
					WithoutStandardSecurityContext: true,
				},
				//service_type: NodePort
			},
		},
	}
}
