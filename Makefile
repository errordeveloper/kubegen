test: build
	@$(MAKE) -C ./pkg/apps/assets test
	@go test -v ./pkg/apps
	@go test -v ./cmd/kubegen
	@go test -v ./cmd/kubegen-experiment-appgen

build: install
	@for cmd in kubegen kubegen-experiment-appgen kubegen-experiment-stack ; do go build ./cmd/$${cmd}/ ; done

install:
	@for pkg in apps resources util ; do go install ./pkg/$${pkg}/ ; done

assets:
	@$(MAKE) -C ./pkg/apps/assets rebuild
	@$(MAKE) -C ./cmd/kubegen/assets rebuild
	@$(MAKE) -C ./cmd/kubegen-experiment-appgen/assets rebuild
