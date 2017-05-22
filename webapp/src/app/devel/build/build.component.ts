import { Component, AfterViewChecked, ElementRef, ViewChild, OnInit, Input } from '@angular/core';
import { Observable } from 'rxjs';
import { FormControl, FormGroup, Validators, FormBuilder } from '@angular/forms';
import { CookieService } from 'ngx-cookie';

import 'rxjs/add/operator/scan';
import 'rxjs/add/operator/startWith';

import { XDSServerService, ICmdOutput } from "../../services/xdsserver.service";
import { ConfigService, IConfig, IProject } from "../../services/config.service";
import { AlertService, IAlert } from "../../services/alert.service";
import { SdkService } from "../../services/sdk.service";

@Component({
    selector: 'panel-build',
    moduleId: module.id,
    templateUrl: './build.component.html',
    styleUrls: ['./build.component.css']
})

export class BuildComponent implements OnInit, AfterViewChecked {
    @ViewChild('scrollOutput') private scrollContainer: ElementRef;

    @Input() curProject: IProject;

    buildForm: FormGroup;
    subpathCtrl = new FormControl("", Validators.required);
    debugEnable: boolean = false;

    public cmdOutput: string;
    public cmdInfo: string;

    private startTime: Map<string, number> = new Map<string, number>();

    constructor(private configSvr: ConfigService,
        private xdsSvr: XDSServerService,
        private fb: FormBuilder,
        private alertSvr: AlertService,
        private sdkSvr: SdkService,
        private cookie: CookieService,
    ) {
        this.cmdOutput = "";
        this.cmdInfo = "";      // TODO: to be remove (only for debug)
        this.buildForm = fb.group({
            subpath: this.subpathCtrl,
            cmdArgs: ["", Validators.nullValidator],
            envVars: ["", Validators.nullValidator],
        });
    }

    ngOnInit() {
        // Command output data tunneling
        this.xdsSvr.CmdOutput$.subscribe(data => {
            this.cmdOutput += data.stdout + "\n";
        });

        // Command exit
        this.xdsSvr.CmdExit$.subscribe(exit => {
            if (this.startTime.has(exit.cmdID)) {
                this.cmdInfo = 'Last command duration: ' + this._computeTime(this.startTime.get(exit.cmdID));
                this.startTime.delete(exit.cmdID);
            }

            if (exit && exit.code !== 0) {
                this.cmdOutput += "--- Command exited with code " + exit.code + " ---\n\n";
            }
        });

        this._scrollToBottom();

        // only use for debug
        this.debugEnable = (this.cookie.get("debug_build") !== "");
    }

    ngAfterViewChecked() {
        this._scrollToBottom();
    }

    reset() {
        this.cmdOutput = '';
    }

    preBuild() {
        this._exec(
            "mkdir -p build && cd build && cmake ..",
            this.buildForm.value.subpath,
            [],
            this.buildForm.value.envVars);
    }

    build() {
        this._exec(
            "cd build && make",
            this.buildForm.value.subpath,
            this.buildForm.value.cmdArgs,
            this.buildForm.value.envVars
        );
    }

    populate() {
        this._exec(
            "SEB_TODO_script_populate",
            this.buildForm.value.subpath,
            [], // args
            this.buildForm.value.envVars
        );
    }

    execCmd() {
        this._exec(
            this.buildForm.value.cmdArgs,
            this.buildForm.value.subpath,
            [],
            this.buildForm.value.envVars
        );
    }

    private _exec(cmd: string, dir: string, args: string[], env: string) {
        if (!this.curProject) {
            this.alertSvr.warning('No active project', true);
        }

        let prjID = this.curProject.id;

        this.cmdOutput += this._outputHeader();

        let sdkid = this.sdkSvr.getCurrentId();

        // Detect key=value in env string to build array of string
        let envArr = [];
        env.split(';').forEach(v => envArr.push(v.trim()));

        let t0 = performance.now();
        this.cmdInfo = 'Start build of ' + prjID + ' at ' + t0;

        this.xdsSvr.exec(prjID, dir, cmd, sdkid, args, envArr)
            .subscribe(res => {
                this.startTime.set(String(res.cmdID), t0);
            },
            err => {
                this.cmdInfo = 'Last command duration: ' + this._computeTime(t0);
                this.alertSvr.error('ERROR: ' + err);
            });
    }

    make(args: string) {
        if (!this.curProject) {
            this.alertSvr.warning('No active project', true);
        }

        let prjID = this.curProject.id;

        this.cmdOutput += this._outputHeader();

        let sdkid = this.sdkSvr.getCurrentId();

        let argsArr = args ? args.split(' ') : this.buildForm.value.cmdArgs.split(' ');

        // Detect key=value in env string to build array of string
        let envArr = [];
        this.buildForm.value.envVars.split(';').forEach(v => envArr.push(v.trim()));

        let t0 = performance.now();
        this.cmdInfo = 'Start build of ' + prjID + ' at ' + t0;

        this.xdsSvr.make(prjID, this.buildForm.value.subpath, sdkid, argsArr, envArr)
            .subscribe(res => {
                this.startTime.set(String(res.cmdID), t0);
            },
            err => {
                this.cmdInfo = 'Last command duration: ' + this._computeTime(t0);
                this.alertSvr.error('ERROR: ' + err);
            });
    }

    private _scrollToBottom(): void {
        try {
            this.scrollContainer.nativeElement.scrollTop = this.scrollContainer.nativeElement.scrollHeight;
        } catch (err) { }
    }

    private _computeTime(t0: number, t1?: number): string {
        let enlap = Math.round((t1 || performance.now()) - t0);
        if (enlap < 1000.0) {
            return enlap.toFixed(2) + ' ms';
        } else {
            return (enlap / 1000.0).toFixed(3) + ' seconds';
        }
    }

    private _outputHeader(): string {
        return "--- " + new Date().toString() + " ---\n";
    }

    private _outputFooter(): string {
        return "\n";
    }
}