# Makefile used to build XDS daemon Web Server

# Application Version
VERSION := 0.0.1

# Syncthing version to install
SYNCTHING_VERSION = 0.14.28
# FIXME: use master while waiting a release that include #164
#SYNCTHING_INOTIFY_VERSION = 0.8.5
SYNCTHING_INOTIFY_VERSION=master


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


all: tools/syncthing build

.PHONY: build
build: xds webapp

xds:vendor scripts tools/syncthing/copytobin
	@echo "### Build XDS server (version $(VERSION), subversion $(SUB_VERSION))";
	@cd $(ROOT_SRCDIR); $(BUILD_ENV_FLAGS) go build $(VERBOSE_$(V)) -i -o $(LOCAL_BINDIR)/xds-server -ldflags "-X main.AppVersion=$(VERSION) -X main.AppSubVersion=$(SUB_VERSION)" .

test: tools/glide
	go test --race $(shell $(LOCAL_TOOLSDIR)/glide novendor)

vet: tools/glide
	go vet $(shell $(LOCAL_TOOLSDIR)/glide novendor)

fmt: tools/glide
	go fmt $(shell $(LOCAL_TOOLSDIR)/glide novendor)

run: build/xds tools/syncthing/copytobin
	$(LOCAL_BINDIR)/xds-server --log info -c config.json.in

debug: build/xds webapp/debug tools/syncthing/copytobin
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
install:
	@test -e $(LOCAL_BINDIR)/xds-server -a -d webapp/dist || { echo "Please execute first: make all\n"; exit 1; }
	@test -e $(LOCAL_BINDIR)/xds-start-server.sh -a -d $(LOCAL_BINDIR)/agl || { echo "Please execute first: make all\n"; exit 1; }
	@test -e $(LOCAL_BINDIR)/syncthing -a -e $(LOCAL_BINDIR)/syncthing-inotify || { echo "Please execute first: make all\n"; exit 1; }
	mkdir -p $(INSTALL_DIR) \
		&& cp -a $(LOCAL_BINDIR)/* $(INSTALL_DIR)
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
	@test -e $(LOCAL_TOOLSDIR)/syncthing -a -e $(LOCAL_TOOLSDIR)/syncthing-inotify  || { \
	mkdir -p $(LOCAL_TOOLSDIR); \
	DESTDIR=$(LOCAL_TOOLSDIR) \
	SYNCTHING_VERSION=$(SYNCTHING_VERSION) \
	SYNCTHING_INOTIFY_VERSION=$(SYNCTHING_INOTIFY_VERSION) \
	./scripts/get-syncthing.sh; }

.PHONY:
tools/syncthing/copytobin:
	@test -e $(LOCAL_TOOLSDIR)/syncthing -a -e $(LOCAL_TOOLSDIR)/syncthing-inotify || { echo "Please execute first: make tools/syncthing\n"; exit 1; }
	@cp -f $(LOCAL_TOOLSDIR)/syncthing* $(LOCAL_BINDIR)

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
