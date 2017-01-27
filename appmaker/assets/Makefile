test:
	@for i in basic combined ; do \
	  go run generate-$$i.go > sockshop-$$i.json.new ;\
	  diff -q sockshop-$$i.json sockshop-$$i.json.new \
  	    || diff sockshop-$$i.json sockshop-$$i.json.new ;\
	done
	@rm -f sockshop-*.json.new

rebuild:
	@for i in basic combined ; do go run generate-$$i.go > sockshop-$$i.json ; done
