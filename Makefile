BINARY=lapinou

.DEFAULT_GOAL: $(BINARY)

$(BINARY):
	go build -o ${BINARY} main.go

.PHONY: clean
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
