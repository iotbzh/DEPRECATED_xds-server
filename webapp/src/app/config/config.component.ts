import { Component, OnInit } from "@angular/core";
import { Observable } from 'rxjs/Observable';
import { FormControl, FormGroup, Validators, FormBuilder } from '@angular/forms';

// Import RxJs required methods
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/filter';
import 'rxjs/add/operator/debounceTime';

import { ConfigService, IConfig, IProject, ProjectType, ProjectTypes,
    IxdsAgentPackage } from "../services/config.service";
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
    projectTypes = ProjectTypes;

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
    pathCliCtrl = new FormControl("", Validators.required);
    pathSvrCtrl = new FormControl("", Validators.required);

    constructor(
        private configSvr: ConfigService,
        private xdsServerSvr: XDSServerService,
        private xdsAgentSvr: XDSAgentService,
        private stSvr: SyncthingService,
        private sdkSvr: SdkService,
        private alert: AlertService,
        private fb: FormBuilder
    ) {
        // Define types (first one is special/placeholder)
        this.projectTypes.unshift({value: -1, display: "--Select a type--"});
        let selectedType = this.projectTypes[0].value;

        this.curProj = 0;
        this.addProjectForm = fb.group({
            pathCli: this.pathCliCtrl,
            pathSvr: this.pathSvrCtrl,
            label: ["", Validators.nullValidator],
            type: [selectedType, Validators.pattern("[0-9]+")],
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
        this.pathCliCtrl.valueChanges
            .debounceTime(100)
            .filter(n => n)
            .map(n => "Project_" + n.split('/')[0])
            .subscribe(value => {
                if (value && !this.userEditedLabel) {
                    this.addProjectForm.patchValue({ label: value });
                }
            });

        // Select 1 first type by default
        // SEB this.typeCtrl.setValue({type: ProjectTypes[0].value});
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
        let aUrl = this.xdsAgentUrl;
        this.configSvr.syncToolURL = this.syncToolUrl;
        this.configSvr.xdsAgentUrl = aUrl;
        this.configSvr.loadProjects();
    }

    onSubmit() {
        let formVal = this.addProjectForm.value;

        let type = formVal['type'].value;
        let numType = Number(formVal['type']);
        this.configSvr.addProject({
            label: formVal['label'],
            pathClient: formVal['pathCli'],
            pathServer: formVal['pathSvr'],
            type: numType,
            // FIXME: allow to set defaultSdkID from New Project config panel
        });
    }

}
