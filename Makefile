test: build
	@$(MAKE) -C ./pkg/apps/assets test
	@go test -v ./pkg/apps
	@go test -v ./cmd/kubegen-experiment-appgen
	@$(MAKE) test-cmds

build: install
	@for cmd in kubegen kubegen kubegen-experiment-appgen kubegen-experiment-stack ; do go build ./cmd/$${cmd}/ ; done

install:
	@for pkg in apps resources util ; do go install ./pkg/$${pkg}/ ; done

assets:
	@$(MAKE) -C ./pkg/apps/assets rebuild
	@$(MAKE) -C ./cmd/kubegen/assets rebuild
	@$(MAKE) -C ./cmd/kubegen-experiment-appgen/assets rebuild

image: Boxfile
	@docker run --rm -ti \
	  -v $(PWD):$(PWD) \
	  -v /var/run/docker.sock:/var/run/docker.sock \
	  -w $(PWD) \
	    erikh/box:master Boxfile --debug
	@docker push errordeveloper/kubegen

test-cmds:
	@for cmd in kubegen ; do cd cmd/$${cmd} ; \
            ln -sf ../../examples .examples ; \
	    ln -sf ../../../examples assets/.examples ; \
	    go test -v ; \
	    rm -f .examples ; \
	    rm -f assets/.examples ; \
	done
