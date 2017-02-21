namespace = "kube-system"

deployment "weave-cortex-agent" {
  labels = {
    app = "weave-cortex"
    name = "weave-cortex-agent"
    weave-cloud-component = "cortex"
    weave-cortex-component = "agent"
  }

  replicas = 1

  container "agent" {
      image = "prom/prometheus:v1.3.1"
      image_pull_policy = "IfNotPresent"
      args = [
        "-config.file=/etc/prometheus/prometheus.yml",
        "-web.listen-address=:80"
      ]
      port "agent" {
        container_port = 80
        protocol = "TCP"
      }
      mount "agent-config-volume" {
        mount_path = "/etc/prometheus"
      }
  }

  volume "agent-config-volume" {
    configmap {
      name = "weave-cortex-agent-config"
    }
  }
}


service "weave-cortex-agent" {
    labels {
      app = "weave-cortex"
      name = "weave-cortex-agent"
      weave-cloud-component = "cortex"
      weave-cortex-component = "agent"
    }

    port "agent" { }
}

daemonset "weave-cortex-node-exporter" {
  labels {
    app = "weave-cortex"
    name = "weave-cortex-node-exporter"
    weave-cloud-component = "cortex"
    weave-cortex-component = "node-exporter"
  }

  pod_annotations {
    "prometheus.io.scrape" = "true"
  }

  container "agent" {
    image = "prom/node-exporter:0.12.0"
    image_pull_policy = "IfNotPresent"
    security_context {
      privileged = true
    }
    port "agent" {
      container_port = 9100
      protocol = "TCP"
    }
  }
  host_pid = true
  host_network = true
}
