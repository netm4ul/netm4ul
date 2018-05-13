TARGET=netm4ul
GO_FILES=$(shell find . -iname '*.go' -type f 2>&1 -not -path "./vendor/*" | grep -v "Permission denied")
GO_TEST_PKG=$(shell find . -iname '*_test.go' -type f 2>&1 -not -path "./vendor/*" -exec dirname {} \; | grep -v "Permission denied")
all: fmt vet deps build
	@echo "All done"

test: build
	@go test $(GO_TEST_PKG)

build:
	@echo "Building ..."
	@go build $(FLAGS) -o $(TARGET) .
	@echo "Building done"

vet:
	@go vet $(GO_FILES)

fmt:
	@go fmt $(GO_FILES)

lint:
	@golint $(GO_FILES)

deps:
	@echo "Ensure dependencies"
	@dep ensure

clean:
	@rm -f $(TARGET)
