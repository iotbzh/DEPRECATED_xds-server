# Makefile used to build XDS daemon Web Server

# Application Version
VERSION := 0.1.0

# Syncthing version to install
SYNCTHING_VERSION = 0.14.28
# FIXME: use master while waiting a release that include #164
#SYNCTHING_INOTIFY_VERSION = 0.8.5
SYNCTHING_INOTIFY_VERSION=master


# Retrieve git tag/commit to set sub-version string
ifeq ($(origin SUB_VERSION), undefined)
	SUB_VERSION := $(shell git describe --exact-match --tags 2>/dev/null | sed 's/^v//')
	ifneq ($(SUB_VERSION), )
		VERSION := $(firstword $(subst -, ,$(SUB_VERSION)))
		SUB_VERSION := $(word 2,$(subst -, ,$(SUB_VERSION)))
	else
		SUB_VERSION := $(shell git describe --tags --always  | sed 's/^v//')
		ifeq ($(SUB_VERSION), )
			SUB_VERSION := unknown-dev
		endif
	endif
endif

# Configurable variables for installation (default /usr/local/...)
ifeq ($(origin INSTALL_DIR), undefined)
	INSTALL_DIR := /usr/local/bin
endif
ifeq ($(origin INSTALL_WEBAPP_DIR), undefined)
	INSTALL_WEBAPP_DIR := $(INSTALL_DIR)/www-xds-server
endif

HOST_GOOS=$(shell go env GOOS)
HOST_GOARCH=$(shell go env GOARCH)
ARCH=$(HOST_GOOS)-$(HOST_GOARCH)
REPOPATH=github.com/iotbzh/xds-server

EXT=
ifeq ($(HOST_GOOS), windows)
	EXT=.exe
endif

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
ROOT_SRCDIR := $(patsubst %/,%,$(dir $(mkfile_path)))
ROOT_GOPRJ := $(abspath $(ROOT_SRCDIR)/../../../..)
LOCAL_BINDIR := $(ROOT_SRCDIR)/bin
LOCAL_TOOLSDIR := $(ROOT_SRCDIR)/tools
PACKAGE_DIR := $(ROOT_SRCDIR)/package


export GOPATH := $(shell go env GOPATH):$(ROOT_GOPRJ)
export PATH := $(PATH):$(LOCAL_TOOLSDIR)

VERBOSE_1 := -v
VERBOSE_2 := -v -x

# Release or Debug mode
ifeq ($(filter 1,$(RELEASE) $(REL)),)
	GORELEASE=
	BUILD_MODE="Debug mode"
else
	# optimized code without debug info
	GORELEASE= -s -w
	BUILD_MODE="Release mode"
endif

ifeq ($(SUB_VERSION), )
	PACKAGE_ZIPFILE := xds-server_$(ARCH)-v$(VERSION).zip
else
	PACKAGE_ZIPFILE := xds-server_$(ARCH)-v$(VERSION)_$(SUB_VERSION).zip
endif


all: tools/syncthing build

.PHONY: build
build: xds webapp

xds:vendor scripts tools/syncthing/copytobin
	@echo "### Build XDS server (version $(VERSION), subversion $(SUB_VERSION))";
	@cd $(ROOT_SRCDIR); $(BUILD_ENV_FLAGS) go build $(VERBOSE_$(V)) -i -o $(LOCAL_BINDIR)/xds-server$(EXT) -ldflags "$(GORELEASE) -X main.AppVersion=$(VERSION) -X main.AppSubVersion=$(SUB_VERSION)" .

test: tools/glide
	go test --race $(shell $(LOCAL_TOOLSDIR)/glide novendor)

vet: tools/glide
	go vet $(shell $(LOCAL_TOOLSDIR)/glide novendor)

fmt: tools/glide
	go fmt $(shell $(LOCAL_TOOLSDIR)/glide novendor)

run: build/xds tools/syncthing/copytobin
	$(LOCAL_BINDIR)/xds-server$(EXT) --log info -c config.json.in

debug: build/xds webapp/debug tools/syncthing/copytobin
	$(LOCAL_BINDIR)/xds-server$(EXT) --log debug -c config.json.in

.PHONY: clean
clean:
	rm -rf $(LOCAL_BINDIR)/* debug $(ROOT_GOPRJ)/pkg/*/$(REPOPATH) $(PACKAGE_DIR)

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
	@mkdir -p $(LOCAL_BINDIR) && cp -rf scripts/xds-server-st*.sh scripts/xds-utils $(LOCAL_BINDIR)

.PHONY: conffile
conffile:
	cat config.json.in \
		| sed -e s,"webapp/dist","$(INSTALL_WEBAPP_DIR)",g \
		| sed -e s,"\./bin","",g \
		 > $(PACKAGE_DIR)/xds-server/config.json

.PHONY: install
install:
	@test -e $(LOCAL_BINDIR)/xds-server$(EXT) -a -d webapp/dist || { echo "Please execute first: make all\n"; exit 1; }
	@test -e $(LOCAL_BINDIR)/xds-server-start.sh -a -d $(LOCAL_BINDIR)/xds-utils || { echo "Please execute first: make all\n"; exit 1; }
	@test -e $(LOCAL_BINDIR)/syncthing$(EXT) -a -e $(LOCAL_BINDIR)/syncthing-inotify$(EXT) || { echo "Please execute first: make all\n"; exit 1; }
	mkdir -p $(INSTALL_DIR) \
		&& cp -a $(LOCAL_BINDIR)/* $(INSTALL_DIR)
	mkdir -p $(INSTALL_WEBAPP_DIR) \
		&& cp -a webapp/dist/* $(INSTALL_WEBAPP_DIR)

.PHONY: package
package: clean
	INSTALL_DIR=$(PACKAGE_DIR)/xds-server make -f $(ROOT_SRCDIR)/Makefile all install
	INSTALL_DIR=$(PACKAGE_DIR)/xds-server INSTALL_WEBAPP_DIR=www-xds-server	make -f $(ROOT_SRCDIR)/Makefile conffile
	(cd $(PACKAGE_DIR) && zip -r $(ROOT_SRCDIR)/$(PACKAGE_ZIPFILE) ./xds-server)

.PHONY: package-all
package-all:
	@echo "# Build linux amd64..."
	GOOS=linux GOARCH=amd64 RELEASE=1 make -f $(ROOT_SRCDIR)/Makefile package
	@echo "# Build windows amd64..."
	GOOS=windows GOARCH=amd64 RELEASE=1 make -f $(ROOT_SRCDIR)/Makefile package

vendor: tools/glide glide.yaml
	$(LOCAL_TOOLSDIR)/glide install --strip-vendor

tools/glide:
	@echo "Downloading glide"
	mkdir -p $(LOCAL_TOOLSDIR)
	curl --silent -L https://glide.sh/get | GOBIN=$(LOCAL_TOOLSDIR)  sh

.PHONY: tools/syncthing
tools/syncthing:
	@test -e $(LOCAL_TOOLSDIR)/syncthing$(EXT) -a -e $(LOCAL_TOOLSDIR)/syncthing-inotify$(EXT)  || { \
	mkdir -p $(LOCAL_TOOLSDIR); \
	DESTDIR=$(LOCAL_TOOLSDIR) \
	SYNCTHING_VERSION=$(SYNCTHING_VERSION) \
	SYNCTHING_INOTIFY_VERSION=$(SYNCTHING_INOTIFY_VERSION) \
	./scripts/xds-utils/get-syncthing.sh; }

.PHONY:
tools/syncthing/copytobin:
	@test -e $(LOCAL_TOOLSDIR)/syncthing$(EXT) -a -e $(LOCAL_TOOLSDIR)/syncthing-inotify$(EXT) || { echo "Please execute first: make tools/syncthing\n"; exit 1; }
	@cp -f $(LOCAL_TOOLSDIR)/syncthing$(EXT) $(LOCAL_TOOLSDIR)/syncthing-inotify$(EXT) $(LOCAL_BINDIR)

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
