import { Component, Input, ViewChild } from '@angular/core';
import { ModalDirective } from 'ngx-bootstrap/modal';

@Component({
    selector: 'sdk-add-modal',
    templateUrl: './app/sdks/sdkAddModal.component.html',
})
export class SdkAddModalComponent {
    @ViewChild('sdkChildModal') public sdkChildModal: ModalDirective;

    @Input() title?: string;

    // TODO
    constructor() {
    }

    show() {
        this.sdkChildModal.show();
    }

    hide() {
        this.sdkChildModal.hide();
    }
}
