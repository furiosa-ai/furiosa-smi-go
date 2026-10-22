.PHONY: all
all: build fmt lint vet test tidy

.PHONY: build
build:
	go build ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: vet
vet:
	go vet -v ./...

.PHONY: test
test:
	go test ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: example
example:
	@mkdir -p example/bin
	@for dir in example/*/; do \
		name=$$(basename $$dir); \
		[ "$$name" = "bin" ] && continue; \
		go build -o example/bin/$$name ./$$dir || exit $$?; \
		echo "Built example/bin/$$name"; \
	done
	@go build -o example/bin/example ./example && echo "Built example/bin/example"
