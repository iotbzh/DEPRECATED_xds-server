(function (global) {
    System.config({
        paths: {
            // paths serve as alias
            'npm:': 'lib/'
        },
        bundles: {
            "npm:rxjs-system-bundle/Rx.system.min.js": [
                "rxjs",
                "rxjs/*",
                "rxjs/operator/*",
                "rxjs/observable/*",
                "rxjs/scheduler/*",
                "rxjs/symbol/*",
                "rxjs/add/operator/*",
                "rxjs/add/observable/*",
                "rxjs/util/*"
            ]
        },
        // map tells the System loader where to look for things
        map: {
            // our app is within the app folder
            app: 'app',
            // angular bundles
            '@angular/core': 'npm:@angular/core/bundles/core.umd.js',
            '@angular/common': 'npm:@angular/common/bundles/common.umd.js',
            '@angular/compiler': 'npm:@angular/compiler/bundles/compiler.umd.js',
            '@angular/platform-browser': 'npm:@angular/platform-browser/bundles/platform-browser.umd.js',
            '@angular/platform-browser-dynamic': 'npm:@angular/platform-browser-dynamic/bundles/platform-browser-dynamic.umd.js',
            '@angular/http': 'npm:@angular/http/bundles/http.umd.js',
            '@angular/router': 'npm:@angular/router/bundles/router.umd.js',
            '@angular/forms': 'npm:@angular/forms/bundles/forms.umd.js',
            'ngx-cookie': 'npm:ngx-cookie/bundles/ngx-cookie.umd.js',
            // ng2-bootstrap
            'moment': 'npm:moment',
            'ngx-bootstrap/alert': 'npm:ngx-bootstrap/bundles/ngx-bootstrap.umd.min.js',
            'ngx-bootstrap/modal': 'npm:ngx-bootstrap/bundles/ngx-bootstrap.umd.min.js',
            'ngx-bootstrap/accordion': 'npm:ngx-bootstrap/bundles/ngx-bootstrap.umd.min.js',
            'ngx-bootstrap/carousel': 'npm:ngx-bootstrap/bundles/ngx-bootstrap.umd.min.js',
            'ngx-bootstrap/popover': 'npm:ngx-bootstrap/bundles/ngx-bootstrap.umd.min.js',
            'ngx-bootstrap/dropdown': 'npm:ngx-bootstrap/bundles/ngx-bootstrap.umd.min.js',
            // other libraries
            'socket.io-client': 'npm:socket.io-client/dist/socket.io.min.js'
        },
        // packages tells the System loader how to load when no filename and/or no extension
        packages: {
            'app': {
                main: './main.js',
                defaultExtension: 'js'
            },
            'rxjs': {
                defaultExtension: false
            },
            'socket.io-client': {
                defaultExtension: 'js'
            },
            'ngx-bootstrap': {
                format: 'cjs',
                main: 'bundles/ng2-bootstrap.umd.js',
                defaultExtension: 'js'
            },
            'moment': {
                main: 'moment.js',
                defaultExtension: 'js'
            }
        }
    });
})(this);