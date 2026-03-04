PROTO_DIR := proto
GO_OUT := ./gen/

PROTO_FILES := $(shell find $(PROTO_DIR) -name '*.proto')

PB_GO_FILES := $(PROTO_FILES:$(PROTO_DIR)/%.proto=$(GO_OUT)%.pb.go)
PB_GRPC_GO_FILES := $(PROTO_FILES:$(PROTO_DIR)/%.proto=$(GO_OUT)%_grpc.pb.go)
PB_ALL := $(PB_GO_FILES) $(PB_GRPC_GO_FILES)

.PHONY: proto clean

proto: $(PB_GO_FILES)

$(GO_OUT)%.pb.go: $(PROTO_DIR)/%.proto
	@mkdir -p $(dir $@)
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(GO_OUT) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(GO_OUT) \
		--go-grpc_opt=paths=source_relative \
		$<

clean:
	rm -f $(PB_ALL)
