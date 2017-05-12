# Makefile used to build XDS daemon Web Server

# Retrieve git tag/commit to set sub-version string
ifeq ($(origin VERSION), undefined)
	VERSION := $(shell git describe --tags --always | sed 's/^v//')
	ifeq ($(VERSION), )
		VERSION=unknown-dev
	endif
endif

# Configurable variables for installation (default /usr/local/...)
ifeq ($(origin INSTALL_DIR), undefined)
	INSTALL_DIR := /usr/local/bin
endif
ifeq ($(origin INSTALL_WEBAPP_DIR), undefined)
	INSTALL_WEBAPP_DIR := ${INSTALL_DIR}/xds-server-www
endif

HOST_GOOS=$(shell go env GOOS)
HOST_GOARCH=$(shell go env GOARCH)
REPOPATH=github.com/iotbzh/xds-server

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
ROOT_SRCDIR := $(patsubst %/,%,$(dir $(mkfile_path)))
ROOT_GOPRJ := $(abspath $(ROOT_SRCDIR)/../../../..)

export GOPATH := $(shell go env GOPATH):$(ROOT_GOPRJ)
export PATH := $(PATH):$(ROOT_SRCDIR)/tools

VERBOSE_1 := -v
VERBOSE_2 := -v -x

#WHAT := xds-make

all: build webapp

#build: build/xds build/cmds
build: build/xds

xds: build/xds

build/xds: vendor
	@echo "### Build XDS server (version $(VERSION))";
	@cd $(ROOT_SRCDIR); $(BUILD_ENV_FLAGS) go build $(VERBOSE_$(V)) -i -o bin/xds-server -ldflags "-X main.AppVersionGitTag=$(VERSION)" .

#build/cmds: vendor
#	@for target in $(WHAT); do \
#		echo "### Build $$target"; \
#		$(BUILD_ENV_FLAGS) go build $(VERBOSE_$(V)) -i -o bin/$$target -ldflags "-X main.AppVersionGitTag=$(VERSION)" ./cmd/$$target; \
#	done

test: tools/glide
	go test --race $(shell ./tools/glide novendor)

vet: tools/glide
	go vet $(shell ./tools/glide novendor)

fmt: tools/glide
	go fmt $(shell ./tools/glide novendor)

run: build/xds
	./bin/xds-server --log info -c config.json.in

debug: build/xds webapp/debug
	./bin/xds-server --log debug -c config.json.in

clean:
	rm -rf ./bin/* debug cmd/*/debug $(ROOT_GOPRJ)/pkg/*/$(REPOPATH)

distclean: clean
	rm -rf bin tools glide.lock vendor cmd/*/vendor webapp/node_modules webapp/dist

run3:
	goreman start

webapp: webapp/install
	(cd webapp && gulp build)

webapp/debug:
	(cd webapp && gulp watch &)

webapp/install:
	(cd webapp && npm install)


install: all
	mkdir -p ${INSTALL_DIR} && cp bin/xds-server ${INSTALL_DIR}
	mkdir -p ${INSTALL_WEBAPP_DIR} && cp -a webapp/dist/* ${INSTALL_WEBAPP_DIR}

vendor: tools/glide glide.yaml
	./tools/glide install --strip-vendor

tools/glide:
	@echo "Downloading glide"
	mkdir -p tools
	curl --silent -L https://glide.sh/get | GOBIN=./tools  sh

goenv:
	@go env

help:
	@echo "Main supported rules:"
	@echo "  build               (default)"
	@echo "  build/xds"
	@echo "  build/cmds"
	@echo "  release"
	@echo "  clean"
	@echo "  distclean"
	@echo ""
	@echo "Influential make variables:"
	@echo "  V                 - Build verbosity {0,1,2}."
	@echo "  BUILD_ENV_FLAGS   - Environment added to 'go build'."
#	@echo "  WHAT              - Command to build. (e.g. WHAT=xds-make)"
