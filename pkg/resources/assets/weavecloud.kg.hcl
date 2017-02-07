deployment "weave-flux-agent" {
  labels {
    app = "weave-flux"
    name = "weave-flux-agent"
    weave-cloud-component = "flux"
    weave-flux-component = "agent"
  }

  replicas = 1

  container "agent" {
    image = "quay.io/weaveworks/fluxd:0.1.0"
    image_pull_policy = "IfNotPresent"
    args = [
          "--token={{.Values.ServiceToken}}"
    ]
  }
}

service "weave-flux-agent" {
  labels {
    app = "weave-flux"
    name = "weave-flux-agent"
    weave-cloud-component = "flux"
    weave-flux-component = "agent"
  }
}

daemonset "weave-scope-agent" {
  labels {
    app = "weave-scope"
    name = "weave-scope-agent"
    weave-cloud-component = "scop"
    weave-scope-component = "agent"
  }

  host_pid = true
  host_network = true

  container "agent" {
    image = "weaveworks/scope:latest"
    image_pull_policy = "IfNotPresent"
    args = [
      "--no-app",
      "--probe.docker.bridge=docker0",
      "--probe.docker=true",
      "--probe.kubernetes=true",
      "--service-token={{.Values.ServiceToken}}"
    ]
    security_context = {
      privileged = true
    }
    mounts "docker-socket" {
        mount_path = "/var/run/docker.sock"
    }
    mount "scope-plugins" {
        mount_path = "/var/run/scope/plugins"
    }
  }

  volume "docker-socket" {
    host_path {
      path = "/var/run/docker.sock"
    }
  }

  volume "scope-plugins" {
    host_path {
      path = "/var/run/scope/plugins"
    }
  }
}

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
      #mount "agent-config-volume" {
      #  mount_path = "/etc/prometheus"
      #}
  }

  #volume "agent-config-volume" {
  #  configmap {
  #    name = "weave-cortex-agent-config"
  #  }
  #}
}

configmap "weave-cortex-agent-config" {
  labels {
    app = "weave-cortex"
    name = "weave-cortex-agent-config"
    weave-cloud-component = "cortex"
    weave-cortex-component = "agent-config"
  }
  data {
    "prometheus.yml" = <<EOF
        global:
          scrape_interval: 15s
        remote_write:
          url: 'https://cloud.weave.works/api/prom/push'
          basic_auth:
            password: '{{.Values.ServiceToken}}'
        scrape_configs:
          - job_name: kubernetes-service-endpoints
            kubernetes_sd_configs:
              - role: endpoints
            tls_config:
              ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
            bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
            relabel_configs:
              - source_labels:
                  - __meta_kubernetes_service_label_component
                regex: apiserver
                action: replace
                target_label: __scheme__
                replacement: https
              - source_labels:
                  - __meta_kubernetes_service_label_kubernetes_io_cluster_service
                action: drop
                regex: 'true'
              - source_labels:
                  - __meta_kubernetes_service_annotation_prometheus_io_scrape
                action: drop
                regex: 'false'
              - source_labels:
                  - __meta_kubernetes_pod_container_port_name
                action: drop
                regex: .*-noscrape
              - source_labels:
                  - __meta_kubernetes_service_annotation_prometheus_io_scheme
                action: replace
                target_label: __scheme__
                regex: ^(https?)$
                replacement: $1
              - source_labels:
                  - __meta_kubernetes_service_annotation_prometheus_io_path
                action: replace
                target_label: __metrics_path__
                regex: ^(.+)$
                replacement: $1
              - source_labels:
                  - __address__
                  - __meta_kubernetes_service_annotation_prometheus_io_port
                action: replace
                target_label: __address__
                regex: '^(.+)(?::\d+);(\d+)$'
                replacement: '$1:$2'
              - action: labelmap
                regex: ^__meta_kubernetes_service_label_(.+)$
                replacement: $1
              - source_labels:
                  - __meta_kubernetes_namespace
                  - __meta_kubernetes_service_name
                separator: /
                target_label: job
          - job_name: kubernetes-pods
            kubernetes_sd_configs:
              - role: pod
            relabel_configs:
              - source_labels:
                  - __meta_kubernetes_pod_annotation_prometheus_io_scrape
                action: keep
                regex: 'true'
              - source_labels:
                  - __meta_kubernetes_namespace
                  - __meta_kubernetes_pod_label_name
                separator: /
                target_label: job
              - source_labels:
                  - __meta_kubernetes_pod_node_name
                target_label: node
          - job_name: kubernetes-nodes
            kubernetes_sd_configs:
              - role: node
            tls_config:
              insecure_skip_verify: true
            bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
            relabel_configs:
              - target_label: __scheme__
                replacement: https
              - source_labels:
                  - __meta_kubernetes_node_label_kubernetes_io_hostname
                target_label: instance
          - job_name: weave
            kubernetes_sd_configs:
              - role: pod
            relabel_configs:
              - source_labels:
                  - __meta_kubernetes_namespace
                  - __meta_kubernetes_pod_label_name
                action: keep
                regex: ^kube-system;weave-net$
              - source_labels:
                  - __meta_kubernetes_pod_container_name
                  - __address__
                action: replace
                target_label: __address__
                regex: '^weave;(.+?)(?::\d+)?$'
                replacement: '$1:6782'
              - source_labels:
                  - __meta_kubernetes_pod_container_name
                  - __address__
                action: replace
                target_label: __address__
                regex: '^weave-npc;(.+?)(?::\d+)?$'
                replacement: '$1:6781'
              - source_labels:
                  - __meta_kubernetes_pod_container_name
                action: replace
                target_label: job
EOF
  }

  data_to_json {
    foo {
      bar = 1
      baz {
        foo = [1, 2, 3]
      }
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
