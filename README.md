# `kubegen` – simple way to describe Kubernetes resources

Kubernetes resource definitions are too verbose, and there is no built-in framework for reusability.
Writing good resource templates is hard, whether you are rolling your own or using Helm.

***The aim of this project is to make it easier to write reusable Kubernetes resource definitions.***

> It should be useful as is, but it's ambition is to drive the community towards an improvement
upstream (so I'd hope that Helm and other related project could make use of one standard format).
However, please note that it is **WORK IN PROGRESS** rigth now.

## Motivation & High-level Goals

First of all, it should be simple to define a Kubernetes resource, users should be able to specify key fields required
for a basic resource without refering to documentation.

Secondly, there should exist a simple model for defining collections of re-usable resources, let's call
them modules.

As a black box, a module could be described as the following:

***Given a set of input values, produce a set of Kubernetes resource that belong to one logical group.***

If one wanted to implement such black box they have the following to their disposal:

  1. simple model for re-usability of sub-components
  2. stateless model for input parameter substitution
  3. built-in output validation

Additionally, it should have the following properties:

  - simple and contsrained type system for input parameters
  - simple rules of inheritance and scoping
  - familiar syntax
  - all state is local
  - remote state can be obtain easily, but only to be used as input parameters
  - few simple helpers for reading local files to store as data
  - absence of any module/package management features
  - absence of resource dependency management system
  - absence of dangerous toys

## Current Implementation

Firtsly, `kubegen` provides a simple non-recursive system of modules, which allows you to define resource with
a few simple parameters once and instantiate those multiple times with different values for those parameters.

For example, you can use it to describe two different environments where your app runs.

The following _bundle_ instatiates the same _module_ twice. The module defintion is located in `SourceDir: module/myapp`
directory, and the generated Kubernetes resources will be written to `OutputDir: env/prod`.

```YAML
Modules:
  - Name: prodApp
    SourceDir: modules/myapp
    OutputDir: env/prod
    Variables:
      api_service_replicas: 100
      domain_name: errors.io
  - Name: testApp
    SourceDir: modules/myapp
    OutputDir: env/test
    Variables:
      api_service_replicas: 10
      use_rds: false
      domain_name: testing.errors.io
```

Additionly, `kubegen` simplifies the format of the definition format for resources within the modules.
It keeps familiar YAML format, yet reduces nesting of certain fields to make it more intuitive to write
a resource definition (perhaps even without having to consult docs or the one you wrote earlier).

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

If you are not yet very familiar with Kubernetes, this format should be much easier to write from memory.
If you are already using Kubernetes, the rules of how this maps to a "native" format are really quite
simple and are outlined down below.

#### Use-case

The main use-case for which `kubegen` caters right now is about de-duplicating resource definitions for a set
of environments, e.g. development and production.

`kubegen` is all about _generating_ files locally and checking in to a repository for use with other tools that
take care of managing releases (e.g. [Weave Flux](https://github.com/weaveworks/flux)).
Nothing stops you from finding other uses for it, and e.g. pipe the output to `kubectl` for testing, but it's not
recommended to rely on every version of `kubegen` to generate output cosisten with the output of the previous version,
as it is still in active development.

### Usage

#### Installation

You can build it from source, if you wish to hack on it, otherwise you can download [binaries from Equinox][dl].

[dl]: (https://dl.equinox.io/errordeveloper/kubegen/latest).

#### `kubegen module`

This command is designed for testing how individual modules render with different parameters supplied as command-line flags.

TODO: flags and usage with examples in the repo

#### `kubegen bundle`

TODO: description

TODO: flags and usage with examples in the repo

#### `kubegen self-upgrade`

This command allows you simply upgrade the binary you have downloaded to latest version.

### General Specification

There are 2 main layers in `kubegen`:

- _bundle_ provides a way of instatiating one or more _modules_
- _module_ is a collection of one or more YAML, JSON or HCL maifests

A manifest withing a module may contain the following top-level keys:

- `Variables`
- `Deployments`
- `Services`
- `DaemonSets`
- `ReplicaSets`
- `StatefulSets`
- `ConfigMaps`
- `Secrets`

Each of those keys is expected to contains a list of objects of the same type (as denoted by the key).

Variabels are scoped globaly per-module, and you can provide a varibale-only manifest, which is useful for denoting variables that are shared by all manifests withing a module.

A manifest is converted to `List` of objects defined within it and results in one file. In other words, module instance will result in as many native manifest files as there are manifests withing a module, unless variable-only manifests are used.

### Resource Conversion Rules

Broadly, `kubegen` flattens the most non-intuitive parts of a Kubernetes resource.
For example, a native `Deployment` object has `spec.template.spec.containers`, for `kubegen` that simply become `containers`.
Additionally, you shouldn't have to specify `metadata.name` along with `metadata.labels.name`, you simply set `name` along with optional `labels`, and selectors are also infered unless specified otherwise.

<!-- TODO more details or an example
Additionally, there are currently some minor details in how container and service ports are specified a little differently...
-->

#### HCL translation

`kubegen` is polyglot and supports HCL in addition to traditional Kubernetes JSON and YAML formats.

The style of HCL keys is a little different.
First of all, top-level keys are singular instead of plural, e.g.
```HCL
variable "my_replicas" {
  type = "string"
  required = true
}
```

All of attributes within `Deployment` or other resources use `snake_case` instead of `lowerCamel`, e.g.
```HCL
deployment "my_deployment" {
  labels {
    app = "my-app"
  }

  replicas = "<my_replicas>"

  container "main" {
    image = "myorg/app"
    image_pull_policy = "IfNotPresent"
  }
}
```

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
