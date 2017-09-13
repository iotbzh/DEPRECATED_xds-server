#!/bin/bash

# Configurable variables
[ -z "$SYNCTHING_VERSION" ] && SYNCTHING_VERSION=0.14.28
[ -z "$SYNCTHING_INOTIFY_VERSION" ] && SYNCTHING_INOTIFY_VERSION=0.8.6
# XXX - may be cleanup
# Used as temporary HACK while waiting merge of #165
#[ -z "$SYNCTHING_INOTIFY_VERSION" ] && { SYNCTHING_INOTIFY_VERSION=master; SYNCTHING_INOTIFY_CMID=af6fbf9d63f95a0; }
[ -z "$DESTDIR" ] && DESTDIR=/opt/AGL/xds/server
[ -z "$TMPDIR" ] && TMPDIR=/tmp
[ -z "$GOOS" ] && GOOS=$(go env GOOS)
[ -z "$GOARCH" ] && GOARCH=$(go env GOARCH)
[ -z "$CLEANUP" ] && CLEANUP=false


TEMPDIR=$TMPDIR/.get-syncthing.tmp
mkdir -p ${TEMPDIR} && cd ${TEMPDIR} || exit 1
trap "cleanExit" 0 1 2 15
cleanExit ()
{
    if [ "$CLEANUP" = "true" ]; then
        rm -rf ${TEMPDIR}
    fi
}

TB_EXT="tar.gz"
EXT=""
[[ "$GOOS" = "windows" ]] && { TB_EXT="zip"; EXT=".exe"; }

GOOS_ST=${GOOS}
GOOS_STI=${GOOS}
[[ "$GOOS" = "darwin" ]] && GOOS_ST="macosx"

echo "Get Syncthing..."



## Install Syncthing + Syncthing-inotify
## gpg: key 00654A3E: public key "Syncthing Release Management <release@syncthing.net>" imported
GPG=$(which gpg)
if [ "$?" != 0 ]; then
    echo "You must install first gpg ( eg.: sudo apt install gpg )"
    exit 1
fi

gpg -q --keyserver pool.sks-keyservers.net --recv-keys 37C84554E7E0A261E4F76E1ED26E6ED000654A3E || exit 1

tarball="syncthing-${GOOS_ST}-${GOARCH}-v${SYNCTHING_VERSION}.${TB_EXT}" \
	&& curl -sfSL "https://github.com/syncthing/syncthing/releases/download/v${SYNCTHING_VERSION}/${tarball}" -O \
    && curl -sfSL "https://github.com/syncthing/syncthing/releases/download/v${SYNCTHING_VERSION}/sha1sum.txt.asc" -O \
	&& gpg -q --verify sha1sum.txt.asc \
	&& grep -E " ${tarball}\$" sha1sum.txt.asc | sha1sum -c - \
	&& rm -rf sha1sum.txt.asc syncthing-${GOOS_ST}-${GOARCH}-v${SYNCTHING_VERSION}
	if [ "${TB_EXT}" = "tar.gz" ]; then
        tar -xvf "$tarball" --strip-components=1 "$(basename "$tarball" .tar.gz)"/syncthing \
        && mv syncthing ${DESTDIR}/syncthing || exit 1
    else
        unzip "$tarball" && mv syncthing-windows-*/syncthing.exe ${DESTDIR}/syncthing.exe || exit 1
    fi

echo "Get Syncthing-inotify..."
if [ "$SYNCTHING_INOTIFY_VERSION" = "master" ]; then
    mkdir -p ${TEMPDIR}/syncthing-inotify-build/src/github.com/syncthing || exit 1
    cd ${TEMPDIR}/syncthing-inotify-build/src/github.com/syncthing
    [[ ! -d ./syncthing ]] && (git clone https://github.com/syncthing/syncthing || exit 1; )
    [[ ! -d ./syncthing-inotify ]] && (git clone https://github.com/syncthing/syncthing-inotify || exit 1; )
    cd syncthing-inotify
    git status
    version=$(git describe --tags --always | sed 's/^v//')__patch_165
    if [ "$SYNCTHING_INOTIFY_CMID" != "" ]; then
        git checkout -q $SYNCTHING_INOTIFY_CMID || exit 1
        version=${version}__patch_165
    fi

    # Workaround about "cannot find package "golang.org/x/sys/unix"
    go get -u golang.org/x/sys/unix

    # Workaround about "undefined: stream" error when cross-building MacOS
    # https://github.com/rjeczalik/notify/issues/108
    OPTS=""
    [[ "$GOOS_STI" = "darwin" ]] && OPTS="-tags kqueue"

    export GOPATH=$(cd ../../../.. && pwd)
    go build -v -i -ldflags "-w -X main.Version=$version" -o ${DESTDIR}/syncthing-inotify${EXT} || exit 1
else

    tarball="syncthing-inotify-${GOOS_STI}-${GOARCH}-v${SYNCTHING_INOTIFY_VERSION}.${TB_EXT}"
    curl -sfSL "https://github.com/syncthing/syncthing-inotify/releases/download/v${SYNCTHING_INOTIFY_VERSION}/${tarball}" -O || exit 1
    rm -rf syncthing-inotify-${GOOS_STI}-${GOARCH}-v${SYNCTHING_INOTIFY_VERSION}
    if [ "${TB_EXT}" = "tar.gz" ]; then
        tar -xvf "${tarball}" syncthing-inotify && mv syncthing-inotify ${DESTDIR}/syncthing-inotify || exit 1
    else
        unzip "$tarball" && mv syncthing-inotify.exe ${DESTDIR}/syncthing-inotify.exe || exit 1
    fi
fi

echo "DONE: syncthing and syncthing-inotify successfuly installed in ${DESTDIR}"
