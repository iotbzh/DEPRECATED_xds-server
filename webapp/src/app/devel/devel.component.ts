import { Component } from '@angular/core';

import { Observable } from 'rxjs';

import { ConfigService, IConfig, IProject } from "../services/config.service";

@Component({
    selector: 'devel',
    moduleId: module.id,
    templateUrl: './devel.component.html',
    styleUrls: ['./devel.component.css'],
})

export class DevelComponent {

    curPrj: IProject;
    config$: Observable<IConfig>;

    constructor(private configSvr: ConfigService) {
    }

    ngOnInit() {
        this.config$ = this.configSvr.conf;
        this.config$.subscribe((cfg) => {
            if ("projects" in cfg) {
                this.curPrj = cfg.projects[0];
            } else {
                this.curPrj = null;
            }
        });
    }
}