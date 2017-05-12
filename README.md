# XDS - X(cross) Development System

XDS-server is a web server that allows user to remotely cross build applications.

The first goal is to provide a multi-platform cross development tool with
near-zero installation.
The second goals is to keep application sources locally (on user's machine) to
make it compatible with existing IT policies (e.g. corporate backup or SCM).

This powerful webserver (written in [Go](https://golang.org)) exposes a REST
interface over HTTP and also provides a Web dashboard to configure projects and execute only _(for now)_ basics commands.

XDS-server also uses [Syncthing](https://syncthing.net/) tool to synchronize
projects files from user machine to build server machine or container.

> **NOTE**: For now, only Syncthing sharing method is supported to synchronize
projects files.

> **SEE ALSO**: [xds-make](https://github.com/iotbzh/xds-make), a wrapper on `make`
command that allows you to build your application from command-line through
xds-server.


## How to build

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

## How to run

## Configuration

xds-server configuration is driven by a JSON config file (`config.json`).

Here is the logic to determine which `config.json` file will be used:
1. from command line option: `--config myConfig.json`
2. `$HOME/.xds/config.json` file
3. `<xds-server executable dir>/config.json` file

Supported fields in configuration file are:
```json
{
    "webAppDir": "location of client dashboard (default: webapp/dist)",
    "shareRootDir": "root directory where projects will be copied",
    "syncthing": {
        "home": "syncthing home directory (usually .../syncthing-config)",
        "gui-address": "syncthing gui url (default http://localhost:8384)"
    }
}
```

>**NOTE:** environment variables are supported by using `${MY_VAR}` syntax.

## Start-up

```bash
./bin/xds-server -c config.json
```

**TODO**: add notes about Syncthing setup and startup


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
