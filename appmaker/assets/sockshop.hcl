group_name = "sockshop"

//altPromPath := appmaker.AppComponentOpts{PrometheusPath: "/prometheus"}

//zipkinEnv := map[string]string{
//	"ZIPKIN": "http://zipkin:9411/api/v1/spans",
//}

component_template "mongo" {
  image  = "mongo"
  port   = 27017
  flavor = "minimal"
  //Env: zipkinEnv,
}

component_from_template "myStandardMongo" {
  name = "cart-db"
}

component_from_image "weaveworksdemos/cart:0.4.0" {
  //Opts: altPromPath,
}

component_from_image "weaveworksdemos/catalogue-db:0.3.0" {
  port = 3306
  //Opts: appmaker.AppComponentOpts{
  //PrometheusScrape:               false,
  //WithoutStandardProbes:          true,
  //WithoutStandardSecurityContext: true,
  //},
  env = {
    MYSQL_ROOT_PASSWORD = "fake_password"
    MYSQL_DATABASE      = "socksdb"
  }
}

component_from_image "weaveworksdemos/catalogue:0.3.0" {
  //Env:   zipkinEnv,
}

component_from_image "weaveworksdemos/front-end:0.3.0" {
  port = 8079
  //service_port: 80
  //service_type: NodePort
  //service_session_affinity: ClientIP
}

component_from_template "mongo" {
  name = "orders-db"
}

component_from_image "weaveworksdemos/orders:0.4.2" {
  //Opts:  altPromPath,
}

component_from_image "weaveworksdemos/payment:0.4.0" {
  //Env:   zipkinEnv,
}

component_from_image "weaveworksdemos/queue-master:0.3.0" {
  //Opts: appmaker.AppComponentOpts{
  //PrometheusPath:                 "/prometheus",
  //WithoutStandardSecurityContext: true,
  //}
}

component_from_image "rabbitmq:3" {
  port = 5672
  //Opts: appmaker.AppComponentOpts{
  //PrometheusScrape:      false,
  //WithoutStandardProbes: true,
  //},
  //security:
  //  capabilities:
  //    drop: [ all ]
  //    add: [ CHOWN, SETGID, SETUID, DAC_OVERRIDE ]
  //  readOnlyRootFilesystem: true
}

component_from_image "weaveworksdemos/shipping:0.4.0" {
  //Opts:  altPromPath,
}

component_from_template "mongo" {
  image = "weaveworksdemos/user-db:0.3.0"
}

component_from_image "weaveworksdemos/user:0.4.0" {
  env = {
    ZIPKIN = "http://zipkin:9411/api/v1/spans"
    MONGO_HOST = "user-db:27017"
  }
}

component_from_image "openzipkin/zipkin" {
  port = 9411
  //Opts: appmaker.AppComponentOpts{
  //PrometheusScrape:               false,
  //WithoutStandardProbes:          true,
  //WithoutStandardSecurityContext: true,
  //},
  //service_type: NodePort
}
