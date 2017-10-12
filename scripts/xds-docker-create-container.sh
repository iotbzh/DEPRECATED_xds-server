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
VERSION=4.0

# ---------------------------------------------------
# --- computed - don't touch !
# ---------------------------------------------------
DOCKER_USER=devel

DEFIMAGE=$REGISTRY/$REPO/$NAME-$FLAVOUR:$VERSION

function usage() {
	echo "Usage: $(basename $0) [-h] [-fr] [-v] <instance ID> [image name]"  >&2
	echo "Instance ID must be 0 or a positive integer (1,2,...)" >&2
	echo "Image name is optional: 'make show-image' is used by default to get image" >&2
	echo "Default image: $DEFIMAGE" >&2
    echo ""
    echo "Options:"
    echo " -fr | --force-restart   Force restart of xds-server service"
    echo " -nuu | --no-uid-update  Don't update user/group id within docker"
    echo " -v | --volume           Additional docker volume to bind, syntax is -v /InDockerPath:/HostPath "
	exit 1
}

ID=""
IMAGE=""
FORCE_RESTART=false
UPDATE_UID=true
USER_VOLUME_OPTION=""
while [ $# -ne 0 ]; do
    case $1 in
        -h|--help|"")
            usage
            ;;
        -fr|--force-restart)
            FORCE_RESTART=true
            ;;
        -nuu|--no-uid-update)
            UPDATE_UID=false
            ;;
        -v|--volume)
            shift
            if [[ "$1" =~ .*:.* ]]; then
                USER_VOLUME_OPTION="-v $1"
            else
                echo "Invalid volume option, format must be /InDockerPath:/hostPath"
                exit 1
            fi
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

[ "$ID" = "" ] && ID=0

# Dynamically retrieve image name
if [ "$IMAGE" = "" ]; then

    VER_NUM=`docker images $REGISTRY/$REPO/$NAME-$FLAVOUR:* --format {{.Tag}} | wc -l`
    if [ $VER_NUM -gt 1 ]; then
        echo "ERROR: more than one xds image found, please set explicitly the image to use !"
        exit 1
    elif [ $VER_NUM -lt 1 ]; then
        echo "ERROR: cannot automatically retrieve image tag for $REGISTRY/$REPO/$NAME-$FLAVOUR"
        exit 1
    fi

    VERSION=`docker images $REGISTRY/$REPO/$NAME-$FLAVOUR:* --format {{.Tag}}`
    if [ "$VERSION" = "" ]; then
        echo "ERROR: cannot automatically retrieve image tag for $REGISTRY/$REPO/$NAME-$FLAVOUR"
        usage
        exit 1
    fi

    IMAGE=$REGISTRY/$REPO/$NAME-$FLAVOUR:$VERSION
fi

USER=$(id -un)
echo "Using instance ID #$ID (user $(id -un))"

NAME=agl-xds-$(hostname|cut -f1 -d'.')-$ID-$USER

docker ps -a |grep "$NAME" > /dev/null
[ "$?" = "0" ] && { echo "Image name already exist ! (use -h option to read help)"; exit 1; }

XDS_WKS=$HOME/xds-workspace
XDTDIR=$XDS_WKS/.xdt_$ID

SSH_PORT=$((2222 + ID))
WWW_PORT=$((8000 + ID))
BOOT_PORT=$((69 + ID))
NBD_PORT=$((10809 + ID))

### Create the new container
mkdir -p $XDS_WKS $XDTDIR  || exit 1
docker run \
	--publish=${SSH_PORT}:22 \
	--publish=${WWW_PORT}:8000 \
	--publish=${BOOT_PORT}:69/udp \
	--publish=${NBD_PORT}:10809 \
	--detach=true \
	--hostname=$NAME --name=$NAME \
	--privileged -v /sys/fs/cgroup:/sys/fs/cgroup:ro \
	-v $XDS_WKS:/home/$DOCKER_USER/xds-workspace \
	-v $XDTDIR:/xdt \
    $USER_VOLUME_OPTION \
	-it $IMAGE
if [ "$?" != "0" ]; then
    echo "An error was encountered while creating docker container."
    exit 1
fi

### Ssh key
echo "Copying your identity to container $NAME"
echo -n wait ssh service .
res=3
max=30
count=0
while [ $res -ne 0 ] && [ $count -le $max ]; do
    sleep 1
    docker exec ${NAME} bash -c "systemctl status ssh" 2>/dev/null 1>&2
    res=$?
    echo -n "."
    count=$(expr $count + 1);
done
echo

ssh-keygen -R [localhost]:$SSH_PORT -f ~/.ssh/known_hosts
docker exec ${NAME} bash -c "mkdir -p /home/$DOCKER_USER/.ssh"
docker cp ~/.ssh/id_rsa.pub ${NAME}:/home/$DOCKER_USER/.ssh/authorized_keys
docker exec ${NAME} bash -c "chown $DOCKER_USER:$DOCKER_USER -R /home/$DOCKER_USER/.ssh ;chmod 0700 /home/$DOCKER_USER/.ssh; chmod 0600 /home/$DOCKER_USER/.ssh/*"
ssh -o StrictHostKeyChecking=no -p $SSH_PORT $DOCKER_USER@localhost exit

echo "You can now login using:"
echo "   ssh -p $SSH_PORT $DOCKER_USER@localhost"

### User / Group id
if ($UPDATE_UID); then
    echo "Setup docker user and group id to match yours"
    docker exec -t ${NAME} bash -c "systemctl stop xds-server" || exit 1
    docker exec -t ${NAME} bash -c "usermod -u $(id -u) $DOCKER_USER && groupmod -g $(id -g) $DOCKER_USER" || exit 1
    docker exec -t ${NAME} bash -c "chown -R $DOCKER_USER:$DOCKER_USER /home/$DOCKER_USER /tmp/xds*" || exit 1
    docker exec -t ${NAME} bash -c "systemctl start xds-server" || exit 1
    docker exec -t ${NAME} bash -c "systemctl start xds-server" || exit 1
fi

### Force xds-server restart
if ($FORCE_RESTART); then
    echo "Stopping xds-server..."
    docker exec -t ${NAME} bash -c "systemctl stop xds-server" || exit 1
    sleep 1
    echo "Starting xds-server..."
    docker exec -t ${NAME} bash -c "systemctl start xds-server" || exit 1
fi

echo "Done, docker container $NAME is ready to be used."
