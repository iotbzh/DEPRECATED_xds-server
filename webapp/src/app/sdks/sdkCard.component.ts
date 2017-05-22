import { Component, Input } from '@angular/core';
import { ISdk } from "../services/sdk.service";

@Component({
    selector: 'sdk-card',
    template: `
        <div class="row">
            <div class="col-xs-12">
                <div class="text-right" role="group">
                    <button disabled class="btn btn-link" (click)="delete(sdk)"><span class="fa fa-trash fa-size-x2"></span></button>
                </div>
            </div>
        </div>

        <table class="table table-striped">
            <tbody>
            <tr>
                <th><span class="fa fa-fw fa-id-badge"></span>&nbsp;<span>Profile</span></th>
                <td>{{ sdk.profile }}</td>
            </tr>
            <tr>
                <th><span class="fa fa-fw fa-tasks"></span>&nbsp;<span>Architecture</span></th>
                <td>{{ sdk.arch }}</td>
            </tr>
            <tr>
                <th><span class="fa fa-fw fa-code-fork"></span>&nbsp;<span>Version</span></th>
                <td>{{ sdk.version }}</td>
            </tr>
            <tr>
                <th><span class="fa fa-fw fa-folder-open-o"></span>&nbsp;<span>Sdk path</span></th>
                <td>{{ sdk.path}}</td>
            </tr>

            </tbody>
        </table >
    `,
    styleUrls: ['./app/config/config.component.css']
})

export class SdkCardComponent {

    @Input() sdk: ISdk;

    constructor() { }


    delete(sdk: ISdk) {
        // Not supported for now
    }

}
