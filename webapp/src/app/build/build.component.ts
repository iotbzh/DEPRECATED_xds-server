import { Component, AfterViewChecked, ElementRef, ViewChild, OnInit } from '@angular/core';
import { Observable } from 'rxjs';
import { FormControl, FormGroup, Validators, FormBuilder } from '@angular/forms';

import 'rxjs/add/operator/scan';
import 'rxjs/add/operator/startWith';

import { XDSServerService, ICmdOutput } from "../common/xdsserver.service";
import { ConfigService, IConfig, IProject } from "../common/config.service";
import { AlertService, IAlert } from "../common/alert.service";

@Component({
    selector: 'build',
    moduleId: module.id,
    templateUrl: './build.component.html',
    styleUrls: ['./build.component.css']
})

export class BuildComponent implements OnInit, AfterViewChecked {
    @ViewChild('scrollOutput') private scrollContainer: ElementRef;

    config$: Observable<IConfig>;

    buildForm: FormGroup;
    subpathCtrl = new FormControl("", Validators.required);

    public cmdOutput: string;
    public confValid: boolean;
    public curProject: IProject;
    public cmdInfo: string;

    private startTime: Map<string, number> = new Map<string, number>();

    // I initialize the app component.
    constructor(private configSvr: ConfigService, private sdkSvr: XDSServerService,
        private fb: FormBuilder, private alertSvr: AlertService
    ) {
        this.cmdOutput = "";
        this.confValid = false;
        this.cmdInfo = "";      // TODO: to be remove (only for debug)
        this.buildForm = fb.group({ subpath: this.subpathCtrl });
    }

    ngOnInit() {
        this.config$ = this.configSvr.conf;
        this.config$.subscribe((cfg) => {
            this.curProject = cfg.projects[0];

            this.confValid = (cfg.projects.length && this.curProject.id != null);
        });

        // Command output data tunneling
        this.sdkSvr.CmdOutput$.subscribe(data => {
            this.cmdOutput += data.stdout + "\n";
        });

        // Command exit
        this.sdkSvr.CmdExit$.subscribe(exit => {
            if (this.startTime.has(exit.cmdID)) {
                this.cmdInfo = 'Last command duration: ' + this._computeTime(this.startTime.get(exit.cmdID));
                this.startTime.delete(exit.cmdID);
            }

            if (exit && exit.code !== 0) {
                this.cmdOutput += "--- Command exited with code " + exit.code + " ---\n\n";
            }
        });

        this._scrollToBottom();
    }

    ngAfterViewChecked() {
        this._scrollToBottom();
    }

    reset() {
        this.cmdOutput = '';
    }

    make(args: string) {
        let prjID = this.curProject.id;

        this.cmdOutput += this._outputHeader();

        let t0 = performance.now();
        this.cmdInfo = 'Start build of ' + prjID + ' at ' + t0;

        this.sdkSvr.make(prjID, this.buildForm.value.subpath, args)
            .subscribe(res => {
                this.startTime.set(String(res.cmdID), t0);
            },
            err => {
                this.cmdInfo = 'Last command duration: ' + this._computeTime(t0);
                this.alertSvr.add({ type: "danger", msg: 'ERROR: ' + err });
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