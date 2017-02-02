test:
	@../../../kubegen-test-stack --manifest=sockshop.hcl > sockshop.yml.new \
	  && diff -q sockshop.yml sockshop.yml.new \
	  || diff sockshop.yml sockshop.yml.new
	@for i in basic combined ; do \
	  go run generate-$$i.go > sockshop-$$i.json.new ;\
	  diff -q sockshop-$$i.json sockshop-$$i.json.new \
  	    || diff sockshop-$$i.json sockshop-$$i.json.new ;\
	done
	@rm -f sockshop-*.json.new sockshop.yml.new

rebuild:
	@for i in basic combined ; do go run generate-$$i.go > sockshop-$$i.json ; done
	@../../../kubegen-test-stack --manifest=sockshop.hcl > sockshop.yml
