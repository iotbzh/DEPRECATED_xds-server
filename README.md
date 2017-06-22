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

> **SEE ALSO**: [xds-exec and xds-make](https://github.com/iotbzh/xds-make),
wrappers on `exec` and `make` commands that allows you to send command to
`xds-server` and consequently build your application from command-line or from
your favorite IDE (eg. Netbeans or Visual Studio Code) through `xds-server`.

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

### Start xds-server within the container

Use provided script to create a new docker image and start a new container:
```bash
> wget https://raw.githubusercontent.com/iotbzh/xds-server/master/scripts/xds-docker-create-container.sh
> bash ./xds-docker-create-container.sh 0 docker.automotivelinux.org/agl/worker-xds:3.99.1

> docker ps
CONTAINER ID        IMAGE                                               COMMAND                  CREATED              STATUS              PORTS                                                                                         NAMES
b985d81af40c        docker.automotivelinux.org/agl/worker-xds:3.99.1       "/usr/bin/wait_for..."   6 days ago           Up 4 hours          0.0.0.0:8000->8000/tcp, 0.0.0.0:69->69/udp, 0.0.0.0:10809->10809/tcp, 0.0.0.0:2222->22/tcp    agl-worker-seb-laptop-0-seb
```

This container exposes following ports:
  - 8000 : `xds-server` to serve XDS Dashboard
  - 69   : TFTP
  - 2222 : ssh

`xds-server` is automatically started as a service on container startup.
If needed you can stop / start it manually using following commands:
```bash
> ssh -p 2222 devel@localhost

[15:59:58] devel@agl-worker-seb-laptop-0-seb:~$ /usr/local/bin/xds-server-stop.sh

[15:59:58] devel@agl-worker-seb-laptop-0-seb:~$ /usr/local/bin/xds-server-start.sh
```

On `xds-server` startup, you should get the following output:
```
### Configuration in config.json:
{
    "webAppDir": "/usr/local/bin/www-xds-server",
    "shareRootDir": "/home/devel/.xds/share",
    "logsDir": "/tmp/xds-server/logs",
    "sdkRootDir": "/xdt/sdk",
    "syncthing": {
        "binDir": "/usr/local/bin",
        "home": "/home/devel/.xds/syncthing-config",
        "gui-address": "http://localhost:8384",
        "gui-apikey": "1234abcezam"
    }
}

### Start XDS server
nohup /usr/local/bin/xds-server --config /home/devel/.xds/config.json -log warn > /tmp/xds-server/logs/xds-server.log 2>&1
pid=22379
```

>**NOTE:** You can set LOGLEVEL env variable to increase log level if you need it.
> For example, to set log level to "debug" mode : ` LOGLEVEL=debug /usr/local/bin/xds-server-start.sh`

### Install SDK cross-toolchain

`xds-server` uses cross-toolchain install into directory pointed by `sdkRootDir` setting (see configuration section below for more details).
For now, you need to install manually SDK cross toolchain. There are not embedded into docker image by default because the size of these tarballs is too big.

Use provided `install-agl-sdks` script, for example to install SDK for ARM64:

```bash
/usr/local/bin/xds-utils/install-agl-sdks.sh --aarch aarch64
```

### XDS Dashboard

`xds-server` serves a web-application (default port 8000:
[http://localhost:8000](http://localhost:8000) ). So you can now connect your browser to this url and use what we call the **XDS dashboard**.

Then follow instructions provided by this dashboard, knowing that the first time
you need to download and start `xds-agent` on your local machine. To download
this tool, just click on download icon in dashboard configuration page or download one of `xds-agent` released tarball: [https://github.com/iotbzh/xds-agent/releases](https://github.com/iotbzh/xds-agent/releases).

See also `xds-agent` [README file](https://github.com/iotbzh/xds-agent) for more
details.


## Build xds-server from scratch

### Dependencies

- Install and setup [Go](https://golang.org/doc/install) version 1.7 or
higher to compile this tool.
- Install [npm](https://www.npmjs.com/) : `sudo apt install npm`
- Install [gulp](http://gulpjs.com/) : `sudo npm install -g gulp-cli`

### Building

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

### Configuration

`xds-server` configuration is driven by a JSON config file (`config.json`).

Here is the logic to determine which `config.json` file will be used:
1. from command line option: `--config myConfig.json`
2. `$HOME/.xds/config.json` file
3. `<current dir>/config.json` file
4. `<xds-server executable dir>/config.json` file

Supported fields in configuration file are (all fields are optional and listed values are the default values):
```json
{
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

## Start-up

Use `xds-server-start.sh` script to start all requested tools
```bash
/usr/local/bin/xds-server-start.sh
```

>**NOTE** you can define some environment variables to setup for example
logging level `LOGLEVEL` or change logs directory `LOGDIR`.
See head section of `xds-server-start.sh` file to see all configurable variables.


### Create XDS AGL docker worker container

`xds-server` has been integrated as a flavour of AGL SDK docker image. So to rebuild
docker image just execute following commands:
```bash
git clone https://git.automotivelinux.org/AGL/docker-worker-generator
cd docker-worker-generator
make build FLAVOUR=xds
```

You should get `docker.automotivelinux.org/agl/worker-xds:X.Y` image

```bash
docker images
REPOSITORY                                      TAG                 IMAGE ID            CREATED             SIZE
docker.automotivelinux.org/agl/worker-xds       3.2                 786d65b2792c        6 days ago          602MB
```

## Debugging

### XDS server architecture

The server part is written in *Go* and web app / dashboard (client part) in
*Angular2*.

```
|
+-- bin/                where xds-server binary file will be built
|
+-- config.json.in      example of config.json file
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
