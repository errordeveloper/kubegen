test: build
	@$(MAKE) -C ./pkg/appmaker/assets test
	@go test -v ./pkg/appmaker

build: install
	@go build ./

install:
	@go install ./pkg/appmaker

assets:
	@$(MAKE) -C ./pkg/appmaker/assets rebuild
