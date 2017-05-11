import { Component, OnInit, OnDestroy } from "@angular/core";
import { Router } from '@angular/router';
//TODO import {TranslateService} from "ng2-translate";

@Component({
    selector: 'app',
    templateUrl: './app/app.component.html',
    styleUrls: ['./app/app.component.css']
})

export class AppComponent implements OnInit, OnDestroy {
    private defaultLanguage: string = 'en';

    // I initialize the app component.
    //TODO constructor(private translate: TranslateService) {
    constructor(public router: Router) {
    }

    ngOnInit() {

        /* TODO
        this.translate.addLangs(["en", "fr"]);
        this.translate.setDefaultLang(this.defaultLanguage);

        let browserLang = this.translate.getBrowserLang();
        this.translate.use(browserLang.match(/en|fr/) ? browserLang : this.defaultLanguage);
        */
    }

    ngOnDestroy(): void {
    }


}
