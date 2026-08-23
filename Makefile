GOTOOLCHAIN ?= local

run:
	GOTOOLCHAIN=$(GOTOOLCHAIN) go run ./cmd/server

test:
	GOTOOLCHAIN=$(GOTOOLCHAIN) go test ./... -count=1

race:
	GOTOOLCHAIN=$(GOTOOLCHAIN) go test -race ./... -count=1

vet:
	GOTOOLCHAIN=$(GOTOOLCHAIN) go vet ./...

build:
	GOTOOLCHAIN=$(GOTOOLCHAIN) go build ./...

