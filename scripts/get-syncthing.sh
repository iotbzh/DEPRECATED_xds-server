#!/bin/bash

# Configurable variables
[ -z "$SYNCTHING_VERSION" ] && SYNCTHING_VERSION=0.14.25

# FIXME: temporary HACK while waiting merge of #165
#[ -z "$SYNCTHING_INOTIFY_VERSION" ] && SYNCTHING_INOTIFY_VERSION=0.8.5
[ -z "$SYNCTHING_INOTIFY_VERSION" ] && SYNCTHING_INOTIFY_VERSION=master_and_patch165
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
if [ "$SYNCTHING_INOTIFY_VERSION" = "master_and_patch165" ]; then
    mkdir -p ${TEMPDIR}/syncthing-inotify-build/src/github.com/syncthing || exit 1
    cd ${TEMPDIR}/syncthing-inotify-build/src/github.com/syncthing
    git clone https://github.com/syncthing/syncthing || exit 1
    git clone https://github.com/syncthing/syncthing-inotify || exit 1
    cd syncthing-inotify
    cat <<EOF > 165.patch
    diff --git a/syncwatcher.go b/syncwatcher.go
index c36b034..5175c12 100644
--- a/syncwatcher.go
+++ b/syncwatcher.go
@@ -677,7 +677,10 @@ func accumulateChanges(debounceTimeout time.Duration,
 		if flushTimerNeedsReset {
 			flushTimerNeedsReset = false
 			if !flushTimer.Stop() {
-				<-flushTimer.C
+				select {
+				case <-flushTimer.C:
+				default:
+				}
 			}
 			flushTimer.Reset(currInterval)
 		}
EOF
    git apply 165.patch || exit 1
    export GOPATH=$(realpath `pwd`/../../../..)
    version=$(git describe --tags --always | sed 's/^v//')__patch_165
    go build -v -i -ldflags "-w -X main.Version=$version" -o ${DESTDIR}/syncthing-inotify || exit 1
else

tarball="syncthing-inotify-linux-amd64-v${SYNCTHING_INOTIFY_VERSION}.tar.gz" \
    && curl -sfSL "https://github.com/syncthing/syncthing-inotify/releases/download/v${SYNCTHING_INOTIFY_VERSION}/${tarball}" -O \
    && tar -xvf "${tarball}" syncthing-inotify \
	&& mv syncthing-inotify ${DESTDIR}/syncthing-inotify
fi

echo "DONE: syncthing and syncthing-inotify successfuly installed in ${DESTDIR}"