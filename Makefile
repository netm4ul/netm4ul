TARGET=netm4ul

all: fmt vet deps build
	@echo "All done"

test: build
	@go test ./...

build:
	@echo "Building ..."
	@go build $(FLAGS) -o $(TARGET) .
	@echo "Building done"

vet:
	@go vet ./...

fmt:
	@go fmt ./...

lint:
	@golint ./...

deps:
	@echo "Ensure dependencies"
	@dep ensure

clean:
	@rm -f $(TARGET)
