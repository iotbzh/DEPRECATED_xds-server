#!/bin/bash
# shellcheck disable=SC2086

##########################################
# WARNING WARNING WARNING WARNING
#
# This script is an example to start a new AGL XDS container
#
# You should customize it to fit your environment and in particular
# adjust the paths and permissions where needed.
#
##########################################

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
	echo "Usage: $(basename $0) [-h|--help] [-fr] [-id <instance container ID>] "
    echo "          [-nc] [-nuu] [-v|--volume <inpath:outpath>] [image name]"
	echo "Image name is optional; 'docker images' is used by default to get image"
	echo "Default image:"
    echo " $DEFIMAGE"
    echo ""
    echo "Options:"
    echo " -fr | --force-restart   Force restart of xds-server service"
    echo " -id                     Instance ID used to build container name, a positive integer (0,1,2,...)"
    echo " -nuu | --no-uid-update  Don't update user/group id within docker"
    echo " -v | --volume           Additional docker volume to bind, syntax is -v /InDockerPath:/HostPath "
	exit 1
}

ID=""
IMAGE=""
FORCE_RESTART=false
UPDATE_UID=true
USER_VOLUME_OPTION=""
NO_CLEANUP=false
while [ $# -ne 0 ]; do
    case $1 in
        -h|--help|"")
            usage
            ;;
        -fr|--force-restart)
            FORCE_RESTART=true
            ;;
        -nc|--no-cleanup)
            NO_CLEANUP=true
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
        -id)
            shift
            ID=$1
            ;;
        *)
            if [[ "$1" =~ ^[\.0-9]+$ ]]; then
                IMAGE=$REGISTRY/$REPO/$NAME-$FLAVOUR:$1
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

    VERSION_LIST=$(docker images $REGISTRY/$REPO/$NAME-$FLAVOUR --format '{{.Tag}}')
    VER_NUM=$(echo "$VERSION_LIST" | wc -l)
    if [ "$VER_NUM" -gt 1 ]; then
        echo "ERROR: more than one xds image found, please set explicitly the image to use !"
        echo "List of found images:"
        echo "$VERSION_LIST"
        exit 1
    elif [ "$VER_NUM" -lt 1 ]; then
        echo "ERROR: cannot automatically retrieve image tag for $REGISTRY/$REPO/$NAME-$FLAVOUR"
        exit 1
    fi
    if [ "$VERSION_LIST" = "" ]; then
        echo "ERROR: cannot automatically retrieve image tag for $REGISTRY/$REPO/$NAME-$FLAVOUR"
        usage
        exit 1
    fi

    IMAGE=$REGISTRY/$REPO/$NAME-$FLAVOUR:$VERSION_LIST
fi

USER=$(id -un)
echo "Using instance ID #$ID (user $(id -un))"

NAME=agl-xds-$(hostname|cut -f1 -d'.')-$ID-$USER

if docker ps -a |grep "$NAME" > /dev/null; then
    echo "Image name already exist ! (use -h option to read help)"
    exit 1
fi

XDS_WKS=$HOME/xds-workspace
XDTDIR=$XDS_WKS/.xdt_$ID

SSH_PORT=$((2222 + ID))
WWW_PORT=$((8000 + ID))
BOOT_PORT=$((69 + ID))
NBD_PORT=$((10809 + ID))

# Delete container on error
creation_done=false
trap "cleanExit" 0 1 2 15
cleanExit ()
{
    if [ "$creation_done" != "true" ] && [ "$NO_CLEANUP" != "true" ]; then
        echo "Error detected, remove unusable docker image ${NAME}"
        docker rm -f "${NAME}" > /dev/null 2>&1
    fi
}

### Create the new container
mkdir -p $XDS_WKS $XDTDIR  || exit 1
if ! docker run \
	--publish=${SSH_PORT}:22 \
	--publish=${WWW_PORT}:8000 \
	--publish=${BOOT_PORT}:69/udp \
	--publish=${NBD_PORT}:10809 \
	--detach=true \
	--hostname="$NAME" --name="$NAME" \
	--privileged -v /sys/fs/cgroup:/sys/fs/cgroup:ro \
	-v $XDS_WKS:/home/$DOCKER_USER/xds-workspace \
	-v $XDTDIR:/xdt \
    $USER_VOLUME_OPTION \
	-it $IMAGE;
then
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
    count=$((count + 1));
done
echo

[ -f ~/.ssh/known_hosts ] && { ssh-keygen -R "[localhost]:$SSH_PORT" -f ~/.ssh/known_hosts || exit 1; }
[ ! -f ~/.ssh/id_rsa.pub ] && { ssh-keygen -t rsa -f ~/.ssh/id_rsa -P "" > /dev/null || exit 1; }
docker exec ${NAME} bash -c "mkdir -p /home/$DOCKER_USER/.ssh" || exit 1
docker cp ~/.ssh/id_rsa.pub ${NAME}:/home/$DOCKER_USER/.ssh/authorized_keys || exit 1
docker exec ${NAME} bash -c "chown $DOCKER_USER:$DOCKER_USER -R /home/$DOCKER_USER/.ssh ;chmod 0700 /home/$DOCKER_USER/.ssh; chmod 0600 /home/$DOCKER_USER/.ssh/*" || exit 1
ssh -n -o StrictHostKeyChecking=no -p $SSH_PORT $DOCKER_USER@localhost exit || exit 1

echo "You can now login using:"
echo "   ssh -p $SSH_PORT $DOCKER_USER@localhost"


### User / Group id
if ($UPDATE_UID); then
    echo -n "Setup docker user and group id to match yours "

    docker exec -t ${NAME} bash -c "/bin/loginctl kill-user devel"
    res=3
    max=30
    count=0
    while [ $res -ne 1 ] && [ $count -le $max ]; do
        sleep 1
        docker exec ${NAME} bash -c "loginctl user-status devel |grep sd-pam" 2>/dev/null 1>&2
        res=$?
        echo -n "."
        count=$((count + 1));
    done

    echo -n " ."

     # Set uid
    if docker exec -t ${NAME} bash -c "id $(id -u)" > /dev/null 2>&1 && [ "$(id -u)" != "1664" ]; then
        echo "Cannot set docker devel user id to your id: conflict id $(id -u) !"
        exit 1
    fi
    docker exec -t ${NAME} bash -c "usermod -u $(id -u) $DOCKER_USER" || exit 1
    echo -n "."

    # Set gid
    if docker exec -t ${NAME} bash -c "grep $(id -g) /etc/group" > /dev/null 2>&1; then
        docker exec -t ${NAME} bash -c "usermod -g $(id -g) $DOCKER_USER" || exit 1
    else
        docker exec -t ${NAME} bash -c "groupmod -g $(id -g) $DOCKER_USER" || exit 1
    fi
    echo -n "."

    docker exec -t ${NAME} bash -c "chown -R $DOCKER_USER:$DOCKER_USER /home/$DOCKER_USER" || exit 1
    echo -n "."
    docker exec -t ${NAME} bash -c "chown -R $DOCKER_USER:$DOCKER_USER /tmp/xds*"
    echo -n "."
    docker exec -t ${NAME} bash -c "systemctl start autologin"
    echo -n "."
    ssh -n -p $SSH_PORT $DOCKER_USER@localhost "systemctl --user start xds-server" || exit 1
    echo "."
    docker restart ${NAME}
fi

creation_done=true

### Force xds-server restart
if ($FORCE_RESTART); then
    echo "Restart xds-server..."
    ssh -n -p $SSH_PORT $DOCKER_USER@localhost "systemctl --user restart xds-server" || exit 1
fi

echo "Done, docker container $NAME is ready to be used."
