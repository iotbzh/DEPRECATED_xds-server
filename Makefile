# Makefile used to build XDS daemon Web Server

# Retrieve git tag/commit to set sub-version string
ifeq ($(origin VERSION), undefined)
	VERSION := $(shell git describe --tags --always | sed 's/^v//')
	ifeq ($(VERSION), )
		VERSION=unknown-dev
	endif
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
	rm -rf bin tools glide.lock vendor cmd/*/vendor webapp/{node_modules,dist}

run3:
	goreman start

webapp: webapp/install
	(cd webapp && gulp build)

webapp/debug:
	(cd webapp && gulp watch &)

webapp/install:
	(cd webapp && npm install)


# FIXME - package webapp
release: releasetar
	goxc -d ./release -tasks-=go-vet,go-test -os="linux darwin" -pv=$(VERSION)  -arch="386 amd64 arm arm64" -build-ldflags="-X main.AppVersionGitTag=$(VERSION)" -resources-include="README.md,Documentation,LICENSE,contrib" -main-dirs-exclude="vendor"

releasetar:
	mkdir -p release/$(VERSION)
	glide install --strip-vcs --strip-vendor --update-vendored --delete
	glide-vc --only-code --no-tests --keep="**/*.json.in"
	git ls-files > /tmp/xds-server-build
	find vendor >> /tmp/xds-server-build
	find webapp/ -path webapp/node_modules -prune -o -print >> /tmp/xds-server-build
	tar -cvf release/$(VERSION)/xds-server_$(VERSION)_src.tar -T /tmp/xds-server-build --transform 's,^,xds-server_$(VERSION)/,'
	rm /tmp/xds-server-build
	gzip release/$(VERSION)/xds-server_$(VERSION)_src.tar


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
