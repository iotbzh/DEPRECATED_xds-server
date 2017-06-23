import { Injectable } from '@angular/core';

@Injectable()
export class UtilsService {
    constructor() { }

    getOSName(lowerCase?: boolean): string {
        var checkField = function (ff) {
            if (ff.indexOf("Linux") !== -1) {
                return "Linux";
            } else if (ff.indexOf("Win") !== -1) {
                return "Windows";
            } else if (ff.indexOf("Mac") !== -1) {
                return "MacOS";
            } else if (ff.indexOf("X11") !== -1) {
                return "UNIX";
            }
            return "";
        };

        let OSName = checkField(navigator.platform);
        if (OSName === "") {
            OSName = checkField(navigator.appVersion);
        }
        if (OSName === "") {
            OSName = "Unknown OS";
        }
        if (lowerCase) {
            return OSName.toLowerCase();
        }
        return OSName;
    }
}