#!/bin/bash

##########################################
# WARNING WARNING WARNING WARNING
#
# This script is an example to start a new AGL XDS container
#
# You should customize it to fit your environment and in particular
# adjust the paths and permissions where needed.
#
# Note that sharing volumes with host system is not mandatory: it
# was just added for performances reasons: building from a SSD is
# just faster than using the container filesystem: that's why /xdt is
# mounted from there. Same applies to ~/mirror and ~/share, which are
# just 2 convenient folders to store reference build caches (used in prepare_meta script)
#
##########################################

CURDIR=$(cd $(dirname $0) && pwd -P)

REGISTRY=docker.automotivelinux.org
REPO=agl
NAME=worker
FLAVOUR=xds
VERSION=3.99.1

# ---------------------------------------------------
# --- computed - don't touch !
# ---------------------------------------------------
DOCKER_USER=devel

DEFIMAGE=$REGISTRY/$REPO/$NAME-$FLAVOUR:$VERSION

function usage() {
	echo "Usage: $(basename $0) <instance ID> [image name]"  >&2
	echo "Instance ID must be 0 or a positive integer (1,2,...)" >&2
	echo "Image name is optional: 'make show-image' is used by default to get image" >&2
	echo "Default image: $DEFIMAGE" >&2
	exit 1
}

ID=""
IMAGE=$DEFIMAGE
FORCE=false
while [ $# -ne 0 ]; do
    case $1 in
        -h|--help|"")
            usage
            ;;
        -fr|-force-restart)
            FORCE=true
            ;;
        *)
            if [[ "$1" =~ ^[0-9]+$ ]]; then
                ID=$1
            else
                IMAGE=$1
            fi
            ;;
    esac
    shift
done

[ "$ID" = "" ] && usage

USER=$(id -un)
echo "Using instance ID #$ID (user $(id -un))"

NAME=agl-xds-$(hostname|cut -f1 -d'.')-$ID-$USER

MIRRORDIR=$HOME/ssd/localmirror_$ID
XDTDIR=$HOME/ssd/xdt_$ID
SHAREDDIR=$HOME/$DOCKER_USER/docker/share

SSH_PORT=$((2222 + ID))
WWW_PORT=$((8000 + ID))
BOOT_PORT=$((69 + ID))
NBD_PORT=$((10809 + ID))

mkdir -p $MIRRORDIR $XDTDIR $SHAREDDIR || exit 1
docker run \
	--publish=${SSH_PORT}:22 \
	--publish=${WWW_PORT}:8000 \
	--publish=${BOOT_PORT}:69/udp \
	--publish=${NBD_PORT}:10809 \
	--detach=true \
	--hostname=$NAME --name=$NAME \
	--privileged -v /sys/fs/cgroup:/sys/fs/cgroup:ro \
	-v $MIRRORDIR:/home/$DOCKER_USER/mirror \
	-v $SHAREDDIR:/home/$DOCKER_USER/share \
	-v $XDTDIR:/xdt \
	-it $IMAGE
if [ "$?" != "0" ]; then
    echo "An error was encountered while creating docker container."
    exit 1
fi

if ($FORCE); then
    echo "Stoping xds-server..."
    docker exec --user $DOCKER_USER  ${NAME} bash -c "/usr/local/bin/xds-server-stop.sh" || exit 1
    sleep 1
    echo "Starting xds-server..."
    docker exec --user $DOCKER_USER  ${NAME} bash -c "nohup /usr/local/bin/xds-server-start.sh" || exit 1
fi

echo "You can now login using:"
echo "   ssh -p $SSH_PORT $DOCKER_USER@$(hostname)"
