test:
	@for i in basic combined ; do \
	  go run generate-$$i.go > sockshop-$$i.json.new ;\
	  diff -q sockshop-$$i.json sockshop-$$i.json.new \
	    && rm -f sockshop-$$i.json.new \
  	    || diff sockshop-$$i.json sockshop-$$i.json.new ;\
	done

rebuild:
	@for i in basic combined ; do go run generate-$$i.go > sockshop-$$i.json ; done
