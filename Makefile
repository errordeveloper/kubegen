test: build
	@$(MAKE) -C ./appmaker/assets test
	@go test -v ./appmaker
	@rm -f test.yml.new

build: install
	@go build ./

install:
	@go install ./appmaker

assets:
	@$(MAKE) -C ./appmaker/assets rebuild
