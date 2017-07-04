# XDS - X(cross) Development System Server

`xds-server` is a web server that allows user to remotely cross build applications.

The first goal is to provide a multi-platform cross development tool with
near-zero installation.
The second goal is to keep application sources locally (on user's machine) to
make it compatible with existing IT policies (e.g. corporate backup or SCM),
and let user to continue to work as usual (use his favorite editor,
keep performance while editing/browsing sources).

This powerful and portable webserver (written in [Go](https://golang.org))
exposes a REST interface over HTTP and also provides a Web dashboard to configure projects and execute _(for now)_ only basics commands.

`xds-server` uses [Syncthing](https://syncthing.net/) tool to synchronize
projects files from user machine to build server machine or container.

> **NOTE**: For now, only Syncthing sharing method is supported to synchronize
projects files. But in a near future and for restricted configurations, `xds-server`
will also support "standard" folder sharing (eg. nfs mount points or docker
volumes).

> **SEE ALSO**: [xds-exec](https://github.com/iotbzh/xds-exec),
wrappers on `exec` commands that allows you to send commands to `xds-server`
and for example build your application from command-line or from your favorite
IDE (such as Netbeans or Visual Studio Code) through `xds-server`.

## How to run

`xds-server` has been designed to easily compile and debug
[AGL](https://www.automotivelinux.org/) applications. That's why `xds-server` has
been integrated into AGL SDK docker container.

>**NOTE** For more info about AGL SDK docker container, please refer to
[AGL SDK Quick Setup](http://docs.automotivelinux.org/docs/getting_started/en/dev/reference/setup-sdk-environment.html)

### Get the container

Load the pre-build AGL SDK docker image including `xds-server`:
```bash
wget -O - http://iot.bzh/download/public/2017/XDS/docker/docker_agl_worker-xds-latest.tar.xz | docker load
```

### List container
You should get `docker.automotivelinux.org/agl/worker-xds:X.Y` image
 
```bash
# List image that we just built
docker images | grep worker-xds

docker.automotivelinux.org/agl/worker-xds       3.2                 786d65b2792c        6 days ago          602MB
```

### Start xds-server within the container

Use provided script to create a new docker image and start a new container:
```bash
# Get script
wget https://raw.githubusercontent.com/iotbzh/xds-server/master/scripts/xds-docker-create-container.sh
# [snip...]

# Create new XDS worker container
bash ./xds-docker-create-container.sh 0 docker.automotivelinux.org/agl/worker-xds:3.99.1
# [snip...]

# Check that new container is running
docker ps | grep worker-xds

b985d81af40c        docker.automotivelinux.org/agl/worker-xds:3.99.1       "/usr/bin/wait_for..."   6 days ago           Up 4 hours          0.0.0.0:8000->8000/tcp, 0.0.0.0:69->69/udp, 0.0.0.0:10809->10809/tcp, 0.0.0.0:2222->22/tcp    agl-worker-seb-laptop-0-seb
```

This container (ID=0) exposes following ports:
  - 8000 : `xds-server` to serve XDS Dashboard
  - 69   : TFTP
  - 2222 : ssh

**`xds-server` is automatically started** as a service on container startup.

If the container is running on your localhost, you can access the web interface (what we call the "Dashboard"):
```bash
xdg-open http://localhost:8000
```

If needed you can status / stop / start  it manually using following commands:
```bash
# Log into docker container
ssh -p 2222 devel@localhost

# Status XDS server:
sudo systemctl status xds-server.service

# Stop XDS server
sudo systemctl stop xds-server.service

# Start XDS server
sudo systemctl start xds-server.service

# Get XDS server logs
sudo journalctl --unit=xds-server.service --output=cat
```

### Manually Start XDS server

XDS server is started as a service by Systemd.
```bash
/lib/systemd/system/xds-server.service
```
This Systemd service starts a bash script `/usr/local/bin/xds-server-start.sh`

If you needed you can change default setting by defining specific environment variables in `/etc/default/xds-server`.
For example to control log level, just set LOGLEVEL env variable knowing that supported *level* are: panic, fatal, error, warn, info, debug.
```bash
echo 'LOGLEVEL=debug' | sudo tee --append /etc/default/xds-server > /dev/null
sudo systemctl restart xds-server.service
tail -f /tmp/xds-server/logs/xds-server.log
```

### Install SDK cross-toolchain

`xds-server` uses cross-toolchain install into directory pointed by `sdkRootDir` setting (see configuration section below for more details).
For now, you need to install manually SDK cross toolchain. There are not embedded into docker image by default because the size of these tarballs is too big.

Use provided `install-agl-sdks` script, for example to install SDK for ARM64 and Intel corei7-64:

```bash
# Install ARM64 SDK (automatic download)
sudo /usr/local/bin/xds-utils/install-agl-sdks.sh --arch aarch64

# Install Intel corei7-64 SDK (using an SDK tarball that has been built or downloaded manually)
sudo /usr/local/bin/xds-utils/install-agl-sdks.sh --arch corei7-64 --file /tmp/poky-agl-glibc-x86_64-agl-demo-platform-crosssdk-corei7-64-toolchain-
3.99.1+snapshot.sh

```

### XDS Dashboard

`xds-server` serves a web-application at [http://localhost:8000](http://localhost:8000) when XDS server is running on your host. Just replace `localhost` by the host name or ip when XDS server is running on another host. So you can now connect your browser to this url and use what we call the **XDS dashboard**.
```bash
xdg-open http://localhost:8000
```

Then follow instructions provided by this dashboard, knowing that the first time
you need to download and start `xds-agent` on your local machine. 

To download this tool, just click on download icon in dashboard configuration page or download one of `xds-agent` released tarball: [https://github.com/iotbzh/xds-agent/releases](https://github.com/iotbzh/xds-agent/releases).

See also `xds-agent` [README file](https://github.com/iotbzh/xds-agent) for more details.

## Build xds-server from scratch

### Dependencies

- Install and setup [Go](https://golang.org/doc/install) version 1.7 or
higher to compile this tool.
- Install [npm](https://www.npmjs.com/)
- Install [gulp](http://gulpjs.com/)

Ubuntu:

```bash
 sudo apt-get install golang npm curl git zip
 sudo npm install -g gulp-cli
```

openSUSE:

```bash
 sudo zypper install go npm git curl zip
 sudo npm install -g gulp-cli
```

Don't forget to open new user session after installing the packages.

### Building

#### Native build

Create a GOPATH variable(must be a full path):
```bash
 export GOPATH=$(realpath ~/workspace_go)
```

Clone this repo into your `$GOPATH/src/github.com/iotbzh` and use delivered Makefile:
```bash
 mkdir -p $GOPATH/src/github.com/iotbzh
 cd $GOPATH/src/github.com/iotbzh
 git clone https://github.com/iotbzh/xds-server.git
 cd xds-server
 make all
```

And to install `xds-server` (by default in `/usr/local/bin`):
```bash
 make install
```

>**NOTE:** Used `DESTDIR` to specify another install directory
>```bash
>make install DESTDIR=$HOME/opt/xds-server
>```

#### XDS docker image

As an alternative to a pre-build image, you can rebuild the container from scratch. 
`xds-server` has been integrated as a flavour of AGL SDK docker image. 
So to rebuild docker image just execute following commands:

```bash
# Clone docker-worker-generator git repo
git clone https://git.automotivelinux.org/AGL/docker-worker-generator
# Start build that will create a docker image
cd docker-worker-generator
make build FLAVOUR=xds
```

### Configuration

`xds-server` configuration is driven by a JSON config file (`config.json`).

Here is the logic to determine which `config.json` file will be used:
1. from command line option: `--config myConfig.json`
2. `$HOME/.xds/config.json` file
3. `<current dir>/config.json` file
4. `<xds-server executable dir>/config.json` file

Supported fields in configuration file are (all fields are optional and listed values are the default values):
```
{
    "httpPort": 8000,                               # HTTP port of client webapp / dashboard
    "webAppDir": "webapp/dist",                     # location of client dashboard (default: webapp/dist)
    "shareRootDir": "${HOME}/.xds/projects",        # root directory where projects will be copied
    "logsDir": "/tmp/logs",                         # directory to store logs (eg. syncthing output)
    "sdkRootDir": "/xdt/sdk",                       # root directory where cross SDKs are installed
    "syncthing": {
        "binDir": "./bin",                          # syncthing binaries directory (default: executable directory)
        "home": "${HOME}/.xds/syncthing-config",    # syncthing home directory (usually .../syncthing-config)
        "gui-address": "http://localhost:8384",     # syncthing gui url (default http://localhost:8384)
        "gui-apikey": "123456789",                  # syncthing api-key to use (default auto-generated)
    }
}
```

>**NOTE:** environment variables are supported by using `${MY_VAR}` syntax.


## Debugging

### XDS server architecture

The server part is written in *Go* and web app / dashboard (client part) in
*Angular2*.

```
|
+-- bin/                where xds-server binary file will be built
|
+-- agent-config.json.in      example of config.json file
|
+-- glide.yaml          Go package dependency file
|
+-- lib/                sources of server part (Go)
|
+-- main.go             main entry point of of Web server (Go)
|
+-- Makefile            makefile including
|
+-- README.md           this readme
|
+-- scripts/            hold various scripts used for installation or startup
|
+-- tools/              temporary directory to hold development tools (like glide)
|
+-- vendor/             temporary directory to hold Go dependencies packages
|
+-- webapp/             source client dashboard (Angular2 app)
```

Visual Studio Code launcher settings can be found into `.vscode/launch.json`.


## TODO:

- replace makefile by build.go to make Windows build support easier
- add more tests
- add more documentation
- add authentication / login (oauth) + HTTPS
- enable syncthing user/password + HTTPS
