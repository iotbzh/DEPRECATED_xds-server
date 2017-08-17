import { Component, Input, Pipe, PipeTransform } from '@angular/core';
import { ConfigService, IProject, ProjectType } from "../services/config.service";
import { AlertService } from "../services/alert.service";

@Component({
    selector: 'project-card',
    template: `
        <div class="row">
            <div class="col-xs-12">
                <div class="text-right" role="group">
                    <button class="btn btn-link" (click)="delete(project)">
                        <span class="fa fa-trash fa-size-x2"></span>
                    </button>
                </div>
            </div>
        </div>

        <table class="table table-striped">
            <tbody>
            <tr>
                <th><span class="fa fa-fw fa-id-badge"></span>&nbsp;<span>Project ID</span></th>
                <td>{{ project.id }}</td>
            </tr>
            <tr>
                <th><span class="fa fa-fw fa-exchange"></span>&nbsp;<span>Sharing type</span></th>
                <td>{{ project.type | readableType }}</td>
            </tr>
            <tr>
                <th><span class="fa fa-fw fa-folder-open-o"></span>&nbsp;<span>Local path</span></th>
                <td>{{ project.pathClient }}</td>
            </tr>
            <tr *ngIf="project.pathServer && project.pathServer != ''">
                <th><span class="fa fa-fw fa-folder-open-o"></span>&nbsp;<span>Server path</span></th>
                <td>{{ project.pathServer }}</td>
            </tr>
            <tr>
                <th><span class="fa fa-fw fa-flag"></span>&nbsp;<span>Status</span></th>
                <td>{{ project.status }} - {{ project.isInSync ? "Up to Date" : "Out of Sync"}}
                    <button *ngIf="!project.isInSync" class="btn btn-link" (click)="sync(project)">
                        <span class="fa fa-refresh fa-size-x2"></span>
                    </button>
                </td>
            </tr>
            </tbody>
        </table >
    `,
    styleUrls: ['./app/config/config.component.css']
})

export class ProjectCardComponent {

    @Input() project: IProject;

    constructor(
        private alert: AlertService,
        private configSvr: ConfigService
    ) {
    }

    delete(prj: IProject) {
        this.configSvr.deleteProject(prj)
            .subscribe(res => {
            }, err => {
                this.alert.error("Delete local ERROR: " + err);
            });
    }

    sync(prj: IProject) {
        this.configSvr.syncProject(prj)
            .subscribe(res => {
            }, err => {
                this.alert.error("ERROR: " + err);
            });
    }

}

// Remove APPS. prefix if translate has failed
@Pipe({
    name: 'readableType'
})

export class ProjectReadableTypePipe implements PipeTransform {
    transform(type: ProjectType): string {
        switch (type) {
            case ProjectType.NATIVE_PATHMAP: return "Native (path mapping)";
            case ProjectType.SYNCTHING: return "Cloud (Syncthing)";
            default: return String(type);
        }
    }
}
