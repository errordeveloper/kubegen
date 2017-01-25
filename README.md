# `kubegen` – generate Kubernetes resources from simple app definitions

### Building

Get the source code and build the dependencies:

```bash
go get github.com/Masterminds/glide
go get -d github.com/errordeveloper/kubegen
cd $GOPATH/src/github.com/errordeveloper/kubegen
$GOPATH/bin/glide up
go install ./appmaker
```

Build `kubegen`:
```bash
go build .
```
