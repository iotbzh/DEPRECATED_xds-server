"use strict";

module.exports = {
    prodMode: process.env.PRODUCTION || false,
    outDir: "dist",
    paths: {
        tsSources: ["src/**/*.ts"],
        srcDir: "src",
        assets: ["assets/**"],
        node_modules_libs: [
            'core-js/client/shim.min.js',
            'reflect-metadata/Reflect.js',
            'rxjs/**/*.js',
            'socket.io-client/dist/socket.io*.js',
            'systemjs/dist/system-polyfills.js',
            'systemjs/dist/system.src.js',
            'zone.js/dist/**',
            '@angular/**/bundles/**',
            'ngx-cookie/bundles/**',
            'ngx-bootstrap/bundles/**',
            'bootstrap/dist/**',
            'moment/*.min.js',
            'font-awesome-animation/dist/font-awesome-animation.min.css',
            'font-awesome/css/font-awesome.min.css',
            'font-awesome/fonts/**'
        ]
    },
    deploy: {
        target_ip: 'ip',
        username: "user",
        //port: 6666,
        dir: '/tmp/xds-server'
    }
}