"use strict";
//FIXME in VSC/eslint or add to typings declare function require(v: string): any;

// FIXME: Rework based on
//   https://github.com/iotbzh/app-framework-templates/blob/master/templates/hybrid-html5/gulpfile.js
// AND
//   https://github.com/antonybudianto/angular-starter
// and/or
//   https://github.com/smmorneau/tour-of-heroes/blob/master/gulpfile.js

const gulp = require("gulp"),
    gulpif = require('gulp-if'),
    del = require("del"),
    sourcemaps = require('gulp-sourcemaps'),
    tsc = require("gulp-typescript"),
    tsProject = tsc.createProject("tsconfig.json"),
    tslint = require('gulp-tslint'),
    gulpSequence = require('gulp-sequence'),
    rsync = require('gulp-rsync'),
    conf = require('./gulp.conf');


var tslintJsonFile = "./tslint.json"
if (conf.prodMode) {
    tslintJsonFile = "./tslint.prod.json"
}


/**
 * Remove output directory.
 */
gulp.task('clean', (cb) => {
    return del([conf.outDir], cb);
});

/**
 * Lint all custom TypeScript files.
 */
gulp.task('tslint', function() {
    return gulp.src(conf.paths.tsSources)
        .pipe(tslint({
            formatter: 'verbose',
            configuration: tslintJsonFile
        }))
        .pipe(tslint.report());
});

/**
 * Compile TypeScript sources and create sourcemaps in build directory.
 */
gulp.task("compile", ["tslint"], function() {
    var tsResult = gulp.src(conf.paths.tsSources)
        .pipe(sourcemaps.init())
        .pipe(tsProject());
    return tsResult.js
        .pipe(sourcemaps.write(".", { sourceRoot: '/src' }))
        .pipe(gulp.dest(conf.outDir));
});

/**
 * Copy all resources that are not TypeScript files into build directory.
 */
gulp.task("resources", function() {
    return gulp.src(["src/**/*", "!**/*.ts"])
        .pipe(gulp.dest(conf.outDir));
});

/**
 * Copy all assets into build directory.
 */
gulp.task("assets", function() {
    return gulp.src(conf.paths.assets)
        .pipe(gulp.dest(conf.outDir + "/assets"));
});

/**
 * Copy all required libraries into build directory.
 */
gulp.task("libs", function() {
    return gulp.src(conf.paths.node_modules_libs,
        { cwd: "node_modules/**" })    /* Glob required here. */
        .pipe(gulp.dest(conf.outDir + "/lib"));
});

/**
 * Watch for changes in TypeScript, HTML and CSS files.
 */
gulp.task('watch', function () {
    gulp.watch([conf.paths.tsSources], ['compile']).on('change', function (e) {
        console.log('TypeScript file ' + e.path + ' has been changed. Compiling.');
    });
    gulp.watch(["src/**/*.html", "src/**/*.css"], ['resources']).on('change', function (e) {
        console.log('Resource file ' + e.path + ' has been changed. Updating.');
    });
});

/**
 * Build the project.
 */
gulp.task("build", ['compile', 'resources', 'libs', 'assets'], function() {
    console.log("Building the project ...");
});

/**
 * Deploy the project on another machine/container
 */
gulp.task('rsync', function () {
    return gulp.src(conf.outDir)
        .pipe(rsync({
            root: conf.outDir,
            username: conf.deploy.username,
            hostname: conf.deploy.target_ip,
            port: conf.deploy.port || null,
            archive: true,
            recursive: true,
            compress: true,
            progress: false,
            incremental: true,
            destination: conf.deploy.dir
        }));
});

gulp.task('deploy', gulpSequence('build', 'rsync'));