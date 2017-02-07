# `kubegen` – simple way to describe Kubernetes resources

The aim of this project is to make it _easier to write reusable_ Kubernetes resource definitions.


***WORK IN PROGRESS***

<!--

There is no clear answer to re-usability and de-duplication for Kubernetes resource definitions.

The aim of `kubegen` is to offer the community with a few different options, and we will decide later
which one is level of abstraction is best, or if there is a need for abstractions at different levels.

## Service/Deployment pair based on container image

```
kg component-from-image

## CLI Usage

You can generate a simple app with just one flag, the image name:
```
kubegen single --image "errordeveloper/foo:latest"
```

This will geneate app named `foo`, but if you'd like to call it something else,
you can pass `--name` flag.

By default, a `Deployment`/`Service` pair is generated with a number of known-best-practice
options included for you, e.g. liveness and readiness probes as well as Prometheus annotations.
You can change this behaviour by using high-lvel `--flavor` flag or one of more specific flags
described below.

-->

### Building

[![Build Status](https://travis-ci.org/errordeveloper/kubegen.svg?branch=master)](https://travis-ci.org/errordeveloper/kubegen)

Get the source code and build the dependencies:

```bash
go get github.com/Masterminds/glide
go get -d github.com/errordeveloper/kubegen
cd $GOPATH/src/github.com/errordeveloper/kubegen
$GOPATH/bin/glide up --strip-vendor
go install ./appmaker
```

Build `kubegen`:
```bash
go build .
```
