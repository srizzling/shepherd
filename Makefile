GO_FILES := $(find . -iname '*.go' -type f | grep -v /vendor/)
BINARY = shepherd
BUILD_DIR := "build"
GOARCH = amd64

TEST_REPORT      = $(BUILD_DIR)/tests.xml
COVERAGE_DIR 	 = $(BUILD_DIR)/coverage
COVERAGE_MODE    = atomic
COVERAGE_PROFILE = $(COVERAGE_DIR)/profile.out
COVERAGE_XML     = $(COVERAGE_DIR)/coverage.xml
COVERAGE_HTML    = $(COVERAGE_DIR)/index.html

VERSION?=?
COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.VERSION=${VERSION} -X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH}"

all: deps lint build test

clean:
	@rm -rf $(BUILD_DIR)/*

linux:
	GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o $(BUILD_DIR)/${BINARY}-linux-${GOARCH}

darwin:
	GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o $(BUILD_DIR)/${BINARY}-windows-${GOARCH}

windows:
	GOOS=windows GOARCH=${GOARCH} go build ${LDFLAGS} -o $(BUILD_DIR)/${BINARY}-windows-${GOARCH}

build-all: linux darwin windows

build: clean
	@mkdir -p build && go build -o build/$(BINARY)

run: build
	./build/$(BINARY)

lint:
	@test -z $(gofmt -s -l $GO_FILES)
	@go vet $(GO_FILES)
	@megacheck $(GO_FILES)
	@golint -set_exit_status $(GO_FILES)

test:
	@go test -v -race $(GO_FILES)

coverage:
	@mkdir -p $(COVERAGE_DIR)
	@go test -v -race $(GO_FILES) -coverprofile $(COVERAGE_PROFILE) -covermode=$(COVERAGE_MODE)
	go tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_HTML)

format:
	@gofmt -l -s $(GO_FILES)

tools:
	@go get -u github.com/golang/dep/cmd/dep
	@go get -u github.com/golang/lint/golint
	@go get -u honnef.co/go/tools/cmd/megacheck
	@go get -u golang.org/x/tools/cmd/cover
	@go get -u github.com/tebeka/go2xunit
	@go get -u github.com/goreleaser/goreleaser

deps:
	@dep ensure

release:
	@goreleaser
