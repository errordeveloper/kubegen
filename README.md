# `kubegen` – simple way to describe Kubernetes resources

Kubernetes resource definitions are too verbose, and there is no built-in framework for reusability.
Writing good resource templates is hard, whether you are rolling your own or using Helm.

***The aim of this project is to make it easier to write reusable Kubernetes resource definitions.***

> It should be useful as is, but it's ambition is to drive the community towards an improvement
upstream. However, please note that it is **WORK IN PROGRESS** rigth now.

Firtsly, `kubegen` provides a simple non-recursive system of modules, which allows you to define resource with a few simple parameters once and instantiate it multiple times with different values for those parameters.

For example, you can use it to describe two different environments where your app runs:
```YAML
Modules:
  - Name: "prodApp"
    SourceDir: "modules/myapp"
    OutputDir: "env/prod"
    Variables:
      api_service_replicas: 100
      domain_name: "errors.io"
  - Name: "testApp"
    SourceDir: "modules/myapp"
    OutputDir: "env/test"
    Variables:
      api_service_replicas: 10
      use_rds: false
      domain_name: "testing.errors.io"
```

Additionly, `kubegen` simplifies definition format for resources within the modules. It keeps familiar YAML format, yet reduces nesting of certain fields to make it more intuitive to write a resource definition (perhaps even without having to consult docs or the one you wrote earlier).

For example, a front-end service in `errors.io` app has the following definition:
```YAML
Variables:
  - name: replicas
    type: number
    default: 2
  - name: domain_name
    type: string
    required: true

Deployments:
  - name: frontend
    replicas: <replicas>
    containers:
      - name: agent
        image: 'errordeveloper/errorsio-frontend'
        imagePullPolicy: IfNotPresent
        args:
          - '--domain=<domain_name>'
        ports:
          - name: http
            containerPort: 8080
Services:
  - name: frontend
    port: 8080
```

If you are not yet vert familiar with Kubernetes, this should be much easier to understand.
If you are already using Kubernetes, the rules of how this maps to a "native" format are really quite simple and are outlined down below.

## Usage

The main supported use-case of `kubegen` is for _generating_ files localy and checking in to a repository for use with other tools to implement CD, e.g. [Weave Flux](https://github.com/weaveworks/flux), but piping the output to `kubectl` is also supported for testing purposes.

TODO

## Conversion Rules

TODO

### Building

[![Build Status](https://travis-ci.org/errordeveloper/kubegen.svg?branch=master)](https://travis-ci.org/errordeveloper/kubegen)

Get the source code and build the dependencies:

```bash
go get github.com/Masterminds/glide
go get -d github.com/errordeveloper/kubegen
cd $GOPATH/src/github.com/errordeveloper/kubegen
$GOPATH/bin/glide up --strip-vendor
```

Build `kubegen`:
```bash
make
```
