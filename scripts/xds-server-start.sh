#!/bin/bash

BINDIR=/opt/AGL/bin
[[ -f $BINDIR/xds-server ]] || BINDIR=$(which xds-server)
[[ -f $BINDIR/xds-server ]] || BINDIR=/usr/local/bin    ;# for backward compat
[[ -f $BINDIR/xds-server ]] || { echo "Cannot find xds-server executable !"; exit 1; }

echo "### Start XDS server"
nohup $BINDIR/xds-server $* &
exit $?
