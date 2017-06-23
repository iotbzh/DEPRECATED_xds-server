import { Injectable } from '@angular/core';
import { Http, Headers, RequestOptionsArgs, Response } from '@angular/http';
import { Location } from '@angular/common';
import { Observable } from 'rxjs/Observable';
import { Subject } from 'rxjs/Subject';
import { BehaviorSubject } from 'rxjs/BehaviorSubject';
import * as io from 'socket.io-client';

import { AlertService } from './alert.service';
import { ISdk } from './sdk.service';


// Import RxJs required methods
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/catch';
import 'rxjs/add/observable/throw';
import 'rxjs/add/operator/mergeMap';


export interface IXDSConfigProject {
    id: string;
    path: string;
    hostSyncThingID: string;
    label?: string;
    defaultSdkID?: string;
}

interface IXDSBuilderConfig {
    ip: string;
    port: string;
    syncThingID: string;
}

interface IXDSFolderConfig {
    id: string;
    label: string;
    path: string;
    type: number;
    syncThingID: string;
    builderSThgID?: string;
    status?: string;
    defaultSdkID: string;
}

interface IXDSConfig {
    version: number;
    builder: IXDSBuilderConfig;
    folders: IXDSFolderConfig[];
}

export interface IXDSAgentTarball {
    os: string;
    arch: string;
    version: string;
    rawVersion: string;
    fileUrl: string;
}

export interface IXDSAgentInfo {
    tarballs: IXDSAgentTarball[];
}

export interface ISdkMessage {
    wsID: string;
    msgType: string;
    data: any;
}

export interface ICmdOutput {
    cmdID: string;
    timestamp: string;
    stdout: string;
    stderr: string;
}

export interface ICmdExit {
    cmdID: string;
    timestamp: string;
    code: number;
    error: string;
}

export interface IServerStatus {
    WS_connected: boolean;

}

const FOLDER_TYPE_CLOUDSYNC = 2;

@Injectable()
export class XDSServerService {

    public CmdOutput$ = <Subject<ICmdOutput>>new Subject();
    public CmdExit$ = <Subject<ICmdExit>>new Subject();
    public Status$: Observable<IServerStatus>;

    private baseUrl: string;
    private wsUrl: string;
    private _status = { WS_connected: false };
    private statusSubject = <BehaviorSubject<IServerStatus>>new BehaviorSubject(this._status);


    private socket: SocketIOClient.Socket;

    constructor(private http: Http, private _window: Window, private alert: AlertService) {

        this.Status$ = this.statusSubject.asObservable();

        this.baseUrl = this._window.location.origin + '/api/v1';
        let re = this._window.location.origin.match(/http[s]?:\/\/([^\/]*)[\/]?/);
        if (re === null || re.length < 2) {
            console.error('ERROR: cannot determine Websocket url');
        } else {
            this.wsUrl = 'ws://' + re[1];
            this._handleIoSocket();
        }
    }

    private _WSState(sts: boolean) {
        this._status.WS_connected = sts;
        this.statusSubject.next(Object.assign({}, this._status));
    }

    private _handleIoSocket() {
        this.socket = io(this.wsUrl, { transports: ['websocket'] });

        this.socket.on('connect_error', (res) => {
            this._WSState(false);
            console.error('WS Connect_error ', res);
        });

        this.socket.on('connect', (res) => {
            this._WSState(true);
        });

        this.socket.on('disconnection', (res) => {
            this._WSState(false);
            this.alert.error('WS disconnection: ' + res);
        });

        this.socket.on('error', (err) => {
            console.error('WS error:', err);
        });

        this.socket.on('make:output', data => {
            this.CmdOutput$.next(Object.assign({}, <ICmdOutput>data));
        });

        this.socket.on('make:exit', data => {
            this.CmdExit$.next(Object.assign({}, <ICmdExit>data));
        });

        this.socket.on('exec:output', data => {
            this.CmdOutput$.next(Object.assign({}, <ICmdOutput>data));
        });

        this.socket.on('exec:exit', data => {
            this.CmdExit$.next(Object.assign({}, <ICmdExit>data));
        });

    }

    getSdks(): Observable<ISdk[]> {
        return this._get('/sdks');
    }

    getXdsAgentInfo(): Observable<IXDSAgentInfo> {
        return this._get('/xdsagent/info');
    }

    getProjects(): Observable<IXDSFolderConfig[]> {
        return this._get('/folders');
    }

    addProject(cfg: IXDSConfigProject): Observable<IXDSFolderConfig> {
        let folder: IXDSFolderConfig = {
            id: cfg.id || null,
            label: cfg.label || "",
            path: cfg.path,
            type: FOLDER_TYPE_CLOUDSYNC,
            syncThingID: cfg.hostSyncThingID,
            defaultSdkID: cfg.defaultSdkID || "",
        };
        return this._post('/folder', folder);
    }

    deleteProject(id: string): Observable<IXDSFolderConfig> {
        return this._delete('/folder/' + id);
    }

    exec(prjID: string, dir: string, cmd: string, sdkid?: string, args?: string[], env?: string[]): Observable<any> {
        return this._post('/exec',
            {
                id: prjID,
                rpath: dir,
                cmd: cmd,
                sdkid: sdkid || "",
                args: args || [],
                env: env || [],
            });
    }

    make(prjID: string, dir: string, sdkid?: string, args?: string[], env?: string[]): Observable<any> {
        return this._post('/make',
            {
                id: prjID,
                rpath: dir,
                sdkid: sdkid,
                args: args || [],
                env: env || [],
            });
    }


    private _attachAuthHeaders(options?: any) {
        options = options || {};
        let headers = options.headers || new Headers();
        // headers.append('Authorization', 'Basic ' + btoa('username:password'));
        headers.append('Accept', 'application/json');
        headers.append('Content-Type', 'application/json');
        // headers.append('Access-Control-Allow-Origin', '*');

        options.headers = headers;
        return options;
    }

    private _get(url: string): Observable<any> {
        return this.http.get(this.baseUrl + url, this._attachAuthHeaders())
            .map((res: Response) => res.json())
            .catch(this._decodeError);
    }
    private _post(url: string, body: any): Observable<any> {
        return this.http.post(this.baseUrl + url, JSON.stringify(body), this._attachAuthHeaders())
            .map((res: Response) => res.json())
            .catch((error) => {
                return this._decodeError(error);
            });
    }
    private _delete(url: string): Observable<any> {
        return this.http.delete(this.baseUrl + url, this._attachAuthHeaders())
            .map((res: Response) => res.json())
            .catch(this._decodeError);
    }

    private _decodeError(err: any) {
        let e: string;
        if (typeof err === "object") {
            if (err.statusText) {
                e = err.statusText;
            } else if (err.error) {
                e = String(err.error);
            } else {
                e = JSON.stringify(err);
            }
        } else {
            e = err.json().error || 'Server error';
        }
        return Observable.throw(e);
    }
}
