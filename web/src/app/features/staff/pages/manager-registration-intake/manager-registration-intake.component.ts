import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { StaffApiService } from '../../data/staff-api.service';
import { ChildRecord, ChildWritePayload } from '../../models/children.models';
import { ConsentWritePayload, RegistrationWorkflowStatus } from '../../models/registration-profile.models';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { ButtonComponent } from '../../../../shared/components/ui/button/button.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { FormFieldComponent } from '../../../../shared/components/form/form-field/form-field.component';

type StepperStep = 'child-basics' | 'medical-health' | 'contacts-collection' | 'consents-evidence' | 'review-complete';

@Component({
  selector: 'app-manager-registration-intake',
  imports: [
    CommonModule,
    FormsModule,

    PageHeaderComponent,
    ButtonComponent,
    AlertComponent,
    FormFieldComponent,
  ],
  templateUrl: './manager-registration-intake.component.html',
})
export class ManagerRegistrationIntakeComponent implements OnInit {
  private readonly staffApi = inject(StaffApiService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);

  readonly steps: { key: StepperStep; label: string }[] = [
    { key: 'child-basics', label: 'Child Basics' },
    { key: 'medical-health', label: 'Medical & Health' },
    { key: 'contacts-collection', label: 'Contacts & Collection' },
    { key: 'consents-evidence', label: 'Consents & Evidence' },
    { key: 'review-complete', label: 'Review & Complete' },
  ];

  currentStep: StepperStep = 'child-basics';
  childId: string | null = null;
  child: ChildRecord | null = null;
  workflowStatus: RegistrationWorkflowStatus | null = null;
  isNewRegistration = true;

  isLoading = false;
  isSaving = false;
  errorMessage: string | null = null;
  fieldErrors: Record<string, string> = {};
  successMessage: string | null = null;

  step1 = {
    full_name: '',
    date_of_birth: '',
    start_date: '',
    notes: '',
  };

  step3 = {
    collection_password: '',
  };

  step4 = {
    signer_name: '',
    signed_date: '',
    paper_form_on_file: false,
    urgent_medical_treatment: true,
    plasters: true,
    safeguarding_reporting_acknowledgement: true,
    area_senco_liaison: true,
    health_visitor_liaison: true,
    transition_documents: true,
    local_outings: true,
    face_painting: true,
    parent_supplied_sun_cream: true,
    parent_supplied_nappy_cream: true,
    development_profile_photos: true,
    nursery_display_boards: true,
    promotional_literature: true,
    nursery_website: true,
    staff_student_coursework: true,
    social_media: true,
  };

  ngOnInit(): void {
    const childIdParam = this.route.snapshot.paramMap.get('childId');
    if (childIdParam) {
      this.isNewRegistration = false;
      this.childId = childIdParam;
      this.loadChildAndStatus();
    }
  }

  get stepIndex(): number {
    return this.steps.findIndex(s => s.key === this.currentStep);
  }

  get isFirstStep(): boolean {
    return this.currentStep === 'child-basics';
  }

  get isLastStep(): boolean {
    return this.currentStep === 'review-complete';
  }

  get canMarkComplete(): boolean {
    return this.workflowStatus?.can_mark_complete ?? false;
  }

  get isReviewedComplete(): boolean {
    return this.workflowStatus?.is_reviewed_complete ?? false;
  }

  get needsReview(): boolean {
    return this.workflowStatus?.needs_review ?? false;
  }

  nextStep(): void {
    const idx = this.stepIndex;
    if (idx < this.steps.length - 1) {
      this.currentStep = this.steps[idx + 1].key;
      this.errorMessage = null;
    }
  }

  prevStep(): void {
    const idx = this.stepIndex;
    if (idx > 0) {
      this.currentStep = this.steps[idx - 1].key;
      this.errorMessage = null;
    }
  }

  goToStep(step: StepperStep): void {
    this.currentStep = step;
    this.errorMessage = null;
  }

  saveChildBasics(): void {
    if (!this.step1.full_name.trim() || !this.step1.date_of_birth || !this.step1.start_date) {
      this.errorMessage = 'Full name, date of birth, and start date are required.';
      return;
    }

    this.isSaving = true;
    this.errorMessage = null;
    this.fieldErrors = {};

    const payload: ChildWritePayload = {
      full_name: this.step1.full_name.trim(),
      date_of_birth: this.step1.date_of_birth,
      start_date: this.step1.start_date,
      notes: this.step1.notes.trim() || undefined,
    };

    this.staffApi.createChild(payload).subscribe({
      next: (child) => {
        this.child = child;
        this.childId = child.id;
        this.isSaving = false;
        this.nextStep();
        if (this.step1.notes.trim()) {
          this.staffApi.patchRegistrationProfile(child.id, {
            demographics_home: { child_notes: this.step1.notes.trim() },
          }).subscribe();
        }
      },
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.fieldErrors = mapped.fieldErrors;
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.child'));
      },
    });
  }

  saveMedicalHealth(): void {
    this.isSaving = true;
    this.errorMessage = null;

    this.staffApi.patchRegistrationProfile(this.childId!, {
      medical_dietary: {
        medical_conditions_status: 'unknown',
        prescribed_medication_status: 'unknown',
        dietary_requirements_status: 'unknown',
        immunisation_status: 'unknown',
      },
      demographics_home: { reviewed: true },
    } as any).subscribe({
      next: () => {
        this.isSaving = false;
        this.nextStep();
      },
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'registration.intake'));
      },
    });
  }

  saveContactsCollection(): void {
    this.isSaving = true;
    this.errorMessage = null;

    const patch: Record<string, unknown> = {
      emergency_collection: {
        over_18_collection_acknowledged: true,
        reviewed: true,
      },
    };

    this.staffApi.patchRegistrationProfile(this.childId!, patch).subscribe({
      next: () => {
        if (this.step3.collection_password) {
          this.staffApi.setRegistrationCollectionPassword(this.childId!, this.step3.collection_password).subscribe({
            next: () => {
              this.isSaving = false;
              this.nextStep();
            },
            error: (err) => {
              this.isSaving = false;
              const mapped = this.errorMapper.mapAndHandle(err);
              this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'registration.intake'));
            },
          });
        } else {
          this.isSaving = false;
          this.nextStep();
        }
      },
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'registration.intake'));
      },
    });
  }

  saveConsentsEvidence(): void {
    if (!this.step4.signer_name.trim() || !this.step4.signed_date) {
      this.errorMessage = 'Signer name and signed date are required.';
      return;
    }

    this.isSaving = true;
    this.errorMessage = null;

    const consentPayload: ConsentWritePayload = {
      signer_name: this.step4.signer_name.trim(),
      signed_date: this.step4.signed_date,
      paper_form_on_file: this.step4.paper_form_on_file,
      urgent_medical_treatment: this.step4.urgent_medical_treatment,
      plasters: this.step4.plasters,
      safeguarding_reporting_acknowledgement: this.step4.safeguarding_reporting_acknowledgement,
      area_senco_liaison: this.step4.area_senco_liaison,
      health_visitor_liaison: this.step4.health_visitor_liaison,
      transition_documents: this.step4.transition_documents,
      local_outings: this.step4.local_outings,
      face_painting: this.step4.face_painting,
      parent_supplied_sun_cream: this.step4.parent_supplied_sun_cream,
      parent_supplied_nappy_cream: this.step4.parent_supplied_nappy_cream,
      development_profile_photos: this.step4.development_profile_photos,
      nursery_display_boards: this.step4.nursery_display_boards,
      promotional_literature: this.step4.promotional_literature,
      nursery_website: this.step4.nursery_website,
      staff_student_coursework: this.step4.staff_student_coursework,
      social_media: this.step4.social_media,
    };

    this.staffApi.createRegistrationConsent(this.childId!, consentPayload).subscribe({
      next: () => {
        this.staffApi.patchRegistrationOfficeUseChecklist(this.childId!, {
          birth_certificate_passport_status: 'unknown',
          proof_of_address_status: 'unknown',
        } as any).subscribe({
          next: () => {
            this.isSaving = false;
            this.loadStatus();
            this.nextStep();
          },
          error: () => {
            this.isSaving = false;
            this.loadStatus();
            this.nextStep();
          },
        });
      },
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.fieldErrors = mapped.fieldErrors;
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'registration.intake'));
      },
    });
  }

  markComplete(): void {
    this.isSaving = true;
    this.errorMessage = null;

    this.staffApi.createRegistrationCompletionAttestation(this.childId!).subscribe({
      next: () => {
        this.isSaving = false;
        this.successMessage = 'Registration marked as reviewed and complete.';
        this.loadStatus();
      },
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'registration.intake'));
      },
    });
  }

  navigateToChildDetail(): void {
    if (this.childId) {
      this.router.navigate(['/staff/manager/children', this.childId]);
    }
  }

  private loadChildAndStatus(): void {
    if (!this.childId) return;
    this.isLoading = true;

    this.staffApi.getChild(this.childId).subscribe({
      next: (child) => {
        this.child = child;
        this.loadStatus();
      },
      error: (error) => {
        this.isLoading = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.child'));
      },
    });
  }

  private loadStatus(): void {
    if (!this.childId) return;
    this.staffApi.getRegistrationWorkflowStatus(this.childId).subscribe({
      next: (status) => {
        this.workflowStatus = status;
        this.isLoading = false;
      },
      error: (err) => {
        this.isLoading = false;
        this.workflowStatus = null;
      },
    });
  }
}
