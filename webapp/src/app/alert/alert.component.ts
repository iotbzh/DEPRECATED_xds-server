import { Component } from '@angular/core';
import { Observable } from 'rxjs';

import {AlertService, IAlert} from '../services/alert.service';

@Component({
    selector: 'app-alert',
    template: `
        <div style="width:80%; margin-left:auto; margin-right:auto;" *ngFor="let alert of (alerts$ | async)">
            <alert *ngIf="alert.show" [type]="alert.type" [dismissible]="alert.dismissible" [dismissOnTimeout]="alert.dismissTimeout"
            (onClose)="onClose(alert)">
                <div style="text-align:center;" [innerHtml]="alert.msg"></div>
            </alert>
        </div>
    `
})

export class AlertComponent {

    alerts$: Observable<IAlert[]>;

    constructor(private alertSvr: AlertService) {
        this.alerts$ = this.alertSvr.alerts;
    }

    onClose(al) {
        this.alertSvr.del(al);
    }

}
