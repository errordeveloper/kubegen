all:
	@$(MAKE) install
	@$(MAKE) test
	@$(MAKE) build

install:
	go install ./appmaker

test:
	@$(MAKE) -C ./appmaker/assets test
	@go test -v ./appmaker
	@./kubegen stack --manifest=test.hcl --output-format=json
	@./kubegen stack --manifest=test.hcl > test.yml.new \
	  && diff -q test.yml test.yml.new \
	  || diff test.yml test.yml.new
	@rm -f test.yml.new

build:
	go build ./

assets:
	@$(MAKE) -C ./appmaker/assets rebuild
