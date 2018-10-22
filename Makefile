TARGET=netm4ul
GO_LIST=$(shell go list ./... 2>&1 | grep -v /vendor/ | grep -v "permission denied")

.PHONY: all
all: vet fmt deps build
	@echo "All done"

.PHONY: test
test:
	@go test $(GO_LIST)

.PHONY: build
build:
	@echo "Building ..."
	@go build $(FLAGS) -o $(TARGET) .
	@echo "Building done"

.PHONY: vet
vet:
	@go vet $(GO_LIST)

.PHONY: fmt
fmt:
	@go fmt $(GO_LIST)

.PHONY: lint
lint:
	@golint $(GO_LIST)

.PHONY: deps
deps:
	@echo "Ensure dependencies"
	@dep ensure

.PHONY: clean
clean:
	@rm -f $(TARGET)
