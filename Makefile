build:
	@go build -o bin/qs

run: build
	@./bin/qs

test: 
	@go test ./... -v