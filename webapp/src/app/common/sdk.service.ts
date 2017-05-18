import { Injectable, SecurityContext } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import { BehaviorSubject } from 'rxjs/BehaviorSubject';

import { XDSServerService } from "../common/xdsserver.service";

export interface ISdk {
    id: string;
    profile: string;
    version: string;
    arch: number;
    path: string;
}

@Injectable()
export class SdkService {
    public Sdks$: Observable<ISdk[]>;

    private _sdksList = [];
    private current: ISdk;
    private sdksSubject = <BehaviorSubject<ISdk[]>>new BehaviorSubject(this._sdksList);

    constructor(private xdsSvr: XDSServerService) {
        this.current = null;
        this.Sdks$ = this.sdksSubject.asObservable();

        this.xdsSvr.getSdks().subscribe((s) => {
            this._sdksList = s;
            this.sdksSubject.next(s);
        });
    }

    public setCurrent(s: ISdk) {
        this.current = s;
    }

    public getCurrent(): ISdk {
        return this.current;
    }

    public getCurrentId(): string {
        if (this.current && this.current.id) {
            return this.current.id;
        }
        return "";
    }
}