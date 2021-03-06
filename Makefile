PACKAGE_NAME=github.com/netm4ul/netm4ul/
TARGET=netm4ul

GO_LIST=$(shell go list ./... 2>&1 | grep -v /vendor/ | grep -v "permission denied")
PACKAGES=$(shell echo $(GO_LIST) | sed -e "s!$(PACKAGE_NAME)!!g" | sed -e "s!github.com/netm4ul/netm4ul!!g")

.PHONY: all
all: vet fmt build
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
	@go get -t -v ./...
	@go build $(FLAGS) -o $(TARGET) .
	@echo "Building done"

.PHONY: vet
vet:
	@go vet ./...

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: gofmt
gofmt:
	@gofmt -s -w .

.PHONY: lint
lint:
	@golint ./...

.PHONY: docker-build
docker-build: build
	@./Dockerfiles/build-all.sh

.PHONY: docker-publish
docker-publish: docker-build
	@./Dockerfiles/publish-all.sh

.PHONY: clean
clean:
	@rm -f $(TARGET)
