import { Component, OnInit } from '@angular/core';

export interface ISlide {
    img?: string;
    imgAlt?: string;
    hText?: string;
    pText?: string;
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
            h1, h2, h3, h4, p {
                color: #330066;
            }

        </style>
        <div class="wide">
            <carousel [interval]="carInterval" [(activeSlide)]="activeSlideIndex">
                <slide *ngFor="let sl of slides; let index=index">
                    <img [src]="sl.img" [alt]="sl.imgAlt">
                    <div class="carousel-caption" *ngIf="sl.hText">
                        <h2>{{ sl.hText }}</h2>
                        <p>{{ sl.pText }}</p>
                    </div>
                </slide>
            </carousel>
        </div>
    `
})

export class HomeComponent {

    public carInterval: number = 2000;

    // FIXME SEB - Add more slides and info
    public slides: ISlide[] = [
        {
            img: 'assets/images/iot-graphx.jpg',
            imgAlt: "iot graphx image",
            hText: "Welcome to XDS Dashboard !",
            pText: "X(cross) Development System allows developers to easily cross-compile applications.",
        },
        {
            //img: 'assets/images/beige.jpg',
            //imgAlt: "beige image",
            img: 'assets/images/iot-graphx.jpg',
            imgAlt: "iot graphx image",
            hText: "Create, Build, Deploy, Enjoy !",
            pText: "TODO...",
        }
    ];

    constructor() { }
}