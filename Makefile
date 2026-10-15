.PHONY: check fmt vet test build install

# check is what every step must pass.
check: fmt vet test

fmt:
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

vet:
	go vet ./...

test:
	go test ./...

build:
	go build -o bin/mia ./cmd/mia

install:
	go install ./cmd/mia
