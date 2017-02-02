package sockshop

import (
	appmaker "github.com/errordeveloper/kubegen/pkg/apps"
)

func MakeSockShop() *appmaker.App {
	altPromPath := appmaker.AppComponentOpts{PrometheusPath: "/prometheus"}

	zipkinEnv := map[string]string{
		"ZIPKIN": "http://zipkin:9411/api/v1/spans",
	}

	return &appmaker.App{
		GroupName: "sock-shop",
		Templates: []appmaker.AppComponentTemplate{
			{
				TemplateName: "myStandardMongo",
				Image:        "mongo",
				AppComponent: appmaker.AppComponent{
					Port:   27017,
					Flavor: "minimal",
					//security:
					//  capabilities:
					//    drop: [ all ]
					//    add: [ CHOWN, SETGID, SETUID ]
					//  readOnlyRootFilesystem: true
					Env: zipkinEnv,
				},
			},
		},
		Components: []appmaker.AppComponent{
			{
				Name:                 "cart-db",
				BasedOnNamedTemplate: "myStandardMongo",
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
				Name:                 "orders-db",
				BasedOnNamedTemplate: "myStandardMongo",
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
				Image:                "weaveworksdemos/user-db:0.3.0",
				BasedOnNamedTemplate: "myStandardMongo",
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
