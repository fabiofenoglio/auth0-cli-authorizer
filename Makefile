build:
	go build -o ./dist/main
run:
	go run main.go
lint:
	golangci-lint run
test:
	go test ./...
clean:
	go mod tidy
	go fmt $(go list ./... | grep -v /vendor/)
check:
	make build
	make clean
	make lint
	make test
