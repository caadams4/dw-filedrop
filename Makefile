BIN := bin/dw-filedrop

.PHONY: build test fmt check

build:
	go build -o $(BIN) ./cmd/dw-filedrop

test:
	go test ./...

fmt:
	gofmt -l -w .

check: build
	$(BIN) validate testdata/sample_v5.csv
