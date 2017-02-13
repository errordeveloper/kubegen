namespace = "kube-system"

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
