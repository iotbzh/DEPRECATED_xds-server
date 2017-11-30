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

[ -z "$SDK_BASEURL" ] && SDK_BASEURL="http://iot.bzh/download/public/2017/XDS/sdk/"
[ -z "$XDT_SDK" ] && XDT_SDK=/xdt/sdk

# Support only poky_agl profile for now
PROFILE="poky-agl"

# Dynamically retreive SDKs (reduce timeout in case of iot website is not available)
SDKS=$(curl -s ${SDK_BASEURL} --connect-timeout 10 | grep -oP  'href="[^"]*.sh"' | cut -d '"' -f 2) || echo "WARNING: IoT.bzh download site not available."

usage() {
    echo "Usage: $(basename $0) [-h|--help] [-clean] [-f|--file <agl-sdk-filename>] [-a|--arch <arch name>] [-l|--list]"
	echo "For example, arch name is: aarch64, armv7vehf or corei7-64"
	exit 1
}

getFile() {
    arch=$1
    for sdk in ${SDKS}; do
        echo $sdk | grep "${PROFILE}.*${arch}.*.sh" > /dev/null 2>&1
        if [ "$?" = "0" ]; then
            echo $sdk
            return 0
        fi
    done
    echo "No SDK tarball found for arch $arch"
    return 1
}

do_cleanup=false
FILE=""
ARCH=""
while [ $# -ne 0 ]; do
    case $1 in
        -h|--help|"")
            usage
            ;;
        -f|--file)
            shift
            FILE=$1
            ;;
        -a|--arch)
            shift
            ARCH=$1
            ;;
        -l|--list)
            echo "Available SDKs tarballs:"
            for sdk in $SDKS; do echo " $sdk"; done
            exit 0
            ;;
        -clean)
            do_cleanup=true
            ;;
        *)
            echo "Invalid argument: $1"
            usage
            ;;
    esac
    shift
done

if [ "$ARCH" = "x86-64" ]; then
    echo "Warning x86-64 architure name is deprecated, please use corei7-64 instead !"
    exit 1
fi

[ ! -d ${XDT_SDK} ] && mkdir -p ${XDT_SDK}

if [ "$FILE" = "" ]; then
    FILE=$(getFile $ARCH)
    if [ "$?" != "0" ]; then
        echo "$FILE"
        exit 1
    fi
    SDK_FILE=${XDT_SDK}/${FILE}
elif [ ! -f "$FILE" ]; then
    echo "SDK file not found: $FILE"
    exit 1
else
    DIR=$(cd $(dirname $FILE); pwd)
    FILE=$(basename $FILE)
    SDK_FILE=${DIR}/${FILE}
fi

if [ "$ARCH" = "" ]; then
    echo "Option -a|--arch must be set"
    usage
fi

# Check that ARCH name matching SDK tarball filename
echo "$FILE" | grep "$ARCH" > /dev/null 2>&1
if [ "$?" = "1" ]; then
    echo "ARCH and provided filename doesn't match !"
    exit 1
fi

cd ${XDT_SDK} || exit 1

# Cleanup
trap "cleanExit" 0 1 2 15
cleanExit ()
{
    if ($do_cleanup); then
        [[ -f ${SDK_FILE} ]] && rm -f ${SDK_FILE}
    fi
}

# Get SDK installer
if [ ! -f "${SDK_FILE}" ]; then
    do_cleanup=true
    wget "$SDK_BASEURL/$FILE" -O "${SDK_FILE}" || exit 1
fi

# Retreive default install dir to extract version
offset=$(grep -na -m1 "^MARKER:$" "${SDK_FILE}" | cut -d':' -f1)
eval $(head -n $offset "${SDK_FILE}" | grep ^DEFAULT_INSTALL_DIR= )
VERSION=$(basename $DEFAULT_INSTALL_DIR)

[ "$PROFILE" = "" ] && { echo "PROFILE is not set"; exit 1; }
[ "$VERSION" = "" ] && { echo "VERSION is not set"; exit 1; }

DESTDIR=${XDT_SDK}/${PROFILE}/${VERSION}/${ARCH}

# Cleanup previous install
rm -rf ${DESTDIR} && mkdir -p ${DESTDIR} || exit 1

# Install sdk
chmod +x ${SDK_FILE}
${SDK_FILE}  -y -d ${DESTDIR}
