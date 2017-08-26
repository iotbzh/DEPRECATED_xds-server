import { Injectable, SecurityContext } from '@angular/core';
import { DomSanitizer } from '@angular/platform-browser';
import { Observable } from 'rxjs/Observable';
import { Subject } from 'rxjs/Subject';


export type AlertType = "danger" | "warning" | "info" | "success";

export interface IAlert {
    type: AlertType;
    msg: string;
    show?: boolean;
    dismissible?: boolean;
    dismissTimeout?: number;     // close alert after this time (in seconds)
    id?: number;
}

@Injectable()
export class AlertService {
    public alerts: Observable<IAlert[]>;

    private _alerts: IAlert[];
    private alertsSubject = <Subject<IAlert[]>>new Subject();
    private uid = 0;
    private defaultDissmissTmo = 5; // in seconds

    constructor(private sanitizer: DomSanitizer) {
        this.alerts = this.alertsSubject.asObservable();
        this._alerts = [];
        this.uid = 0;
    }

    public error(msg: string, dismissTime?: number) {
        this.add({
            type: "danger", msg: msg, dismissible: true, dismissTimeout: dismissTime
        });
    }

    public warning(msg: string, dismissible?: boolean) {
        this.add({ type: "warning", msg: msg, dismissible: true, dismissTimeout: (dismissible ? this.defaultDissmissTmo : 0) });
    }

    public info(msg: string) {
        this.add({ type: "info", msg: msg, dismissible: true, dismissTimeout: this.defaultDissmissTmo });
    }

    public add(al: IAlert) {
        this._alerts.push({
            show: true,
            type: al.type,
            msg: this.sanitizer.sanitize(SecurityContext.HTML, al.msg),
            dismissible: al.dismissible || true,
            dismissTimeout: (al.dismissTimeout * 1000) || 0,
            id: this.uid,
        });
        this.uid += 1;
        this.alertsSubject.next(this._alerts);
    }

    public del(al: IAlert) {
        let idx = this._alerts.findIndex((a) => a.id === al.id);
        if (idx > -1) {
            this._alerts.splice(idx, 1);
        }
    }
}
