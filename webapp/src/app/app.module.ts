import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { HttpModule } from "@angular/http";
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { CookieModule } from 'ngx-cookie';

// Import bootstrap
import { AlertModule } from 'ngx-bootstrap/alert';
import { ModalModule } from 'ngx-bootstrap/modal';
import { AccordionModule } from 'ngx-bootstrap/accordion';
import { CarouselModule } from 'ngx-bootstrap/carousel';
import { BsDropdownModule } from 'ngx-bootstrap/dropdown';

// Import the application components and services.
import { Routing, AppRoutingProviders } from './app.routing';
import { AppComponent } from "./app.component";
import { AlertComponent } from './alert/alert.component';
import { ConfigComponent } from "./config/config.component";
import { ProjectCardComponent } from "./projects/projectCard.component";
import { ProjectReadableTypePipe } from "./projects/projectCard.component";
import { ProjectsListAccordionComponent } from "./projects/projectsListAccordion.component";
import { SdkCardComponent } from "./sdks/sdkCard.component";
import { SdksListAccordionComponent } from "./sdks/sdksListAccordion.component";
import { SdkSelectDropdownComponent } from "./sdks/sdkSelectDropdown.component";

import { HomeComponent } from "./home/home.component";
import { DevelComponent } from "./devel/devel.component";
import { BuildComponent } from "./devel/build/build.component";
import { DeployComponent } from "./devel/deploy/deploy.component";
import { XDSServerService } from "./services/xdsserver.service";
import { XDSAgentService } from "./services/xdsagent.service";
import { SyncthingService } from "./services/syncthing.service";
import { ConfigService } from "./services/config.service";
import { AlertService } from './services/alert.service';
import { UtilsService } from './services/utils.service';
import { SdkService } from "./services/sdk.service";



@NgModule({
    imports: [
        BrowserModule,
        HttpModule,
        FormsModule,
        ReactiveFormsModule,
        Routing,
        CookieModule.forRoot(),
        AlertModule.forRoot(),
        ModalModule.forRoot(),
        AccordionModule.forRoot(),
        CarouselModule.forRoot(),
        BsDropdownModule.forRoot(),
    ],
    declarations: [
        AppComponent,
        AlertComponent,
        HomeComponent,
        BuildComponent,
        DevelComponent,
        DeployComponent,
        ConfigComponent,
        ProjectCardComponent,
        ProjectReadableTypePipe,
        ProjectsListAccordionComponent,
        SdkCardComponent,
        SdksListAccordionComponent,
        SdkSelectDropdownComponent,
    ],
    providers: [
        AppRoutingProviders,
        {
            provide: Window,
            useValue: window
        },
        XDSServerService,
        XDSAgentService,
        ConfigService,
        SyncthingService,
        AlertService,
        UtilsService,
        SdkService,
    ],
    bootstrap: [AppComponent]
})
export class AppModule {
}