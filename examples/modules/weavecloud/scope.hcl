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
