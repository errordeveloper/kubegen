all:
	@$(MAKE) install
	@$(MAKE) test
	@$(MAKE) build

install:
	go install ./appmaker

ASSETS_GENERATOR := ./appmaker/assets/generate.go
ASSETS_MANIFEST := ./appmaker/assets/sockshop.json

test:
	go run $(ASSETS_GENERATOR) > $(ASSETS_MANIFEST).new
	diff -q $(ASSETS_MANIFEST) $(ASSETS_MANIFEST).new \
	  && rm -f $(ASSETS_MANIFEST).new \
	  || diff $(ASSETS_MANIFEST) $(ASSETS_MANIFEST).new
	go test -v ./appmaker

build:
	go build ./

assets:
	go run $(ASSETS_GENERATOR) > $(ASSETS_MANIFEST)
