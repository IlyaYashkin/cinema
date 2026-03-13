GO_OUT := ./gen/

.PHONY: proto clean

proto:
	buf generate

clean:
	rm -rf $(GO_OUT)
