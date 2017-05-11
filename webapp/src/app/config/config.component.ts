import { Component, OnInit } from "@angular/core";
import { Observable } from 'rxjs/Observable';
import { FormControl, FormGroup, Validators, FormBuilder } from '@angular/forms';

// Import RxJs required methods
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/filter';
import 'rxjs/add/operator/debounceTime';

import { ConfigService, IConfig, IProject, ProjectType } from "../common/config.service";
import { XDSServerService, IServerStatus } from "../common/xdsserver.service";
import { SyncthingService, ISyncThingStatus } from "../common/syncthing.service";
import { AlertService } from "../common/alert.service";

@Component({
    templateUrl: './app/config/config.component.html',
    styleUrls: ['./app/config/config.component.css']
})

// Inspired from https://embed.plnkr.co/jgDTXknPzAaqcg9XA9zq/
// and from http://plnkr.co/edit/vCdjZM?p=preview

export class ConfigComponent implements OnInit {

    config$: Observable<IConfig>;
    severStatus$: Observable<IServerStatus>;
    localSTStatus$: Observable<ISyncThingStatus>;

    curProj: number;
    userEditedLabel: boolean = false;

    // TODO replace by reactive FormControl + add validation
    syncToolUrl: string;
    syncToolRetry: string;
    projectsRootDir: string;
    showApplyBtn = {    // Used to show/hide Apply buttons
        "retry": false,
        "rootDir": false,
    };

    addProjectForm: FormGroup;
    pathCtrl = new FormControl("", Validators.required);


    constructor(
        private configSvr: ConfigService,
        private sdkSvr: XDSServerService,
        private stSvr: SyncthingService,
        private alert: AlertService,
        private fb: FormBuilder
    ) {
        // FIXME implement multi project support
        this.curProj = 0;
        this.addProjectForm = fb.group({
            path: this.pathCtrl,
            label: ["", Validators.nullValidator],
        });
    }

    ngOnInit() {
        this.config$ = this.configSvr.conf;
        this.severStatus$ = this.sdkSvr.Status$;
        this.localSTStatus$ = this.stSvr.Status$;

        // Bind syncToolUrl to baseURL
        this.config$.subscribe(cfg => {
            this.syncToolUrl = cfg.localSThg.URL;
            this.syncToolRetry = String(cfg.localSThg.retry);
            this.projectsRootDir = cfg.projectsRootDir;
        });

        // Auto create label name
        this.pathCtrl.valueChanges
            .debounceTime(100)
            .filter(n => n)
            .map(n => "Project_" + n.split('/')[0])
            .subscribe(value => {
                if (value && !this.userEditedLabel) {
                    this.addProjectForm.patchValue({ label: value });
                }
            });
    }

    onKeyLabel(event: any) {
        this.userEditedLabel = (this.addProjectForm.value.label !== "");
    }

    submitGlobConf(field: string) {
        switch (field) {
            case "retry":
                let re = new RegExp('^[0-9]+$');
                let rr = parseInt(this.syncToolRetry, 10);
                if (re.test(this.syncToolRetry) && rr >= 0) {
                    this.configSvr.syncToolRetry = rr;
                } else {
                    this.alert.warning("Not a valid number", true);
                }
                break;
            case "rootDir":
                this.configSvr.projectsRootDir = this.projectsRootDir;
                break;
            default:
                return;
        }
        this.showApplyBtn[field] = false;
    }

    syncToolRestartConn() {
        this.configSvr.syncToolURL = this.syncToolUrl;
        this.configSvr.loadProjects();
    }

    onSubmit() {
        let formVal = this.addProjectForm.value;

        this.configSvr.addProject({
            label: formVal['label'],
            path: formVal['path'],
            type: ProjectType.SYNCTHING,
        });
    }

}