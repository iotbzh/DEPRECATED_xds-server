#!/bin/bash

# Configurable variables
[ -z "$BINDIR" ] && BINDIR=/usr/local/bin
[ -z "$XDS_CONFFILE" ] && XDS_CONFFILE=$HOME/.xds/config.json
[ -z "$XDS_SHAREDIR" ] && XDS_SHAREDIR=$HOME/.xds/share
[ -z "$ST_CONFDIR" ] && ST_CONFDIR=$HOME/.xds/syncthing-config
[ -z "$LOGLEVEL" ] && LOGLEVEL=warn
[ -z "$LOGDIR" ] && LOGDIR=/tmp/xds-server/logs
[ -z "$PORT_GUI" ] && PORT_GUI=8384
[ -z "$API_KEY" ] && API_KEY="1234abcezam"


[[ -f $BINDIR/xds-server ]] || { echo "Cannot find xds-server in BINDIR !"; exit 1; }

# Create config.json file when needed
[[ -f ${XDS_CONFFILE} ]] || { mv ${XDS_CONFFILE} ${XDS_CONFFILE}.old; }

cat <<EOF > ${XDS_CONFFILE}
{
    "webAppDir": "webapp/dist",
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

echo "### Configuration in config.json: "
cat ${XDS_CONFFILE}
echo ""

mkdir -p ${LOGDIR}
LOG_XDS=${LOGDIR}/xds-server.log

echo "### Start XDS server"
echo " $BINDIR/xds-server --config $XDS_CONFFILE -log $LOGLEVEL > $LOG_XDS 2>&1"
if [ "$1" != "-dryrun" ]; then
    $BINDIR/xds-server --config $XDS_CONFFILE -log $LOGLEVEL > $LOG_XDS 2>&1 &
    pid_xds=$(jobs -p)
    echo "pid=${pid_xds}"
fi
