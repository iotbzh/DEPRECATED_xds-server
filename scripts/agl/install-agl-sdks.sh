#!/bin/bash

. /etc/xdtrc

[ -z "$SDK_BASEURL" ] && SDK_BASEURL="http://iot.bzh/download/public/2017/XDS/sdk/"
[ -z "$XDT_SDK" ] && XDT_SDK=/xdt/sdk

# Support only poky_agl profile for now
PROFILE="poky-agl"

usage() {
    echo "Usage: $(basename $0) [-h|--help] [-noclean] -a|--arch <arch name>"
	echo "Sdk arch name is: aarch64 or arm32 or x86-64"
	exit 1
}

do_cleanup=true
FILE=""
ARCH=""
while [ $# -ne 0 ]; do
    case $1 in
        -h|--help|"")
            usage
            ;;
        -a|--arch)
            shift
            ARCH=$1
            case $1 in
                aarch64)    FILE="poky-agl-glibc-x86_64-agl-demo-platform-crosssdk-aarch64-toolchain-3.90.0+snapshot.sh";;
                arm32)      FILE="poky-agl-glibc-x86_64-agl-demo-platform-crosssdk-armv7vehf-neon-vfpv4-toolchain-3.90.0+snapshot.sh";;
                x86-64)     FILE="poky-agl-glibc-x86_64-agl-demo-platform-crosssdk-corei7-64-toolchain-3.90.0+snapshot.sh";;
            esac
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

if [ "$FILE" = "" ]; then
    echo "Option -a|--arch must be set"
    usage
fi
if [ "$ARCH" = "" ]; then
    echo "Unsupport architecture name !"
    usage
fi

cd ${XDT_SDK} || exit 1

# Cleanup
trap "cleanExit" 0 1 2 15
cleanExit ()
{
    if ($do_cleanup); then
        rm -f ${XDT_SDK}/${FILE}
    fi
}

# Get SDK installer
if [ ! -f $FILE ]; then
    wget "$SDK_BASEURL/$FILE" -O ${XDT_SDK}/${FILE} || exit 1
fi

# Retreive default install dir to extract version
offset=$(grep -na -m1 "^MARKER:$" $FILE | cut -d':' -f1)
eval $(head -n $offset $FILE | grep ^DEFAULT_INSTALL_DIR= )
VERSION=$(basename $DEFAULT_INSTALL_DIR)

DESTDIR=${XDT_SDK}/${PROFILE}/${VERSION}/${ARCH}

# Cleanup previous install
rm -rf ${DESTDIR} && mkdir -p ${DESTDIR} || exit 1

# Install sdk
chmod +x ${XDT_SDK}/${FILE}
${XDT_SDK}/${FILE}  -y -d ${DESTDIR}
