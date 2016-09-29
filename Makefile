BINARY=lapinou
VERSION=0.1.0

.DEFAULT_GOAL: $(BINARY)

$(BINARY):
	go build -o ${BINARY} main.go

deb:
	fpm -s dir -t deb -n $(BINARY) -v $(VERSION) --prefix /usr/local/bin lapinou

.PHONY: clean
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
	if [ -f ${BINARY}_${VERSION}_amd64.deb ] ; then rm ${BINARY}_${VERSION}_amd64.deb ; fi
