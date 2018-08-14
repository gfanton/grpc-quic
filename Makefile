rwildcard = $(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2) $(filter $(subst *,%,$2),$d))
PROTOS = $(call rwildcard, $(wildcard proto/), *.proto)
GENERATED_FILES = $(patsubst %.proto,%.pb.go,$(PROTOS)) \

test:
	echo $(PROTOS)

.PHONY: generate
generate: $(GENERATED_FILES)

.PHONY: clean
clean: $(GENERATED_FILES)
	rm $(GENERATED_FILES)

%.pb.go: %.proto
	protoc $(PROTOC_OPTS) --gofast_out=plugins=grpc:"$(GOPATH)/src" "$(dir $<)"/*.proto
