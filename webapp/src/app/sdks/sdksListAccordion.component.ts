import { Component, Input } from "@angular/core";

import { ISdk } from "../common/sdk.service";

@Component({
    selector: 'sdks-list-accordion',
    template: `
        <accordion>
            <accordion-group #group *ngFor="let sdk of sdks">
                <div accordion-heading>
                    {{ sdk.name }}
                    <i class="pull-right float-xs-right fa"
                    [ngClass]="{'fa-chevron-down': group.isOpen, 'fa-chevron-right': !group.isOpen}"></i>
                </div>
                <sdk-card [sdk]="sdk"></sdk-card>
            </accordion-group>
        </accordion>
    `
})
export class SdksListAccordionComponent {

    @Input() sdks: ISdk[];

}


