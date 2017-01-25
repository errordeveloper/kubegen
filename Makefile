all:
	@$(MAKE) install
	@$(MAKE) test
	@$(MAKE) build

install:
	go install ./appmaker

test:
	go test -v ./appmaker

build:
	go build ./

assets:
	go run ./appmaker/assets/generate.go > ./appmaker/assets/sockshop.json
