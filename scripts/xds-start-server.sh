#!/bin/bash

# Configurable variables
[ -z "$BINDIR" ] && BINDIR=/usr/local/bin
[ -z "$ST_CONFDIR" ] && ST_CONFDIR=$HOME/.xds/syncthing-config
[ -z "$XDS_CONFFILE" ] && XDS_CONFFILE=$HOME/.xds/config.json
[ -z "$LOGLEVEL" ] && LOGLEVEL=warn
[ -z "$LOGDIR" ] && LOGDIR=/tmp/xds-logs
[ -z "$PORT_GUI" ] && PORT_GUI=8384
[ -z "$API_KEY" ] && API_KEY="1234abcezam"


mkdir -p ${LOGDIR}
LOG_XDS=${LOGDIR}/xds-server.log
LOG_SYNC=${LOGDIR}/syncthing.log
LOG_SYNCI=${LOGDIR}/syncthing-inotify.log

echo "### Info"
echo "XDS server config: $XDS_CONFFILE"
echo "Syncthing GUI on port: $PORT_GUI"
echo "Syncthing Config: $ST_CONFDIR"
echo "XDS server output redirected in: $LOG_XDS"
echo "Syncthing-inotify output redirected in: $LOG_SYNCI"
echo "Syncthing output redirected in: $LOG_SYNC"
echo ""

pwd
[[ -f $BINDIR/xds-server ]] || { BINDIR=$(cd `dirname $0` && pwd); }
pwd
[[ -f $BINDIR/xds-server ]] || { echo "Cannot find xds-server in BINDIR !"; exit 1; }

echo "### Start syncthing-inotify:"
$BINDIR/syncthing-inotify --home=$ST_CONFDIR -target=http://localhost:$PORT_GUI -verbosity=4 > $LOG_SYNCI  2>&1 &
pid_synci=$(jobs -p)
echo "pid=${pid_synci}"
echo ""

echo "### Start Syncthing:"
STNODEFAULTFOLDER=1 $BINDIR/syncthing --home=$ST_CONFDIR -no-browser -verbose --gui-address=0.0.0.0:$PORT_GUI -gui-apikey=${API_KEY} > $LOG_SYNC 2>&1 &
pid_sync=$(jobs -p)
echo "pid=${pid_sync}"
echo ""

if [ "$1" == "-noserver" ]; then
    echo "## XDS server NOT STARTED"
    echo "  Command to start it:"
    echo "  $BINDIR/xds-server --config $XDS_CONFFILE -log $LOGLEVEL > $LOG_XDS 2>&1"
else
    # Wait a bit so make connection to Syncthing possible
    sleep 1
    echo "### Start XDS server"
    $BINDIR/xds-server --config $XDS_CONFFILE -log $LOGLEVEL > $LOG_XDS 2>&1 &
    pid_xds=$(jobs -p)
    echo "pid=${pid_xds}"
fi
echo ""
