TARGET=netm4ul
GO_LIST=$(shell go list ./... 2>&1 | grep -v /vendor/ | grep -v "permission denied")

all: vet fmt deps build
	@echo "All done"

test: all
	@go test $(GO_LIST)

build:
	@echo "Building ..."
	@go build $(FLAGS) -o $(TARGET) .
	@echo "Building done"

vet:
	@go vet $(GO_LIST)

fmt:
	@go fmt $(GO_LIST)

lint:
	@golint $(GO_LIST)

deps:
	@echo "Ensure dependencies"
	@dep ensure

clean:
	@rm -f $(TARGET)
