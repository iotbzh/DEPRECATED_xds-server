#!/bin/bash

# Configurable variables
[ -z "$SYNCTHING_VERSION" ] && SYNCTHING_VERSION=0.14.25
[ -z "$SYNCTHING_INOTIFY_VERSION" ] && SYNCTHING_INOTIFY_VERSION=0.8.5
[ -z "$DESTDIR" ] && DESTDIR=/usr/local/bin
[ -z "$TMPDIR" ] && TMPDIR=/tmp


TEMPDIR=$TMPDIR/.get-st.$$
mkdir -p ${TEMPDIR} && cd ${TEMPDIR} || exit 1
trap "cleanExit" 0 1 2 15
cleanExit ()
{
   rm -rf ${TEMPDIR}
}

echo "Get Syncthing..."

## Install Syncthing + Syncthing-inotify
## gpg: key 00654A3E: public key "Syncthing Release Management <release@syncthing.net>" imported
gpg -q --keyserver pool.sks-keyservers.net --recv-keys 37C84554E7E0A261E4F76E1ED26E6ED000654A3E || exit 1

tarball="syncthing-linux-amd64-v${SYNCTHING_VERSION}.tar.gz" \
	&& curl -sfSL "https://github.com/syncthing/syncthing/releases/download/v${SYNCTHING_VERSION}/${tarball}" -O \
    && curl -sfSL "https://github.com/syncthing/syncthing/releases/download/v${SYNCTHING_VERSION}/sha1sum.txt.asc" -O \
	&& gpg -q --verify sha1sum.txt.asc \
	&& grep -E " ${tarball}\$" sha1sum.txt.asc | sha1sum -c - \
	&& rm sha1sum.txt.asc \
	&& tar -xvf "$tarball" --strip-components=1 "$(basename "$tarball" .tar.gz)"/syncthing \
	&& mv syncthing ${DESTDIR}/syncthing


echo "Get Syncthing-inotify..."
tarball="syncthing-inotify-linux-amd64-v${SYNCTHING_INOTIFY_VERSION}.tar.gz" \
    && curl -sfSL "https://github.com/syncthing/syncthing-inotify/releases/download/v${SYNCTHING_INOTIFY_VERSION}/${tarball}" -O \
    && tar -xvf "${tarball}" syncthing-inotify \
	&& mv syncthing-inotify ${DESTDIR}/syncthing-inotify

echo "DONE: syncthing and syncthing-inotify successfuly installed in ${DESTDIR}"