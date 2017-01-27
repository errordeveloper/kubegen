all:
	@$(MAKE) install
	@$(MAKE) test
	@$(MAKE) build

install:
	go install ./appmaker

test:
	@$(MAKE) -C ./appmaker/assets
	@go test -v ./appmaker

build:
	go build ./

assets:
	go run $(ASSETS_GENERATOR) > $(ASSETS_MANIFEST)
