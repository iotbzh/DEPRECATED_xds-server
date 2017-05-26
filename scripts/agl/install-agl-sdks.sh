#!/bin/bash

. /etc/xdtrc

[ -z "$SDK_BASEURL" ] && SDK_BASEURL="http://iot.bzh/download/public/2017/XDS/sdk/"
[ -z "$XDT_SDK" ] && XDT_SDK=/xdt/sdk

# Support only poky_agl profile for now
PROFILE="poky-agl"

SDKS=$(curl -s ${SDK_BASEURL} | grep -oP  'href="[^"]*.sh"' | cut -d '"' -f 2)

usage() {
    echo "Usage: $(basename $0) [-h|--help] [-noclean] [-a|--arch <arch name>] [-l|--list]"
	echo "For example, arch name is: aarch64, armv7vehf or x86-64"
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
            FILE=$(getFile $ARCH)
            if [ "$?" != 0 ]; then
                exit 1
            fi
            ;;
        -l|--list)
            echo "Available SDKs tarballs:"
            for sdk in $SDKS; do echo " $sdk"; done
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
        [[ -f ${XDT_SDK}/${FILE} ]] && rm -f ${XDT_SDK}/${FILE}
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
