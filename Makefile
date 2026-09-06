.PHONY: build
build:
	go build -o dist/maketui .

.PHONY: fmt
## Format the code with gofmt
fmt:
	gofmt -w .
