.PHONY: build clean test help default tag fmt vendor vet install release

BIN_NAME := bin/remco

GOARCH ?= amd64

VERSION := 0.12.1
GIT_COMMIT := $(shell git rev-parse HEAD)
GIT_DIRTY := $(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
BUILD_DATE := $(shell date '+%Y-%m-%d-%H:%M:%S')

SYSCONFDIR := /etc/
PREFIX := /usr/local
BINDIR := ${PREFIX}/bin

DEFAULTCONFDIR := ${SYSCONFDIR}/remco/
DEFAULTCONF := ${DEFAULTCONFDIR}/config


GO_SRC := $(shell find ./ -type f -name '*.go' -and -not -name '*_test.go')
GO_TEST_SRC := $(shell find ./ -type f -name '*_test.go')

GO := go

GO_OPTS := -mod=mod

OS_LIST := linux darwin windows

OUT_RELEASE_ZIP := $(addsuffix _$(GOARCH).zip, $(addprefix bin/remco_$(VERSION)_, $(OS_LIST)))

default: build

help:
	@echo 'Management commands for remco:'
	@echo
	@echo 'Usage:'
	@echo '    make build           Compile the project.'
	@echo '    make release         Create all the releases for [$(OS_LIST)]'
	@echo '    make test            Run the unit tests.'
	@echo '    make vendor          Recover the deps (put them in /vendor)'
	@echo '    make fmt             use go fmt on the code.'
	@echo '    make vet             use go vet on the code.'
	@echo '    make clean           Clean the directory tree.'
	@echo

build: ${BIN_NAME}

${BIN_NAME}: $(GO_SRC)
	@echo "building ${BIN_NAME} ${VERSION}"
	$(GO) build -a -tags netgo -ldflags "-s -w -X main.version=${VERSION} \
		-X main.buildDate=${BUILD_DATE} \
		-X main.commit=${GIT_COMMIT}${GIT_DIRTY}" \
		-o ${BIN_NAME} ${GO_OPTS} ./cmd/remco/

vendor:
	@echo "recovering/vendoring the dependencies"
	$(GO) mod vendor

clean:
	@echo "Cleaning up"
	@test ! -e ${BIN_NAME} || rm ${BIN_NAME}
	@test ! -e coverage.out || rm coverage.out
	@test ! -e coverage.html || rm coverage.html
	@rm -f bin/*.zip
	@rm -f bin/remco_*

test-browser-cov: test
	$(GO) tool cover -html=coverage.out

test: coverage.out

fmt:
	$(GO) fmt ...

vet:
	$(GO) vet ...

coverage.out: $(GO_SRC) $(GO_TEST_SRC) build
	@echo "Running the test"
	$(GO) test ./... -race ${GO_OPTS} -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "report available in coverage.html"

install: build
	@echo "Installing remco"
	@mkdir -p ${DESTDIR}/${BINDIR}
	@mkdir -p ${DESTDIR}/${DEFAULTCONFDIR}
	@install -m 755 ${BIN_NAME} ${DESTDIR}/${BINDIR}
	@if ! [ -e "${DESTDIR}/${DEFAULTCONF}" ];\
	then \
		install -m 640 ./integration/file/file.toml ${DESTDIR}/${DEFAULTCONF};\
	else \
		echo "conf  file '${DESTDIR}/${DEFAULTCONF}' already present";\
	fi

tag:
	git tag -a v${VERSION} -m "version ${VERSION}"
	git push origin v${VERSION}

release: $(OUT_RELEASE_ZIP)

$(OUT_RELEASE_ZIP): $(GO_SRC)
	GOARCH=$(GOARCH) CGO_ENABLED=0 GOOS=$(subst bin/remco_${VERSION}_,,$(subst _$(GOARCH).zip,,$@)) \
	     $(MAKE) build \
	     BIN_NAME=$(subst ${VERSION}_,,$(subst _$(GOARCH).zip,,$@))
	cd bin && zip -r $(shell basename $@) $(shell basename $(subst ${VERSION}_,,$(subst _$(GOARCH).zip,,$@)))
