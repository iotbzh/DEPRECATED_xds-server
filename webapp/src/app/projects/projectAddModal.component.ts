import { Component, Input, ViewChild, OnInit } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import { ModalDirective } from 'ngx-bootstrap/modal';
import { FormControl, FormGroup, Validators, FormBuilder, ValidatorFn, AbstractControl } from '@angular/forms';

// Import RxJs required methods
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/filter';
import 'rxjs/add/operator/debounceTime';

import { AlertService, IAlert } from "../services/alert.service";
import {
    ConfigService, IConfig, IProject, ProjectType, ProjectTypes,
    IxdsAgentPackage
} from "../services/config.service";


@Component({
    selector: 'project-add-modal',
    templateUrl: './app/projects/projectAddModal.component.html',
    styleUrls: ['./app/projects/projectAddModal.component.css']
})
export class ProjectAddModalComponent {
    @ViewChild('childProjectModal') public childProjectModal: ModalDirective;
    @Input() title?: string;

    config$: Observable<IConfig>;

    cancelAction: boolean = false;
    userEditedLabel: boolean = false;
    projectTypes = ProjectTypes;

    addProjectForm: FormGroup;
    typeCtrl: FormControl;
    pathCliCtrl: FormControl;
    pathSvrCtrl: FormControl;

    constructor(
        private alert: AlertService,
        private configSvr: ConfigService,
        private fb: FormBuilder
    ) {
        // Define types (first one is special/placeholder)
        this.projectTypes.unshift({ value: -1, display: "--Select a type--" });

        this.typeCtrl = new FormControl(this.projectTypes[0].value, Validators.pattern("[0-9]+"));
        this.pathCliCtrl = new FormControl("", Validators.required);
        this.pathSvrCtrl = new FormControl({ value: "", disabled: true }, [Validators.required, Validators.minLength(1)]);

        this.addProjectForm = fb.group({
            type: this.typeCtrl,
            pathCli: this.pathCliCtrl,
            pathSvr: this.pathSvrCtrl,
            label: ["", Validators.nullValidator],
        });
    }

    ngOnInit() {
        this.config$ = this.configSvr.conf;

        // Auto create label name
        this.pathCliCtrl.valueChanges
            .debounceTime(100)
            .filter(n => n)
            .map(n => "Project_" + n.split('/')[0])
            .subscribe(value => {
                if (value && !this.userEditedLabel) {
                    this.addProjectForm.patchValue({ label: value });
                }
            });

        // Handle disabling of Server path
        this.typeCtrl.valueChanges
            .debounceTime(500)
            .subscribe(valType => {
                let dis = (valType === String(ProjectType.SYNCTHING));
                this.pathSvrCtrl.reset({ value: "", disabled: dis });
            });
    }

    show() {
        this.cancelAction = false;
        this.childProjectModal.show();
    }

    hide() {
        this.childProjectModal.hide();
    }

    onKeyLabel(event: any) {
        this.userEditedLabel = (this.addProjectForm.value.label !== "");
    }

    /* FIXME: change input to file type
     <td><input type="file" id="select-local-path" webkitdirectory
     formControlName="pathCli" placeholder="myProject" (change)="onChangeLocalProject($event)"></td>

    onChangeLocalProject(e) {
        if e.target.files.length < 1 {
            console.log('SEB NO files');
        }
        let dir = e.target.files[0].webkitRelativePath;
        console.log("SEB files: " + dir);
        let u = URL.createObjectURL(e.target.files[0]);
    }
    */
    onChangeLocalProject(e) {
    }

    onSubmit() {
        if (this.cancelAction) {
            return;
        }

        let formVal = this.addProjectForm.value;

        let type = formVal['type'].value;
        let numType = Number(formVal['type']);
        this.configSvr.addProject({
            label: formVal['label'],
            pathClient: formVal['pathCli'],
            pathServer: formVal['pathSvr'],
            type: numType,
            // FIXME: allow to set defaultSdkID from New Project config panel
        })
            .subscribe(prj => {
                this.alert.info("Project " + prj.label + " successfully created.");
                this.hide();

                // Reset Value for the next creation
                this.addProjectForm.reset();
                let selectedType = this.projectTypes[0].value;
                this.addProjectForm.patchValue({ type: selectedType });

            },
            err => {
                this.alert.error("Configuration ERROR: " + err, 60);
                this.hide();
            });
    }

}
