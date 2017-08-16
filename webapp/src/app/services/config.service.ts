import { Injectable, OnInit } from '@angular/core';
import { Http, Headers, RequestOptionsArgs, Response } from '@angular/http';
import { Location } from '@angular/common';
import { CookieService } from 'ngx-cookie';
import { Observable } from 'rxjs/Observable';
import { Subscriber } from 'rxjs/Subscriber';
import { BehaviorSubject } from 'rxjs/BehaviorSubject';

// Import RxJs required methods
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/catch';
import 'rxjs/add/observable/throw';
import 'rxjs/add/operator/mergeMap';


import { XDSServerService, IXDSFolderConfig } from "../services/xdsserver.service";
import { XDSAgentService } from "../services/xdsagent.service";
import { SyncthingService, ISyncThingProject, ISyncThingStatus } from "../services/syncthing.service";
import { AlertService, IAlert } from "../services/alert.service";
import { UtilsService } from "../services/utils.service";

export enum ProjectType {
    NATIVE_PATHMAP = 1,
    SYNCTHING = 2
}

export var ProjectTypes = [
    { value: ProjectType.NATIVE_PATHMAP, display: "Path mapping" },
    { value: ProjectType.SYNCTHING, display: "Cloud Sync" }
];

export interface INativeProject {
    // TODO
}

export interface IProject {
    id?: string;
    label: string;
    pathClient: string;
    pathServer?: string;
    type: ProjectType;
    remotePrjDef?: INativeProject | ISyncThingProject;
    localPrjDef?: any;
    isExpanded?: boolean;
    visible?: boolean;
    defaultSdkID?: string;
}

export interface IXDSAgentConfig {
    URL: string;
    retry: number;
}

export interface ILocalSTConfig {
    ID: string;
    URL: string;
    retry: number;
    tilde: string;
}

export interface IxdsAgentPackage {
    os: string;
    arch: string;
    version: string;
    url: string;
}

export interface IConfig {
    xdsServerURL: string;
    xdsAgent: IXDSAgentConfig;
    xdsAgentPackages: IxdsAgentPackage[];
    projectsRootDir: string;
    projects: IProject[];
    localSThg: ILocalSTConfig;
}

@Injectable()
export class ConfigService {

    public conf: Observable<IConfig>;

    private confSubject: BehaviorSubject<IConfig>;
    private confStore: IConfig;
    private AgentConnectObs = null;
    private stConnectObs = null;

    constructor(private _window: Window,
        private cookie: CookieService,
        private xdsServerSvr: XDSServerService,
        private xdsAgentSvr: XDSAgentService,
        private stSvr: SyncthingService,
        private alert: AlertService,
        private utils: UtilsService,
    ) {
        this.load();
        this.confSubject = <BehaviorSubject<IConfig>>new BehaviorSubject(this.confStore);
        this.conf = this.confSubject.asObservable();

        // force to load projects
        this.loadProjects();
    }

    // Load config
    load() {
        // Try to retrieve previous config from cookie
        let cookConf = this.cookie.getObject("xds-config");
        if (cookConf != null) {
            this.confStore = <IConfig>cookConf;
        } else {
            // Set default config
            this.confStore = {
                xdsServerURL: this._window.location.origin + '/api/v1',
                xdsAgent: {
                    URL: 'http://localhost:8000',
                    retry: 10,
                },
                xdsAgentPackages: [],
                projectsRootDir: "",
                projects: [],
                localSThg: {
                    ID: null,
                    URL: "http://localhost:8384",
                    retry: 10,    // 10 seconds
                    tilde: "",
                }
            };
        }

        // Update XDS Agent tarball url
        this.xdsServerSvr.getXdsAgentInfo().subscribe(nfo => {
            this.confStore.xdsAgentPackages = [];
            nfo.tarballs && nfo.tarballs.forEach(el =>
                this.confStore.xdsAgentPackages.push({
                    os: el.os,
                    arch: el.arch,
                    version: el.version,
                    url: el.fileUrl
                })
            );
            this.confSubject.next(Object.assign({}, this.confStore));
        });
    }

    // Save config into cookie
    save() {
        // Notify subscribers
        this.confSubject.next(Object.assign({}, this.confStore));

        // Don't save projects in cookies (too big!)
        let cfg = Object.assign({}, this.confStore);
        delete (cfg.projects);
        this.cookie.putObject("xds-config", cfg);
    }

    loadProjects() {
        // Setup connection with local XDS agent
        if (this.AgentConnectObs) {
            try {
                this.AgentConnectObs.unsubscribe();
            } catch (err) { }
            this.AgentConnectObs = null;
        }

        let cfg = this.confStore.xdsAgent;
        this.AgentConnectObs = this.xdsAgentSvr.connect(cfg.retry, cfg.URL)
            .subscribe((sts) => {
                //console.log("Agent sts", sts);
                // FIXME: load projects from local XDS Agent and
                //  not directly from local syncthing
                this._loadProjectFromLocalST();

            }, error => {
                if (error.indexOf("XDS local Agent not responding") !== -1) {
                    let msg = "<span><strong>" + error + "<br></strong>";
                    msg += "You may need to download and execute XDS-Agent.<br>";

                    let os = this.utils.getOSName(true);
                    let zurl = this.confStore.xdsAgentPackages && this.confStore.xdsAgentPackages.filter(elem => elem.os === os);
                    if (zurl && zurl.length) {
                        msg += " Download XDS-Agent tarball for " + zurl[0].os + " host OS ";
                        msg += "<a class=\"fa fa-download\" href=\"" + zurl[0].url + "\" target=\"_blank\"></a>";
                    }
                    msg += "</span>";
                    this.alert.error(msg);
                } else {
                    this.alert.error(error);
                }
            });
    }

    private _loadProjectFromLocalST() {
        // Remove previous subscriber if existing
        if (this.stConnectObs) {
            try {
                this.stConnectObs.unsubscribe();
            } catch (err) { }
            this.stConnectObs = null;
        }

        // FIXME: move this code and all logic about syncthing inside XDS Agent
        // Setup connection with local SyncThing
        let retry = this.confStore.localSThg.retry;
        let url = this.confStore.localSThg.URL;
        this.stConnectObs = this.stSvr.connect(retry, url).subscribe((sts) => {
            this.confStore.localSThg.ID = sts.ID;
            this.confStore.localSThg.tilde = sts.tilde;
            if (this.confStore.projectsRootDir === "") {
                this.confStore.projectsRootDir = sts.tilde;
            }

            // Rebuild projects definition from local and remote syncthing
            this.confStore.projects = [];

            this.xdsServerSvr.getProjects().subscribe(remotePrj => {
                this.stSvr.getProjects().subscribe(localPrj => {
                    remotePrj.forEach(rPrj => {
                        let lPrj = localPrj.filter(item => item.id === rPrj.id);
                        if (lPrj.length > 0) {
                            let pp: IProject = {
                                id: rPrj.id,
                                label: rPrj.label,
                                pathClient: rPrj.path,
                                pathServer: rPrj.dataPathMap.serverPath,
                                type: rPrj.type,
                                remotePrjDef: Object.assign({}, rPrj),
                                localPrjDef: Object.assign({}, lPrj[0]),
                            };
                            this.confStore.projects.push(pp);
                        }
                    });
                    this.confSubject.next(Object.assign({}, this.confStore));
                }), error => this.alert.error('Could not load initial state of local projects.');
            }), error => this.alert.error('Could not load initial state of remote projects.');

        }, error => {
            if (error.indexOf("Syncthing local daemon not responding") !== -1) {
                let msg = "<span><strong>" + error + "<br></strong>";
                msg += "Please check that local XDS-Agent is running.<br>";
                msg += "</span>";
                this.alert.error(msg);
            } else {
                this.alert.error(error);
            }
        });
    }

    set syncToolURL(url: string) {
        this.confStore.localSThg.URL = url;
        this.save();
    }

    set xdsAgentRetry(r: number) {
        this.confStore.localSThg.retry = r;
        this.confStore.xdsAgent.retry = r;
        this.save();
    }

    set xdsAgentUrl(url: string) {
        this.confStore.xdsAgent.URL = url;
        this.save();
    }


    set projectsRootDir(p: string) {
        if (p.charAt(0) === '~') {
            p = this.confStore.localSThg.tilde + p.substring(1);
        }
        this.confStore.projectsRootDir = p;
        this.save();
    }

    getLabelRootName(): string {
        let id = this.confStore.localSThg.ID;
        if (!id || id === "") {
            return null;
        }
        return id.slice(0, 15);
    }

    addProject(prj: IProject): Observable<IProject> {
        // Substitute tilde with to user home path
        let pathCli = prj.pathClient.trim();
        if (pathCli.charAt(0) === '~') {
            pathCli = this.confStore.localSThg.tilde + pathCli.substring(1);

            // Must be a full path (on Linux or Windows)
        } else if (!((pathCli.charAt(0) === '/') ||
            (pathCli.charAt(1) === ':' && (pathCli.charAt(2) === '\\' || pathCli.charAt(2) === '/')))) {
            pathCli = this.confStore.projectsRootDir + '/' + pathCli;
        }

        let xdsPrj: IXDSFolderConfig = {
            id: "",
            label: prj.label || "",
            path: pathCli,
            type: prj.type,
            defaultSdkID: prj.defaultSdkID,
            dataPathMap: {
                serverPath: prj.pathServer,
            },
            dataCloudSync: {
                syncThingID: this.confStore.localSThg.ID,
            }
        };
        // Send config to XDS server
        let newPrj = prj;
        return this.xdsServerSvr.addProject(xdsPrj)
            .flatMap(resStRemotePrj => {
                newPrj.remotePrjDef = resStRemotePrj;
                newPrj.id = resStRemotePrj.id;
                newPrj.pathClient = resStRemotePrj.path;

                if (newPrj.type === ProjectType.SYNCTHING) {
                    // FIXME REWORK local ST config
                    //  move logic to server side tunneling-back by WS
                    let stData = resStRemotePrj.dataCloudSync;

                    // Now setup local config
                    let stLocPrj: ISyncThingProject = {
                        id: resStRemotePrj.id,
                        label: xdsPrj.label,
                        path: xdsPrj.path,
                        serverSyncThingID: stData.builderSThgID
                    };

                    // Set local Syncthing config
                    return this.stSvr.addProject(stLocPrj);

                } else {
                    newPrj.pathServer = resStRemotePrj.dataPathMap.serverPath;
                    return Observable.of(null);
                }
            })
            .map(resStLocalPrj => {
                newPrj.localPrjDef = resStLocalPrj;

                // FIXME: maybe reduce subject to only .project
                //this.confSubject.next(Object.assign({}, this.confStore).project);
                this.confStore.projects.push(Object.assign({}, newPrj));
                this.confSubject.next(Object.assign({}, this.confStore));

                return newPrj;
            });
    }

    deleteProject(prj: IProject): Observable<IProject> {
        let idx = this._getProjectIdx(prj.id);
        let delPrj = prj;
        if (idx === -1) {
            throw new Error("Invalid project id (id=" + prj.id + ")");
        }
        return this.xdsServerSvr.deleteProject(prj.id)
            .flatMap(res => {
                return this.stSvr.deleteProject(prj.id);
            })
            .map(res => {
                this.confStore.projects.splice(idx, 1);
                return delPrj;
            });
    }

    private _getProjectIdx(id: string): number {
        return this.confStore.projects.findIndex((item) => item.id === id);
    }

}
