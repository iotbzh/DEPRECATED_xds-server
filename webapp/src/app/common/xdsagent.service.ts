import { Injectable } from '@angular/core';
import { Http, Headers, RequestOptionsArgs, Response } from '@angular/http';
import { Location } from '@angular/common';
import { Observable } from 'rxjs/Observable';
import { Subject } from 'rxjs/Subject';
import { BehaviorSubject } from 'rxjs/BehaviorSubject';
import * as io from 'socket.io-client';

import { AlertService } from './alert.service';


// Import RxJs required methods
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/catch';
import 'rxjs/add/observable/throw';

export interface IXDSVersion {
    version: string;
    apiVersion: string;
    gitTag: string;

}

export interface IAgentStatus {
    baseURL: string;
    connected: boolean;
    WS_connected: boolean;
    connectionRetry: number;
    version: string;
}

// Default settings
const DEFAULT_PORT = 8010;
const DEFAULT_API_KEY = "1234abcezam";
const API_VERSION = "v1";

@Injectable()
export class XDSAgentService {
    public Status$: Observable<IAgentStatus>;

    private baseRestUrl: string;
    private wsUrl: string;
    private connectionMaxRetry: number;
    private apikey: string;
    private _status: IAgentStatus = {
        baseURL: "",
        connected: false,
        WS_connected: false,
        connectionRetry: 0,
        version: "",
    };
    private statusSubject = <BehaviorSubject<IAgentStatus>>new BehaviorSubject(this._status);


    private socket: SocketIOClient.Socket;

    constructor(private http: Http, private _window: Window, private alert: AlertService) {

        this.Status$ = this.statusSubject.asObservable();

        this.apikey = DEFAULT_API_KEY; // FIXME Add dynamic allocated key
        this._status.baseURL = 'http://localhost:' + DEFAULT_PORT;
        this.baseRestUrl = this._status.baseURL + '/api/' + API_VERSION;
        let re = this._window.location.origin.match(/http[s]?:\/\/([^\/]*)[\/]?/);
        if (re === null || re.length < 2) {
            console.error('ERROR: cannot determine Websocket url');
        } else {
            this.wsUrl = 'ws://' + re[1];
        }
    }

    connect(retry: number, url?: string): Observable<IAgentStatus> {
        if (url) {
            this._status.baseURL = url;
            this.baseRestUrl = this._status.baseURL + '/api/' + API_VERSION;
        }
        //FIXME [XDS-Agent]: not implemented yet, set always as connected
        //this._status.connected = false;
        this._status.connected = true;
        this._status.connectionRetry = 0;
        this.connectionMaxRetry = retry || 3600;   // 1 hour

        // Init IO Socket connection
        this._handleIoSocket();

        // Get Version in order to check connection via a REST request
        return this.getVersion()
            .map((v) => {
                this._status.version = v.version;
                this.statusSubject.next(Object.assign({}, this._status));
                return this._status;
            });
    }

    public getVersion(): Observable<IXDSVersion> {
        /*FIXME [XDS-Agent]: Not implemented for now
        return this._get('/version');
        */
        return Observable.of({
            version: "NOT_IMPLEMENTED",
            apiVersion: "NOT_IMPLEMENTED",
            gitTag: "NOT_IMPLEMENTED"
        });
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

    }

    private _attachAuthHeaders(options?: any) {
        options = options || {};
        let headers = options.headers || new Headers();
        // headers.append('Authorization', 'Basic ' + btoa('username:password'));
        headers.append('Access-Control-Allow-Origin', '*');
        headers.append('Accept', 'application/json');
        headers.append('Content-Type', 'application/json');
        if (this.apikey !== "") {
            headers.append('X-API-Key', this.apikey);

        }

        options.headers = headers;
        return options;
    }

    private _checkAlive(): Observable<boolean> {
        if (this._status.connected) {
            return Observable.of(true);
        }

        return this.http.get(this._status.baseURL, this._attachAuthHeaders())
            .map((r) => this._status.connected = true)
            .retryWhen((attempts) => {
                this._status.connectionRetry = 0;
                return attempts.flatMap(error => {
                    this._status.connected = false;
                    if (++this._status.connectionRetry >= this.connectionMaxRetry) {
                        return Observable.throw("XDS local Agent not responding (url=" + this._status.baseURL + ")");
                    } else {
                        return Observable.timer(1000);
                    }
                });
            });
    }

    private _get(url: string): Observable<any> {
        return this._checkAlive()
            .flatMap(() => this.http.get(this.baseRestUrl + url, this._attachAuthHeaders()))
            .map((res: Response) => res.json())
            .catch(this._decodeError);
    }
    private _post(url: string, body: any): Observable<any> {
        return this._checkAlive()
            .flatMap(() => this.http.post(this.baseRestUrl + url, JSON.stringify(body), this._attachAuthHeaders()))
            .map((res: Response) => res.json())
            .catch((error) => {
                return this._decodeError(error);
            });
    }
    private _delete(url: string): Observable<any> {
        return this._checkAlive()
            .flatMap(() => this.http.delete(this.baseRestUrl + url, this._attachAuthHeaders()))
            .map((res: Response) => res.json())
            .catch(this._decodeError);
    }

    private _decodeError(err: Response | any) {
        let e: string;
        if (this._status) {
            this._status.connected = false;
        }
        if (typeof err === "object") {
            if (err.statusText) {
                e = err.statusText;
            } else if (err.error) {
                e = String(err.error);
            } else {
                e = JSON.stringify(err);
            }
        } else if (err instanceof Response) {
            const body = err.json() || 'Server error';
            const error = body.error || JSON.stringify(body);
            e = `${err.status} - ${err.statusText || ''} ${error}`;
        } else {
            e = err.message ? err.message : err.toString();
        }
        return Observable.throw(e);
    }
}
