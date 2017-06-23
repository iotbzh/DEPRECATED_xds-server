import { Component, OnInit } from "@angular/core";
import { Observable } from 'rxjs/Observable';
import { FormControl, FormGroup, Validators, FormBuilder } from '@angular/forms';

// Import RxJs required methods
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/filter';
import 'rxjs/add/operator/debounceTime';

import { ConfigService, IConfig, IProject, ProjectType, IxdsAgentPackage } from "../services/config.service";
import { XDSServerService, IServerStatus, IXDSAgentInfo } from "../services/xdsserver.service";
import { XDSAgentService, IAgentStatus } from "../services/xdsagent.service";
import { SyncthingService, ISyncThingStatus } from "../services/syncthing.service";
import { AlertService } from "../services/alert.service";
import { ISdk, SdkService } from "../services/sdk.service";

@Component({
    templateUrl: './app/config/config.component.html',
    styleUrls: ['./app/config/config.component.css']
})

// Inspired from https://embed.plnkr.co/jgDTXknPzAaqcg9XA9zq/
// and from http://plnkr.co/edit/vCdjZM?p=preview

export class ConfigComponent implements OnInit {

    config$: Observable<IConfig>;
    sdks$: Observable<ISdk[]>;
    serverStatus$: Observable<IServerStatus>;
    agentStatus$: Observable<IAgentStatus>;
    localSTStatus$: Observable<ISyncThingStatus>;

    curProj: number;
    userEditedLabel: boolean = false;
    xdsAgentPackages: IxdsAgentPackage[] = [];

    // TODO replace by reactive FormControl + add validation
    syncToolUrl: string;
    xdsAgentUrl: string;
    xdsAgentRetry: string;
    projectsRootDir: string;
    showApplyBtn = {    // Used to show/hide Apply buttons
        "retry": false,
        "rootDir": false,
    };

    addProjectForm: FormGroup;
    pathCtrl = new FormControl("", Validators.required);


    constructor(
        private configSvr: ConfigService,
        private xdsServerSvr: XDSServerService,
        private xdsAgentSvr: XDSAgentService,
        private stSvr: SyncthingService,
        private sdkSvr: SdkService,
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
        this.sdks$ = this.sdkSvr.Sdks$;
        this.serverStatus$ = this.xdsServerSvr.Status$;
        this.agentStatus$ = this.xdsAgentSvr.Status$;
        this.localSTStatus$ = this.stSvr.Status$;

        // Bind xdsAgentUrl to baseURL
        this.config$.subscribe(cfg => {
            this.syncToolUrl = cfg.localSThg.URL;
            this.xdsAgentUrl = cfg.xdsAgent.URL;
            this.xdsAgentRetry = String(cfg.xdsAgent.retry);
            this.projectsRootDir = cfg.projectsRootDir;
            this.xdsAgentPackages = cfg.xdsAgentPackages;
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
                let rr = parseInt(this.xdsAgentRetry, 10);
                if (re.test(this.xdsAgentRetry) && rr >= 0) {
                    this.configSvr.xdsAgentRetry = rr;
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

    xdsAgentRestartConn() {
        let aurl = this.xdsAgentUrl;
        this.configSvr.syncToolURL = this.syncToolUrl;
        this.configSvr.xdsAgentUrl = aurl;
        this.configSvr.loadProjects();
    }

    onSubmit() {
        let formVal = this.addProjectForm.value;

        this.configSvr.addProject({
            label: formVal['label'],
            path: formVal['path'],
            type: ProjectType.SYNCTHING,
            // FIXME: allow to set defaultSdkID from New Project config panel
        });
    }

}