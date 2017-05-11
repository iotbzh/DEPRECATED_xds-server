import { Injectable } from '@angular/core';
import { Http, Headers, RequestOptionsArgs, Response } from '@angular/http';
import { Location } from '@angular/common';
import { Observable } from 'rxjs/Observable';
import { BehaviorSubject } from 'rxjs/BehaviorSubject';

// Import RxJs required methods
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/catch';
import 'rxjs/add/observable/throw';
import 'rxjs/add/observable/of';
import 'rxjs/add/observable/timer';
import 'rxjs/add/operator/retryWhen';

export interface ISyncThingProject {
    id: string;
    path: string;
    remoteSyncThingID: string;
    label?: string;
}

export interface ISyncThingStatus {
    ID: string;
    baseURL: string;
    connected: boolean;
    connectionRetry: number;
    tilde: string;
    rawStatus: any;
}

// Private interfaces of Syncthing
const ISTCONFIG_VERSION = 19;

interface ISTFolderDeviceConfiguration {
    deviceID: string;
    introducedBy: string;
}
interface ISTFolderConfiguration {
    id: string;
    label: string;
    path: string;
    type?: number;
    devices?: ISTFolderDeviceConfiguration[];
    rescanIntervalS?: number;
    ignorePerms?: boolean;
    autoNormalize?: boolean;
    minDiskFreePct?: number;
    versioning?: { type: string; params: string[] };
    copiers?: number;
    pullers?: number;
    hashers?: number;
    order?: number;
    ignoreDelete?: boolean;
    scanProgressIntervalS?: number;
    pullerSleepS?: number;
    pullerPauseS?: number;
    maxConflicts?: number;
    disableSparseFiles?: boolean;
    disableTempIndexes?: boolean;
    fsync?: boolean;
    paused?: boolean;
}

interface ISTDeviceConfiguration {
    deviceID: string;
    name?: string;
    address?: string[];
    compression?: string;
    certName?: string;
    introducer?: boolean;
    skipIntroductionRemovals?: boolean;
    introducedBy?: string;
    paused?: boolean;
    allowedNetwork?: string[];
}

interface ISTGuiConfiguration {
    enabled: boolean;
    address: string;
    user?: string;
    password?: string;
    useTLS: boolean;
    apiKey?: string;
    insecureAdminAccess?: boolean;
    theme: string;
    debugging: boolean;
    insecureSkipHostcheck?: boolean;
}

interface ISTOptionsConfiguration {
    listenAddresses: string[];
    globalAnnounceServer: string[];
    // To be completed ...
}

interface ISTConfiguration {
    version: number;
    folders: ISTFolderConfiguration[];
    devices: ISTDeviceConfiguration[];
    gui: ISTGuiConfiguration;
    options: ISTOptionsConfiguration;
    ignoredDevices: string[];
}

// Default settings
const DEFAULT_GUI_PORT = 8384;
const DEFAULT_GUI_API_KEY = "1234abcezam";


@Injectable()
export class SyncthingService {

    public Status$: Observable<ISyncThingStatus>;

    private baseRestUrl: string;
    private apikey: string;
    private localSTID: string;
    private stCurVersion: number;
    private connectionMaxRetry: number;
    private _status: ISyncThingStatus = {
        ID: null,
        baseURL: "",
        connected: false,
        connectionRetry: 0,
        tilde: "",
        rawStatus: null,
    };
    private statusSubject = <BehaviorSubject<ISyncThingStatus>>new BehaviorSubject(this._status);

    constructor(private http: Http, private _window: Window) {
        this._status.baseURL = 'http://localhost:' + DEFAULT_GUI_PORT;
        this.baseRestUrl = this._status.baseURL + '/rest';
        this.apikey = DEFAULT_GUI_API_KEY;
        this.stCurVersion = -1;
        this.connectionMaxRetry = 10;   // 10 seconds

        this.Status$ = this.statusSubject.asObservable();
    }

    connect(retry: number, url?: string): Observable<ISyncThingStatus> {
        if (url) {
            this._status.baseURL = url;
            this.baseRestUrl = this._status.baseURL + '/rest';
        }
        this._status.connected = false;
        this._status.ID = null;
        this._status.connectionRetry = 0;
        this.connectionMaxRetry = retry || 3600;   // 1 hour
        return this.getStatus();
    }

    getID(): Observable<string> {
        if (this._status.ID != null) {
            return Observable.of(this._status.ID);
        }
        return this.getStatus().map(sts => sts.ID);
    }

    getStatus(): Observable<ISyncThingStatus> {
        return this._get('/system/status')
            .map((status) => {
                this._status.ID = status["myID"];
                this._status.tilde = status["tilde"];
                console.debug('ST local ID', this._status.ID);

                this._status.rawStatus = status;

                return this._status;
            });
    }

    getProjects(): Observable<ISTFolderConfiguration[]> {
        return this._getConfig()
            .map((conf) => conf.folders);
    }

    addProject(prj: ISyncThingProject): Observable<ISTFolderConfiguration> {
        return this.getID()
            .flatMap(() => this._getConfig())
            .flatMap((stCfg) => {
                let newDevID = prj.remoteSyncThingID;

                // Add new Device if needed
                let dev = stCfg.devices.filter(item => item.deviceID === newDevID);
                if (dev.length <= 0) {
                    stCfg.devices.push(
                        {
                            deviceID: newDevID,
                            name: "Builder_" + newDevID.slice(0, 15),
                            address: ["dynamic"],
                        }
                    );
                }

                // Add or update Folder settings
                let label = prj.label || "";
                let folder: ISTFolderConfiguration = {
                    id: prj.id,
                    label: label,
                    path: prj.path,
                    devices: [{ deviceID: newDevID, introducedBy: "" }],
                    autoNormalize: true,
                };

                let idx = stCfg.folders.findIndex(item => item.id === prj.id);
                if (idx === -1) {
                    stCfg.folders.push(folder);
                } else {
                    let newFld = Object.assign({}, stCfg.folders[idx], folder);
                    stCfg.folders[idx] = newFld;
                }

                // Set new config
                return this._setConfig(stCfg);
            })
            .flatMap(() => this._getConfig())
            .map((newConf) => {
                let idx = newConf.folders.findIndex(item => item.id === prj.id);
                return newConf.folders[idx];
            });
    }

    deleteProject(id: string): Observable<ISTFolderConfiguration> {
        let delPrj: ISTFolderConfiguration;
        return this._getConfig()
            .flatMap((conf: ISTConfiguration) => {
                let idx = conf.folders.findIndex(item => item.id === id);
                if (idx === -1) {
                    throw new Error("Cannot delete project: not found");
                }
                delPrj = Object.assign({}, conf.folders[idx]);
                conf.folders.splice(idx, 1);
                return this._setConfig(conf);
            })
            .map(() => delPrj);
    }

    /*
     * --- Private functions ---
     */
    private _getConfig(): Observable<ISTConfiguration> {
        return this._get('/system/config');
    }

    private _setConfig(cfg: ISTConfiguration): Observable<any> {
        return this._post('/system/config', cfg);
    }

    private _attachAuthHeaders(options?: any) {
        options = options || {};
        let headers = options.headers || new Headers();
        // headers.append('Authorization', 'Basic ' + btoa('username:password'));
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

        return this.http.get(this.baseRestUrl + '/system/version', this._attachAuthHeaders())
            .map((r) => this._status.connected = true)
            .retryWhen((attempts) => {
                this._status.connectionRetry = 0;
                return attempts.flatMap(error => {
                    this._status.connected = false;
                    if (++this._status.connectionRetry >= this.connectionMaxRetry) {
                        return Observable.throw("Syncthing local daemon not responding (url=" + this._status.baseURL + ")");
                    } else {
                        return Observable.timer(1000);
                    }
                });
            });
    }

    private _getAPIVersion(): Observable<number> {
        if (this.stCurVersion !== -1) {
            return Observable.of(this.stCurVersion);
        }

        return this.http.get(this.baseRestUrl + '/system/config', this._attachAuthHeaders())
            .map((res: Response) => {
                let conf: ISTConfiguration = res.json();
                this.stCurVersion = (conf && conf.version) || -1;
                return this.stCurVersion;
            })
            .catch(this._handleError);
    }

    private _checkAPIVersion(): Observable<number> {
        return this._getAPIVersion().map(ver => {
            if (ver !== ISTCONFIG_VERSION) {
                throw new Error("Unsupported Syncthing version api (" + ver +
                    " != " + ISTCONFIG_VERSION + ") !");
            }
            return ver;
        });
    }

    private _get(url: string): Observable<any> {
        return this._checkAlive()
            .flatMap(() => this._checkAPIVersion())
            .flatMap(() => this.http.get(this.baseRestUrl + url, this._attachAuthHeaders()))
            .map((res: Response) => res.json())
            .catch(this._handleError);
    }

    private _post(url: string, body: any): Observable<any> {
        return this._checkAlive()
            .flatMap(() => this._checkAPIVersion())
            .flatMap(() => this.http.post(this.baseRestUrl + url, JSON.stringify(body), this._attachAuthHeaders()))
            .map((res: Response) => {
                if (res && res.status && res.status === 200) {
                    return res;
                }
                throw new Error(res.toString());

            })
            .catch(this._handleError);
    }

    private _handleError(error: Response | any) {
        // In a real world app, you might use a remote logging infrastructure
        let errMsg: string;
        if (this._status) {
            this._status.connected = false;
        }
        if (error instanceof Response) {
            const body = error.json() || 'Server error';
            const err = body.error || JSON.stringify(body);
            errMsg = `${error.status} - ${error.statusText || ''} ${err}`;
        } else {
            errMsg = error.message ? error.message : error.toString();
        }
        return Observable.throw(errMsg);
    }
}
