.PHONY: build
build:
	go build -o dist/maketui .

.PHONY: fmt
fmt:
	gofmt -w .
