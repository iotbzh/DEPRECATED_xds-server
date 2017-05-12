# Makefile used to build XDS daemon Web Server

# Syncthing version to install
SYNCTHING_VERSION = 0.14.25
SYNCTHING_INOTIFY_VERSION = 0.8.5

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
	INSTALL_WEBAPP_DIR := $(INSTALL_DIR)/xds-server-www
endif

HOST_GOOS=$(shell go env GOOS)
HOST_GOARCH=$(shell go env GOARCH)
REPOPATH=github.com/iotbzh/xds-server

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
ROOT_SRCDIR := $(patsubst %/,%,$(dir $(mkfile_path)))
ROOT_GOPRJ := $(abspath $(ROOT_SRCDIR)/../../../..)
LOCAL_BINDIR := $(ROOT_SRCDIR)/bin

export GOPATH := $(shell go env GOPATH):$(ROOT_GOPRJ)
export PATH := $(PATH):$(ROOT_SRCDIR)/tools

VERBOSE_1 := -v
VERBOSE_2 := -v -x


all: build webapp

build: build/xds

xds: build/xds

build/xds: vendor scripts
	@echo "### Build XDS server (version $(VERSION))";
	@cd $(ROOT_SRCDIR); $(BUILD_ENV_FLAGS) go build $(VERBOSE_$(V)) -i -o $(LOCAL_BINDIR)/xds-server -ldflags "-X main.AppVersionGitTag=$(VERSION)" .

test: tools/glide
	go test --race $(shell ./tools/glide novendor)

vet: tools/glide
	go vet $(shell ./tools/glide novendor)

fmt: tools/glide
	go fmt $(shell ./tools/glide novendor)

run: build/xds tools/syncthing
	$(LOCAL_BINDIR)/xds-server --log info -c config.json.in

debug: build/xds webapp/debug tools/syncthing
	$(LOCAL_BINDIR)/xds-server --log debug -c config.json.in

.PHONY: clean
clean:
	rm -rf $(LOCAL_BINDIR)/* debug cmd/*/debug $(ROOT_GOPRJ)/pkg/*/$(REPOPATH)

.PHONY: distclean
distclean: clean
	rm -rf $(LOCAL_BINDIR) tools glide.lock vendor cmd/*/vendor webapp/node_modules webapp/dist

run3:
	goreman start

webapp: webapp/install
	(cd webapp && gulp build)

webapp/debug:
	(cd webapp && gulp watch &)

webapp/install:
	(cd webapp && npm install)

.PHONY: scripts
scripts:
	@mkdir -p $(LOCAL_BINDIR) && cp -f scripts/xds-start-server.sh $(LOCAL_BINDIR)

.PHONY: install
install: all scripts tools/syncthing
	mkdir -p $(INSTALL_DIR) && cp $(LOCAL_BINDIR)/* $(INSTALL_DIR)
	mkdir -p $(INSTALL_WEBAPP_DIR) && cp -a webapp/dist/* $(INSTALL_WEBAPP_DIR)

vendor: tools/glide glide.yaml
	./tools/glide install --strip-vendor

tools/glide:
	@echo "Downloading glide"
	mkdir -p tools
	curl --silent -L https://glide.sh/get | GOBIN=./tools  sh

.PHONY: tools/syncthing
tools/syncthing:
	@(test -s $(LOCAL_BINDIR)/syncthing || \
	DESTDIR=$(LOCAL_BINDIR) \
	SYNCTHING_VERSION=$(SYNCTHING_VERSION) \
	SYNCTHING_INOTIFY_VERSION=$(SYNCTHING_INOTIFY_VERSION) \
	./scripts/get-syncthing.sh)

.PHONY: help
help:
	@echo "Main supported rules:"
	@echo "  build               (default)"
	@echo "  build/xds"
	@echo "  release"
	@echo "  clean"
	@echo "  distclean"
	@echo ""
	@echo "Influential make variables:"
	@echo "  V                 - Build verbosity {0,1,2}."
	@echo "  BUILD_ENV_FLAGS   - Environment added to 'go build'."
