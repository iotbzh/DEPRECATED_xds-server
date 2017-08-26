import { Component, ViewChild, OnInit } from "@angular/core";
import { Observable } from 'rxjs/Observable';
import { FormControl, FormGroup, Validators, FormBuilder } from '@angular/forms';
import { CollapseModule } from 'ngx-bootstrap/collapse';

import { ConfigService, IConfig, IxdsAgentPackage } from "../services/config.service";
import { XDSServerService, IServerStatus, IXDSAgentInfo } from "../services/xdsserver.service";
import { XDSAgentService, IAgentStatus } from "../services/xdsagent.service";
import { SyncthingService, ISyncThingStatus } from "../services/syncthing.service";
import { AlertService } from "../services/alert.service";
import { ISdk, SdkService } from "../services/sdk.service";
import { ProjectAddModalComponent } from "../projects/projectAddModal.component";
import { SdkAddModalComponent } from "../sdks/sdkAddModal.component";

@Component({
    templateUrl: './app/config/config.component.html',
    styleUrls: ['./app/config/config.component.css']
})

// Inspired from https://embed.plnkr.co/jgDTXknPzAaqcg9XA9zq/
// and from http://plnkr.co/edit/vCdjZM?p=preview

export class ConfigComponent implements OnInit {
    @ViewChild('childProjectModal') childProjectModal: ProjectAddModalComponent;
    @ViewChild('childSdkModal') childSdkModal: SdkAddModalComponent;

    config$: Observable<IConfig>;
    sdks$: Observable<ISdk[]>;
    serverStatus$: Observable<IServerStatus>;
    agentStatus$: Observable<IAgentStatus>;
    localSTStatus$: Observable<ISyncThingStatus>;

    curProj: number;
    userEditedLabel: boolean = false;
    xdsAgentPackages: IxdsAgentPackage[] = [];

    gConfigIsCollapsed: boolean = true;
    sdksIsCollapsed: boolean = true;
    projectsIsCollapsed: boolean = false;

    // TODO replace by reactive FormControl + add validation
    syncToolUrl: string;
    xdsAgentUrl: string;
    xdsAgentRetry: string;
    projectsRootDir: string;    // FIXME: should be remove when projectAddModal will always return full path
    showApplyBtn = {    // Used to show/hide Apply buttons
        "retry": false,
        "rootDir": false,
    };

    constructor(
        private configSvr: ConfigService,
        private xdsServerSvr: XDSServerService,
        private xdsAgentSvr: XDSAgentService,
        private stSvr: SyncthingService,
        private sdkSvr: SdkService,
        private alert: AlertService,
    ) {
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

}
