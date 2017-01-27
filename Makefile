all:
	@$(MAKE) install
	@$(MAKE) test
	@$(MAKE) build

install:
	go install ./appmaker

test:
	@$(MAKE) -C ./appmaker/assets test
	@go test -v ./appmaker

build:
	go build ./

assets:
	@$(MAKE) -C ./appmaker/assets rebuild
