#!/bin/bash

# Configurable variables
[ -z "$SYNCTHING_VERSION" ] && SYNCTHING_VERSION=0.14.28

# FIXME: temporary HACK while waiting merge of #165
#[ -z "$SYNCTHING_INOTIFY_VERSION" ] && SYNCTHING_INOTIFY_VERSION=0.8.5
[ -z "$SYNCTHING_INOTIFY_VERSION" ] && { SYNCTHING_INOTIFY_VERSION=master; SYNCTHING_INOTIFY_CMID=af6fbf9d63f95a0; }
[ -z "$DESTDIR" ] && DESTDIR=/usr/local/bin
[ -z "$TMPDIR" ] && TMPDIR=/tmp
[ -z "$GOOS" ] && GOOS=$(go env GOOS)
[ -z "$GOARCH" ] && GOARCH=$(go env GOARCH)


TEMPDIR=$TMPDIR/.get-st.$$
mkdir -p ${TEMPDIR} && cd ${TEMPDIR} || exit 1
trap "cleanExit" 0 1 2 15
cleanExit ()
{
   rm -rf ${TEMPDIR}
}

TB_EXT="tar.gz"
EXT=""
[[ "$GOOS" = "windows" ]] && { TB_EXT="zip"; EXT=".exe"; }

echo "Get Syncthing..."

## Install Syncthing + Syncthing-inotify
## gpg: key 00654A3E: public key "Syncthing Release Management <release@syncthing.net>" imported
gpg -q --keyserver pool.sks-keyservers.net --recv-keys 37C84554E7E0A261E4F76E1ED26E6ED000654A3E || exit 1

tarball="syncthing-${GOOS}-${GOARCH}-v${SYNCTHING_VERSION}.${TB_EXT}" \
	&& curl -sfSL "https://github.com/syncthing/syncthing/releases/download/v${SYNCTHING_VERSION}/${tarball}" -O \
    && curl -sfSL "https://github.com/syncthing/syncthing/releases/download/v${SYNCTHING_VERSION}/sha1sum.txt.asc" -O \
	&& gpg -q --verify sha1sum.txt.asc \
	&& grep -E " ${tarball}\$" sha1sum.txt.asc | sha1sum -c - \
	&& rm sha1sum.txt.asc
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
    git clone https://github.com/syncthing/syncthing || exit 1
    git clone https://github.com/syncthing/syncthing-inotify || exit 1
    cd syncthing-inotify
    if [ "$SYNCTHING_INOTIFY_CMID" != "" ]; then
        git checkout -q $SYNCTHING_INOTIFY_CMID || exit 1
    fi
    git status
    export GOPATH=$(realpath `pwd`/../../../..)
    version=$(git describe --tags --always | sed 's/^v//')__patch_165
    go build -v -i -ldflags "-w -X main.Version=$version" -o ${DESTDIR}/syncthing-inotify${EXT} || exit 1
else

    tarball="syncthing-inotify-${GOOS}-${GOARCH}-v${SYNCTHING_INOTIFY_VERSION}.${TB_EXT}"
    curl -sfSL "https://github.com/syncthing/syncthing-inotify/releases/download/v${SYNCTHING_INOTIFY_VERSION}/${tarball}" -O || exit 1
    if [ "${TB_EXT}" = "tar.gz" ]; then
        tar -xvf "${tarball}" syncthing-inotify && mv syncthing-inotify ${DESTDIR}/syncthing-inotify || exit 1
    else
        unzip "$tarball" && mv syncthing-inotify.exe ${DESTDIR}/syncthing-inotify.exe || exit 1
    fi
fi

echo "DONE: syncthing and syncthing-inotify successfuly installed in ${DESTDIR}"