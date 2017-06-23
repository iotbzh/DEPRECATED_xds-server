import { Component, Input, Pipe, PipeTransform } from '@angular/core';

import { IxdsAgentPackage } from "../services/config.service";

@Component({
    selector: 'dl-xds-agent',
    template: `
        <template #popTemplate>
            <h3>Download xds-agent packages:</h3>
            <ul>
                <li *ngFor="let p of packageUrls">
                    <a href="{{p.url}}">{{p.os | capitalize}} - {{p.arch}} ({{p.version}}) </a>
                </li>
            </ul>
            <button type="button" class="btn btn-sm" (click)="pop.hide()"> Cancel </button>
        </template>
        <button type="button" class="btn btn-link fa fa-download fa-size-x2"
            [popover]="popTemplate"
            #pop="bs-popover"
            placement="left">
        </button>
    `,
    styles: [`
        .fa-size-x2 {
            font-size: 20px;
        }
    `]
})

export class DlXdsAgentComponent {

    @Input() packageUrls: IxdsAgentPackage[];

}

@Pipe({
    name: 'capitalize'
})
export class CapitalizePipe implements PipeTransform {
    transform(value: string): string {
        if (value) {
            return value.charAt(0).toUpperCase() + value.slice(1);
        }
        return value;
    }
}
