PACKAGE_NAME=github.com/netm4ul/netm4ul/
TARGET=netm4ul

GO_LIST=$(shell go list ./... 2>&1 | grep -v /vendor/ | grep -v "permission denied")
PACKAGES=$(shell echo $(GO_LIST) | sed -e "s!$(PACKAGE_NAME)!!g" | sed -e "s!github.com/netm4ul/netm4ul!!g")

.PHONY: all
all: vet fmt deps build
	@echo "All done"

.PHONY: test
test:
	@echo "mode: atomic" > coverage.profile
	@for pkg in $(PACKAGES); do \
		touch $$pkg.profile ; \
		go test -race ./$$pkg -coverprofile=$$pkg.profile -covermode=atomic; \
		tail -n +2 $$pkg.profile >> coverage.profile && rm -rf $$pkg.profile ; \
	done

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

.PHONY: gofmt
gofmt:
	@gofmt -s -w $(GO_LIST)

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
