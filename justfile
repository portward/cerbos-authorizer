default:
    just --list

test:
    CERBOS_ADDRESS=127.0.0.1:3593 go test -race -v ./...

lint:
    golangci-lint run

fmt:
    golangci-lint run --fix
