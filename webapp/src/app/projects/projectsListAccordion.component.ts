import { Component, Input } from "@angular/core";

import { IProject } from "../services/config.service";

@Component({
    selector: 'projects-list-accordion',
    template: `
        <style>
            .fa.fa-exclamation-triangle {
                margin-right: 2em;
                color: red;
            }
            .fa.fa-refresh {
                margin-right: 10px;
                color: darkviolet;
            }
        </style>
        <accordion>
            <accordion-group #group *ngFor="let prj of projects">
                <div accordion-heading>
                    {{ prj.label }}
                    <div class="pull-right">
                        <i *ngIf="prj.status == 'Syncing'" class="fa fa-refresh faa-spin animated"></i>
                        <i *ngIf="!prj.isInSync && prj.status != 'Syncing'" class="fa fa-exclamation-triangle"></i>
                        <i class="fa" [ngClass]="{'fa-chevron-down': group.isOpen, 'fa-chevron-right': !group.isOpen}"></i>
                    </div>
                </div>
                <project-card [project]="prj"></project-card>
            </accordion-group>
        </accordion>
    `
})
export class ProjectsListAccordionComponent {

    @Input() projects: IProject[];

}


