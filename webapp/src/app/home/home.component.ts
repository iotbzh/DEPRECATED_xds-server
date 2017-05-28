import { Component, OnInit } from '@angular/core';

export interface ISlide {
    img?: string;
    imgAlt?: string;
    hText?: string;
    hHtml?: string;
    text?: string;
    html?: string;
    btn?: string;
    btnHref?: string;
}

@Component({
    selector: 'home',
    moduleId: module.id,
    template: `
        <style>
            .wide img {
                width: 98%;
            }
            .carousel-item {
                max-height: 90%;
            }
            h1, h2, h3, h4, p {
                color: #330066;
            }
            .html-inner {
                color: #330066;
            }
            h1 {
                font-size: 4em;
            }
            p {
                font-size: 2.5em;
            }

        </style>

        <div class="wide">
            <carousel [interval]="carInterval" [(activeSlide)]="activeSlideIndex">
                <slide *ngFor="let sl of slides; let index=index">
                    <img [src]="sl.img" [alt]="sl.imgAlt">
                    <div class="carousel-caption">
                        <h1 *ngIf="sl.hText">{{ sl.hText }}</h1>
                        <h1 *ngIf="sl.hHtml" class="html-inner" [innerHtml]="sl.hHtml"></h1>
                        <p   *ngIf="sl.text">{{ sl.text }}</p>
                        <div *ngIf="sl.html" class="html-inner" [innerHtml]="sl.html"></div>
                    </div>
                </slide>
            </carousel>
        </div>
    `
})

export class HomeComponent {

    public carInterval: number = 4000;

    // FIXME SEB - Add more slides and info
    public slides: ISlide[] = [
        {
            img: 'assets/images/iot-graphx.jpg',
            imgAlt: "iot graphx image",
            hText: "Welcome to XDS Dashboard !",
            text: "X(cross) Development System allows developers to easily cross-compile applications.",
        },
        {
            img: 'assets/images/iot-graphx.jpg',
            imgAlt: "iot graphx image",
            hText: "Create, Build, Deploy, Enjoy !",
        },
        {
            img: 'assets/images/iot-graphx.jpg',
            imgAlt: "iot graphx image",
            hHtml: '<p>To Start: click on <i class="fa fa-cog" style="color:#9d9d9d;"></i> icon and add new folder</p>',
        }
    ];

    constructor() { }
}