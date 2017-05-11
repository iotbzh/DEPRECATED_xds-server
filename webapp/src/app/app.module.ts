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
import { HomeComponent } from "./home/home.component";
import { BuildComponent } from "./build/build.component";
import { XDSServerService } from "./common/xdsserver.service";
import { SyncthingService } from "./common/syncthing.service";
import { ConfigService } from "./common/config.service";
import { AlertService } from './common/alert.service';



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
        ConfigComponent,
        ProjectCardComponent,
        ProjectReadableTypePipe,
        ProjectsListAccordionComponent,
    ],
    providers: [
        AppRoutingProviders,
        {
            provide: Window,
            useValue: window
        },
        XDSServerService,
        ConfigService,
        SyncthingService,
        AlertService
    ],
    bootstrap: [AppComponent]
})
export class AppModule {
}