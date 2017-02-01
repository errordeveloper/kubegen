variable "zipkin" { }

deployment "catalogue" {
  metadata {
    labels {
      name = "catalogue"
    }
  }

  replicas = 1

  container "catalogue" {
    image = "weaveworksdemos/catalogue:0.3.0"
    env {
      ZIPKIN = "${var.zipkin}"
    }
    port {
      container_port = 80
    }
    security_context {
      run_as_non_root = true
      run_as_user = 10001
      read_only_root_filesystem = true
      capabilities {
        drop = [ "all" ]
        add = [ "NET_BIND_SERVICE" ]
      }
    }
    liveness_probe {
      http_get {
        path = "/health"
        port = 80
      }
      initial_delay_seconds = 300
      period_seconds = 3
    }
    readiness_probe {
      http_get {
        path = "/health"
        port = 80
      }
      initial_delay_seconds = 180
      period_seconds = 3
    }
  }
}

service "catalogue" {
  metadata {
    labels {
      name = "catalogue"
    }
  }

  selector {
    name = "catalogue"
  }

  port "catalogue" {
    target_port = 80
  }
}

deployment "catalogue-db" {
  metadata {
    labels {
      name = "catalogue-db"
    }
  }

  replicas = 1

  container "catalogue-db" {
    image = "weaveworksdemos/catalogue-db:0.3.0"

    # Was debating if this should be an array or k/v object.
    env {
      MYSQLROOT_PASSWORD = "${var.mysqlpassword}"
      MYSQL_DATABASE = "${var.mysqldb}"
    }

    # This port syntax, with the string name here seems a bit inconsistent with
    # the service `port` syntax (below) with a number. Maybe that is ok? I'm
    # not sure how it fits with other port declarations on containers and
    # services (or what they all are). Needs more investigation.
    port "mysql" {
      container_port = 3306
    }
  }
}

service "catalogue-db" {
  metadata {
    labels {
      name = "catalogue-db"
    }
  }

  selector {
    name = "catalogue-db"
  }

  port "catalogue-db" {
    target_port = 3306
  }
}
