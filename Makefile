# Makefile used to build XDS daemon Web Server

# Application Version
VERSION := 0.0.1

# Syncthing version to install
SYNCTHING_VERSION = 0.14.27
# FIXME: use patched version while waiting integration of #165
#SYNCTHING_INOTIFY_VERSION = 0.8.5


# Retrieve git tag/commit to set sub-version string
ifeq ($(origin SUB_VERSION), undefined)
	SUB_VERSION := $(shell git describe --tags --always | sed 's/^v//')
	ifeq ($(SUB_VERSION), )
		SUB_VERSION=unknown-dev
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
LOCAL_TOOLSDIR := $(ROOT_SRCDIR)/tools


export GOPATH := $(shell go env GOPATH):$(ROOT_GOPRJ)
export PATH := $(PATH):$(LOCAL_TOOLSDIR)

VERBOSE_1 := -v
VERBOSE_2 := -v -x


all: build webapp

.PHONY: build
build: xds

xds:vendor scripts
	@echo "### Build XDS server (version $(VERSION), subversion $(SUB_VERSION))";
	@cd $(ROOT_SRCDIR); $(BUILD_ENV_FLAGS) go build $(VERBOSE_$(V)) -i -o $(LOCAL_BINDIR)/xds-server -ldflags "-X main.AppVersion=$(VERSION) -X main.AppSubVersion=$(SUB_VERSION)" .

test: tools/glide
	go test --race $(shell $(LOCAL_TOOLSDIR)/glide novendor)

vet: tools/glide
	go vet $(shell $(LOCAL_TOOLSDIR)/glide novendor)

fmt: tools/glide
	go fmt $(shell $(LOCAL_TOOLSDIR)/glide novendor)

run: build/xds tools/syncthing
	$(LOCAL_BINDIR)/xds-server --log info -c config.json.in

debug: build/xds webapp/debug tools/syncthing
	$(LOCAL_BINDIR)/xds-server --log debug -c config.json.in

.PHONY: clean
clean:
	rm -rf $(LOCAL_BINDIR)/* debug $(ROOT_GOPRJ)/pkg/*/$(REPOPATH)

.PHONY: distclean
distclean: clean
	rm -rf $(LOCAL_BINDIR) $(LOCAL_TOOLSDIR) glide.lock vendor webapp/node_modules webapp/dist

webapp: webapp/install
	(cd webapp && gulp build)

webapp/debug:
	(cd webapp && gulp watch &)

webapp/install:
	(cd webapp && npm install)

.PHONY: scripts
scripts:
	@mkdir -p $(LOCAL_BINDIR) && cp -rf scripts/xds-start-server.sh scripts/agl $(LOCAL_BINDIR)

.PHONY: install
install: all scripts tools/syncthing
	mkdir -p $(INSTALL_DIR) \
		&& cp $(LOCAL_BINDIR)/* $(INSTALL_DIR) \
		&& cp $(LOCAL_TOOLSDIR)/syncthing* $(INSTALL_DIR)
	mkdir -p $(INSTALL_WEBAPP_DIR) \
		&& cp -a webapp/dist/* $(INSTALL_WEBAPP_DIR)

vendor: tools/glide glide.yaml
	$(LOCAL_TOOLSDIR)/glide install --strip-vendor

tools/glide:
	@echo "Downloading glide"
	mkdir -p $(LOCAL_TOOLSDIR)
	curl --silent -L https://glide.sh/get | GOBIN=$(LOCAL_TOOLSDIR)  sh

.PHONY: tools/syncthing
tools/syncthing:
	@(test -s $(LOCAL_TOOLSDIR)/syncthing || \
	DESTDIR=$(LOCAL_TOOLSDIR) \
	SYNCTHING_VERSION=$(SYNCTHING_VERSION) \
	SYNCTHING_INOTIFY_VERSION=$(SYNCTHING_INOTIFY_VERSION) \
	./scripts/get-syncthing.sh)

.PHONY: help
help:
	@echo "Main supported rules:"
	@echo "  all                (default)"
	@echo "  build"
	@echo "  install"
	@echo "  clean"
	@echo "  distclean"
	@echo ""
	@echo "Influential make variables:"
	@echo "  V                 - Build verbosity {0,1,2}."
	@echo "  BUILD_ENV_FLAGS   - Environment added to 'go build'."
