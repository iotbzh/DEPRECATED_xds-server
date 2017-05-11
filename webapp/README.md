XDS Dashboard
=============

This is the web application dashboard for Cross Development System.

## 1. Prerequisites

*nodejs* must be installed on your system and the below global node packages must be installed:

> sudo npm install -g gulp-cli

## 2. Installing dependencies

Install dependencies by running the following command:

> npm install

`node_modules` and `typings` directories will be created during the install.

## 3. Building the project

Build the project by running the following command:

> npm run clean & npm run build

`dist` directory will be created during the build

## 4. Starting the application

Start the application by running the following command:

> npm start

The application will be displayed in the browser.


## TODO

- Upgrade to angular 2.4.9 or 2.4.10 AND rxjs 5.2.0
- Complete README + package.json
- Add prod mode and use update gulpfile tslint: "./tslint/prod.json"
- Generate a bundle minified file, using systemjs-builder or find a better way
   http://stackoverflow.com/questions/35280582/angular2-too-many-file-requests-on-load
- Add SASS support
   http://foundation.zurb.com/sites/docs/sass.html