import { Component, Input } from "@angular/core";

import { ISdk, SdkService } from "../common/sdk.service";

@Component({
    selector: 'sdk-select-dropdown',
    template: `
        <div class="btn-group" dropdown *ngIf="curSdk" >
            <button dropdownToggle type="button" class="btn btn-primary dropdown-toggle" style="width: 20em;">
                {{curSdk.name}} <span class="caret" style="float: right; margin-top: 8px;"></span>
            </button>
            <ul *dropdownMenu class="dropdown-menu" role="menu">
                <li role="menuitem"><a class="dropdown-item" *ngFor="let sdk of sdks" (click)="select(sdk)">
                    {{sdk.name}}</a>
                </li>
            </ul>
        </div>
    `
})
export class SdkSelectDropdownComponent {

    // FIXME investigate to understand why not working with sdks as input
    // <sdk-select-dropdown [sdks]="(sdks$ | async)"></sdk-select-dropdown>
    //@Input() sdks: ISdk[];
    sdks: ISdk[];

    curSdk: ISdk;

    constructor(private sdkSvr: SdkService) { }

    ngOnInit() {
        this.sdkSvr.Sdks$.subscribe((s) => {
            this.sdks = s;
            this.curSdk = this.sdks.length ? this.sdks[0] : null;
            this.sdkSvr.setCurrent(this.curSdk);
        });
    }

    select(s) {
         this.sdkSvr.setCurrent(this.curSdk = s);
    }
}


