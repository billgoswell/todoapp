.PHONY: build test clean run help

build:
	go build -o todo ./cmd/app

test:
	go test ./...

run: build
	./todo

clean:
	rm -f todo

help:
	@echo "Available commands:"
	@echo "  make build  - Build the todo binary"
	@echo "  make test   - Run tests"
	@echo "  make run    - Build and run the todo app"
	@echo "  make clean  - Remove the built binary"
	@echo "  make help   - Show this help message"
