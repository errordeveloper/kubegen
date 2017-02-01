deployment "cart" {
  metadata {
    labels {
      name = "cart"
    }
  }

  # Probable the default, but just added here for clarity
  replicas = 1

  # Define volumes.
  volume "tmp-volume" {
    emptyDir {
      medium = "Memory"
    }
  }

  # Add container specs, this could be done multiple times.
  # It is unclear from the k8s manifest yaml whether you need the template
  # metadata, or what...
  container "cart" {
    image = "weaveworksdemos/cart:0.4.0"

    # expose a port from this container
    port {
      containerPort = 80
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

    # Mount the volume into this container
    mount "tmp-volume" {
      mount_path = "/tmp"
    }

    # We use these in a few places, would be nice to be able to parameterize
    # them, and define them in one place.
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

service "cart" {
  metadata {
    labels {
      name = "cart"
    }
    annotations {
      "prometheus.io/path" = "/prometheus"
    }
  }

  # Maybe selector for the same name should be the default?
  selector {
    name = "cart"
  }

  port 80 {
    targetPort = 80
  }
}

deployment "cart-db" {
  metadata {
    labels {
      name = "cart-db"
    }
  }

  replicas = 1

  volume "tmp-volume" {
    emptyDir {
      medium = "Memory"
    }
  }

  container "cart-db" {
    image = "mongo"
    port {
      containerPort = 27017
    }
    security_context {
      read_only_root_filesystem = true
      capabilities {
        drop = [ "all" ]
        add = [ "CHOWN", "SETGID", "SETUID" ]
      }
    }
    mount "tmp-volume" {
      mount_path = "/tmp"
    }
  }
}

service "cart-db" {
  metadata {
    labels {
      name = "cart-db"
    }
    annotations {
      # is this valid hcl?
      "prometheus.io/path" = "/prometheus"
    }
  }

  selector {
    name = "cart-db"
  }

  port 27017 {
    targetPort = 27017
  }
}
