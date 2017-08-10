import { Component, Input, Pipe, PipeTransform } from '@angular/core';
import { ConfigService, IProject, ProjectType } from "../services/config.service";

@Component({
    selector: 'project-card',
    template: `
        <div class="row">
            <div class="col-xs-12">
                <div class="text-right" role="group">
                    <button class="btn btn-link" (click)="delete(project)"><span class="fa fa-trash fa-size-x2"></span></button>
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
            <tr *ngIf="project.pathServer != ''">
                <th><span class="fa fa-fw fa-folder-open-o"></span>&nbsp;<span>Server path</span></th>
                <td>{{ project.pathServer }}</td>
            </tr>
            <!--
            <tr>
                <th><span class="fa fa-fw fa-status"></span>&nbsp;<span>Status</span></th>
                <td>{{ project.remotePrjDef.status }}</td>
            </tr>
            -->
            </tbody>
        </table >
    `,
    styleUrls: ['./app/config/config.component.css']
})

export class ProjectCardComponent {

    @Input() project: IProject;

    constructor(private configSvr: ConfigService) {
    }


    delete(prj: IProject) {
        this.configSvr.deleteProject(prj);
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
