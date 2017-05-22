import { Component, OnInit, Input } from "@angular/core";
import { Observable } from 'rxjs';
import { FormControl, FormGroup, Validators, FormBuilder } from '@angular/forms';

import 'rxjs/add/operator/scan';
import 'rxjs/add/operator/startWith';

import { XDSAgentService, IXDSDeploy } from "../../services/xdsagent.service";
import { ConfigService, IConfig, IProject } from "../../services/config.service";
import { AlertService, IAlert } from "../../services/alert.service";
import { SdkService } from "../../services/sdk.service";

@Component({
    selector: 'panel-deploy',
    moduleId: module.id,
    templateUrl: './deploy.component.html',
    styleUrls: ['./deploy.component.css']
})

export class DeployComponent implements OnInit {

    @Input() curProject: IProject;

    deploying: boolean;
    deployForm: FormGroup;

    constructor(private configSvr: ConfigService,
        private xdsAgent: XDSAgentService,
        private fb: FormBuilder,
        private alert: AlertService,
    ) {
        this.deployForm = fb.group({
            boardIP: ["", Validators.nullValidator],
            wgtFile: ["", Validators.nullValidator],
        });
    }

    ngOnInit() {
        this.deploying = false;
        if (this.curProject && this.curProject.path) {
            this.deployForm.patchValue({ wgtFile: this.curProject.path });
        }
    }

    deploy() {
        this.deploying = true;

        this.xdsAgent.deploy(
            {
                boardIP: this.deployForm.value.boardIP,
                file: this.deployForm.value.wgtFile
            }
        ).subscribe(res => {
            this.deploying = false;
        }, err => {
            this.deploying = false;
            let msg = '<span>ERROR while deploying "' + this.deployForm.value.wgtFile + '"<br>';
            msg += err;
            msg += '</span>';
            this.alert.error(msg);
        });
    }
}