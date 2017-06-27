#!/bin/bash

# Configurable variables
[ -z "$BINDIR" ] && BINDIR=/usr/local/bin
[ -z "$XDS_CONFFILE" ] && XDS_CONFFILE=$HOME/.xds/config.json
[ -z "$XDS_SHAREDIR" ] && XDS_SHAREDIR=$HOME/.xds/share
[ -z "$ST_CONFDIR" ] && ST_CONFDIR=$HOME/.xds/syncthing-config
[ -z "$XDS_WWWDIR" ] && XDS_WWWDIR=webapp/dist
[ -z "$LOGLEVEL" ] && LOGLEVEL=info
[ -z "$LOGDIR" ] && LOGDIR=/tmp/xds-server/logs
[ -z "PORT_SRV" ] && PORT_SRV=8000
[ -z "$PORT_GUI" ] && PORT_GUI=8384
[ -z "$API_KEY" ] && API_KEY="1234abcezam"
[ -z "$UPDATE_XDS_TARBALL" ] && UPDATE_XDS_TARBALL=1

[[ -f $BINDIR/xds-server ]] || { echo "Cannot find xds-server in BINDIR !"; exit 1; }

# Create config.json file when needed
if [ ! -f "${XDS_CONFFILE}" ]; then
    mv ${XDS_CONFFILE} ${XDS_CONFFILE}.old
    [ ! -f "$XDS_WWWDIR/index.html" ] && XDS_WWWDIR=$BINDIR/www-xds-server
    [ ! -f "$XDS_WWWDIR/index.html" ] && XDS_WWWDIR=/var/www/xds-server
    [ ! -f "$XDS_WWWDIR/index.html" ] && { echo "Cannot determine XDS-server webapp directory."; exit 1; }
    cat <<EOF > ${XDS_CONFFILE}
{
    "HTTPPort": ${PORT_SRV},
    "webAppDir": "${XDS_WWWDIR}",
    "shareRootDir": "${XDS_SHAREDIR}",
    "logsDir": "${LOGDIR}",
    "sdkRootDir": "/xdt/sdk",
    "syncthing": {
        "binDir": "${BINDIR}",
        "home": "${ST_CONFDIR}",
        "gui-address": "http://localhost:${PORT_GUI}",
        "gui-apikey": "${API_KEY}"
    }
}
EOF
fi

echo "### Configuration in config.json: "
cat ${XDS_CONFFILE}
echo ""

mkdir -p ${LOGDIR}
LOG_XDS=${LOGDIR}/xds-server.log

# Download xds-agent tarball
if [ "${UPDATE_XDS_TARBALL}" = 1 ]; then
    SCRIPT_GET_XDS_TARBALL=$BINDIR/xds-utils/get-xds-agent.sh
    if [ ! -f ${SCRIPT_GET_XDS_TARBALL} ]; then
        SCRIPT_GET_XDS_TARBALL=$(dirname $0)/xds-utils/get-xds-agent.sh
    fi
    if [ -f ${SCRIPT_GET_XDS_TARBALL} ]; then
        TARBALLDIR=${XDS_WWWDIR}/assets/xds-agent-tarballs
        [ ! -d "$TARBALLDIR" ] && TARBALLDIR=$BINDIR/www-xds-server/assets/xds-agent-tarballs
        [ ! -d "$TARBALLDIR" ] && TARBALLDIR=$(grep webAppDir ~/.xds/config.json|cut -d '"' -f 4)/assets/xds-agent-tarballs
        if [ -d "$TARBALLDIR" ]; then
            DEST_DIR=$TARBALLDIR $SCRIPT_GET_XDS_TARBALL
        else
            echo "WARNING: cannot download / update xds-agent tarballs (DESTDIR error)"
        fi
    else
        echo "WARNING: cannot download / update xds-agent tarballs"
    fi
fi


echo "### Start XDS server"
echo "nohup $BINDIR/xds-server --config $XDS_CONFFILE -log $LOGLEVEL > $LOG_XDS 2>&1"
if [ "$1" != "-dryrun" ]; then
    nohup $BINDIR/xds-server --config $XDS_CONFFILE -log $LOGLEVEL > $LOG_XDS 2>&1 &
    pid_xds=$(jobs -p)
    echo "pid=${pid_xds}"
fi
