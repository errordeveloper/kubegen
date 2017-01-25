package appmaker

import (
	"encoding/json"
	"fmt"
	"testing"
)

func SockShopTest(t *testing.T) {
	altPromPath := AppComponentOpts{PrometheusPath: "/prometheus"}

	zipkinEnv := map[string]string{
		"ZIPKIN": "http://zipkin:9411/api/v1/spans",
	}

	mongo := AppComponent{
		Image: "mongo",
		Port:  27017,
		Opts: AppComponentOpts{
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

	app := App{
		Name: "sock-shop",
		Group: []AppComponent{
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
				Opts: AppComponentOpts{
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
				Opts: AppComponentOpts{
					PrometheusPath:                 "/prometheus",
					WithoutStandardSecurityContext: true,
				},
			},
			{
				Image: "rabbitmq:3",
				Port:  5672,
				Opts: AppComponentOpts{
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
				Opts: AppComponentOpts{
					PrometheusScrape:               false,
					WithoutStandardProbes:          true,
					WithoutStandardSecurityContext: true,
				},
				//service_type: NodePort
			},
		},
	}

	for _, resources := range app.MakeAll() {
		deployment, err := json.MarshalIndent(resources.Deployment, "", "  ")
		if err != nil {
			panic(err)
		}

		fmt.Println(string(deployment))

		service, err := json.MarshalIndent(resources.Service, "", "  ")
		if err != nil {
			panic(err)
		}

		fmt.Println(string(service))
	}
}
