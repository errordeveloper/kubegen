test: build
	@$(MAKE) -C ./pkg/apps/assets test
	@go test -v ./pkg/apps

build: install
	@for cmd in kubegen-test-stack kubegen-test-stack ; do go build ./cmd/$${cmd}/ ; done

install:
	@for pkg in apps resources util ; do go install ./pkg/$${pkg}/ ; done

assets:
	@$(MAKE) -C ./pkg/apps/assets rebuild
