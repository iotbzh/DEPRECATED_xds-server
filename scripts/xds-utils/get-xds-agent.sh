#!/bin/bash
 ###########################################################################
# Copyright 2017 IoT.bzh
#
# author: Sebastien Douheret <sebastien@iot.bzh>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
###########################################################################

. /etc/xdtrc

[ -z "$XDS_AGENT_BASEURL" ] && XDS_AGENT_BASEURL="http://iot.bzh/download/public/2017/XDS/xds-agent/"
[ -z "$DEST_DIR" ] && DEST_DIR=./webapp/dist/assets/xds-agent-tarballs

# Fisrt check if we can access to iot.bzh (aka ovh.iot)
ping -c 1 -W 2 www.ovh.iot > /dev/null
if [ "$?" != "0" ]; then
    echo "iot.bzh website not accessible !"
    exit 1
fi

TARBALLS=$(curl -s ${XDS_AGENT_BASEURL} | grep -oP  'href="[^"]*.zip"' | cut -d '"' -f 2)

usage() {
    echo "Usage: $(basename $0) [-h|--help] [-noclean] [-a|--arch <arch name>] [-l|--list]"
	exit 1
}

do_cleanup=true
while [ $# -ne 0 ]; do
    case $1 in
        -h|--help|"")
            usage
            ;;
        -l|--list)
            echo "Available xds-agent tarballs:"
            for t in $TARBALLS; do echo " $t"; done
            exit 0
            ;;
        -noclean)
            do_cleanup=false
            ;;
        *)
            echo "Invalid argument: $1"
            usage
            ;;
    esac
    shift
done

if [ ! -d ${DEST_DIR} ]; then
    echo "Invalid destination directory: ${DEST_DIR}"
    exit 1
fi

# Get not existing tarballs
exitCode=0
for file in $TARBALLS; do
    DESTFILE=${DEST_DIR}/${file}
    if [ ! -f $DESTFILE ]; then
        echo -n " Downloading ${file}... "
        wget -q "${XDS_AGENT_BASEURL}/${file}" -O ${DESTFILE}
        if [ "$?" != 0 ]; then
            echo "ERROR"
            exitCode=1
        fi
        echo "OK"
    fi
done

exit $exitCode
