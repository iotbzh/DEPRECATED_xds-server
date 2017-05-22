import {Routes, RouterModule} from "@angular/router";
import {ModuleWithProviders} from "@angular/core";
import {ConfigComponent} from "./config/config.component";
import {HomeComponent} from "./home/home.component";
import {DevelComponent} from "./devel/devel.component";


const appRoutes: Routes = [
    {path: '', redirectTo: 'home', pathMatch: 'full'},

    {path: 'config', component: ConfigComponent, data: {title: 'Config'}},
    {path: 'home', component: HomeComponent, data: {title: 'Home'}},
    {path: 'devel', component: DevelComponent, data: {title: 'Build & Deploy'}}
];

export const AppRoutingProviders: any[] = [];
export const Routing: ModuleWithProviders = RouterModule.forRoot(appRoutes, {
    useHash: true
});
