import { Component, Input } from "@angular/core";

import { IProject } from "../common/config.service";

@Component({
    selector: 'projects-list-accordion',
    template: `
        <accordion>
            <accordion-group #group *ngFor="let prj of projects">
                <div accordion-heading>
                    {{ prj.label }}
                    <i class="pull-right float-xs-right fa"
                    [ngClass]="{'fa-chevron-down': group.isOpen, 'fa-chevron-right': !group.isOpen}"></i>
                </div>
                <project-card [project]="prj"></project-card>
            </accordion-group>
        </accordion>
    `
})
export class ProjectsListAccordionComponent {

    @Input() projects: IProject[];

}


