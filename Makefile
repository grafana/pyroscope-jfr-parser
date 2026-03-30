GO ?= go
GO_FLAGS ?= -v -race

.phony: generate-types
generate-types:
	$(GO) run ./internal/cmd/gen
	$(GO) fmt ./parser/types

.PHONY: test
test:
	$(GO) test $(GO_FLAGS) $(shell $(GO) list ./...)

FUZZ_TIME ?= 10m

.PHONY: fuzz
fuzz:
	bash scripts/fuzz-corpus.sh
	$(GO) test -run=^$$ -fuzz=FuzzParseJFR -fuzztime=$(FUZZ_TIME) ./pprof/

