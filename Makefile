test: build
	@go test -v ./pkg/...
	@$(MAKE) test-cmds

install:
	@go install ./pkg/...

build: install
	@go build ./cmd/...

assets:
	@$(MAKE) -C ./cmd/kubegen/assets rebuild

image: Boxfile
	@docker run --rm -ti \
	  -v $(PWD):$(PWD) \
	  -v /var/run/docker.sock:/var/run/docker.sock \
	  -w $(PWD) \
	    erikh/box:master Boxfile --debug

publish: image
	@docker push errordeveloper/kubegen

test-cmds:
	@for cmd in kubegen ; do cd cmd/$${cmd} ; \
            ln -sf ../../examples .examples ; \
	    ln -sf ../../../examples assets/.examples ; \
	    go test -v ; \
	    rc=$$? ; \
	    rm -f .examples ; \
	    rm -f assets/.examples ; \
	    exit $${rc} ; \
	done
