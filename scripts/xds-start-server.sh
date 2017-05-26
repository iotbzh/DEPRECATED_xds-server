#!/bin/bash

# Configurable variables
[ -z "$BINDIR" ] && BINDIR=/usr/local/bin
[ -z "$XDS_CONFFILE" ] && XDS_CONFFILE=$HOME/.xds/config.json
[ -z "$XDS_SHAREDIR" ] && XDS_SHAREDIR=$HOME/.xds/share
[ -z "$ST_CONFDIR" ] && ST_CONFDIR=$HOME/.xds/syncthing-config
[ -z "$XDS_WWWDIR" ] && XDS_WWWDIR=webapp/dist
[ -z "$LOGLEVEL" ] && LOGLEVEL=warn
[ -z "$LOGDIR" ] && LOGDIR=/tmp/xds-server/logs
[ -z "$PORT_GUI" ] && PORT_GUI=8384
[ -z "$API_KEY" ] && API_KEY="1234abcezam"

[[ -f $BINDIR/xds-server ]] || { echo "Cannot find xds-server in BINDIR !"; exit 1; }

# Create config.json file when needed
if [ ! -f "${XDS_CONFFILE}" ]; then
    mv ${XDS_CONFFILE} ${XDS_CONFFILE}.old
    [ ! -f "$XDS_WWWDIR/index.html" ] && XDS_WWWDIR=$BINDIR/www-xds-server
    [ ! -f "$XDS_WWWDIR/index.html" ] && XDS_WWWDIR=/var/www/xds-server
    [ ! -f "$XDS_WWWDIR/index.html" ] && { echo "Cannot determine XDS-server webapp directory."; exit 1; }
    cat <<EOF > ${XDS_CONFFILE}
{
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

echo "### Start XDS server"
echo "nohup $BINDIR/xds-server --config $XDS_CONFFILE -log $LOGLEVEL > $LOG_XDS 2>&1"
if [ "$1" != "-dryrun" ]; then
    nohup $BINDIR/xds-server --config $XDS_CONFFILE -log $LOGLEVEL > $LOG_XDS 2>&1 &
    pid_xds=$(jobs -p)
    echo "pid=${pid_xds}"
fi
