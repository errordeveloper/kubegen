test:
	@for i in basic combined ; do \
	  go run generate-$$i.go > sockshop-$$i.json.new ;\
	  diff -q sockshop-$$i.json sockshop-$$i.json.new \
  	    || diff sockshop-$$i.json sockshop-$$i.json.new ;\
	done
	@../../kubegen stack --manifest=sockshop.hcl > sockshop.yml.new \
	  && diff -q sockshop.yml sockshop.yml.new \
	  || diff sockshop.yml sockshop.yml.new
	@rm -f sockshop-*.json.new sockshop.yml.new

rebuild:
	@for i in basic combined ; do go run generate-$$i.go > sockshop-$$i.json ; done
	@../../kubegen stack --manifest=sockshop.hcl > sockshop.yml
