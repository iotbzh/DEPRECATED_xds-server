import { Injectable } from '@angular/core';

@Injectable()
export class UtilsService {
    constructor() { }

    getOSName(lowerCase?: boolean): string {
        let OSName = "Unknown OS";
        if (navigator.appVersion.indexOf("Linux") !== -1) {
            OSName = "Linux";
        } else if (navigator.appVersion.indexOf("Win") !== -1) {
            OSName = "Windows";
        } else if (navigator.appVersion.indexOf("Mac") !== -1) {
            OSName = "MacOS";
        } else if (navigator.appVersion.indexOf("X11") !== -1) {
            OSName = "UNIX";
        }
        if (lowerCase) {
            return OSName.toLowerCase();
        }
        return OSName;
    }
}