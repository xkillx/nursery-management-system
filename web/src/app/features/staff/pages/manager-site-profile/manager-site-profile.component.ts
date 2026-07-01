import { CommonModule } from '@angular/common';
import { Component, OnInit, inject } from '@angular/core';
import { FormsModule, NgForm } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroArrowLeft,
  heroBuildingOffice2,
  heroEnvelope,
  heroGlobeAlt,
  heroMapPin,
  heroPhone,
  heroInformationCircle,
} from '@ng-icons/heroicons/outline';

import { ROLE_ROUTES } from '../../../../core/constants/roles';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { StaffSiteProfileApiService } from '../../data/staff-site-profile-api.service';
import { SiteProfile, SiteProfileInput } from '../../models/site-profile.models';

interface SiteProfileFormModel {
  nursery_name: string;
  description: string;
  phone: string;
  email: string;
  website: string;
  address_street: string;
  address_city: string;
  address_postcode: string;
}

@Component({
  selector: 'app-manager-site-profile',
  imports: [
    CommonModule,
    FormsModule,
    RouterModule,
    LoadingStateComponent,
    AlertComponent,
    NgIcon,
  ],
  templateUrl: './manager-site-profile.component.html',
  providers: [
    provideIcons({
      heroArrowLeft,
      heroBuildingOffice2,
      heroEnvelope,
      heroGlobeAlt,
      heroMapPin,
      heroPhone,
      heroInformationCircle,
    }),
  ],
})
export class ManagerSiteProfileComponent implements OnInit {
  private readonly api = inject(StaffSiteProfileApiService);
  private readonly router = inject(Router);

  loading = true;
  submitting = false;
  pageError: string | null = null;
  fieldErrors: Partial<Record<keyof SiteProfileFormModel, string>> = {};

  readonly siteSettingsRoute = ROLE_ROUTES.managerSiteSettings;

  model: SiteProfileFormModel = {
    nursery_name: '',
    description: '',
    phone: '',
    email: '',
    website: '',
    address_street: '',
    address_city: '',
    address_postcode: '',
  };

  ngOnInit(): void {
    this.loadProfile();
  }

  submit(form: NgForm): void {
    this.fieldErrors = {};
    this.pageError = null;

    if (!this.validate(form)) return;

    const input: SiteProfileInput = {
      nursery_name: this.model.nursery_name.trim(),
      description: this.model.description.trim(),
      phone: this.model.phone.trim(),
      email: this.model.email.trim(),
      website: this.model.website.trim(),
      address_street: this.model.address_street.trim(),
      address_city: this.model.address_city.trim(),
      address_postcode: this.model.address_postcode.trim(),
    };

    this.submitting = true;
    this.api.updateSiteProfile(input).subscribe({
      next: () => {
        this.submitting = false;
        this.router.navigate([this.siteSettingsRoute]);
      },
      error: (err) => {
        this.submitting = false;
        this.applyApiError(err);
      },
    });
  }

  onCancel(): void {
    this.router.navigate([this.siteSettingsRoute]);
  }

  private loadProfile(): void {
    this.loading = true;
    this.pageError = null;

    this.api.getSiteProfile().subscribe({
      next: (resp) => {
        this.loading = false;
        if (resp.site_profile) {
          const sp = resp.site_profile;
          this.model = {
            nursery_name: sp.nursery_name ?? '',
            description: sp.description ?? '',
            phone: sp.phone ?? '',
            email: sp.email ?? '',
            website: sp.website ?? '',
            address_street: sp.address_street ?? '',
            address_city: sp.address_city ?? '',
            address_postcode: sp.address_postcode ?? '',
          };
        }
      },
      error: () => {
        this.loading = false;
        this.pageError = 'Failed to load site profile. Please try again.';
      },
    });
  }

  private validate(form: NgForm): boolean {
    form.control.markAllAsTouched();

    const trimmed = {
      nursery_name: this.model.nursery_name.trim(),
      phone: this.model.phone.trim(),
      email: this.model.email.trim(),
      website: this.model.website.trim(),
      address_street: this.model.address_street.trim(),
      address_city: this.model.address_city.trim(),
      address_postcode: this.model.address_postcode.trim(),
    };

    if (!trimmed.nursery_name) {
      this.fieldErrors.nursery_name = 'is required';
    } else if (trimmed.nursery_name.length > 120) {
      this.fieldErrors.nursery_name = 'must be 120 characters or fewer';
    }

    if (!trimmed.phone) {
      this.fieldErrors.phone = 'is required';
    } else if (trimmed.phone.length > 32) {
      this.fieldErrors.phone = 'must be 32 characters or fewer';
    }

    if (!trimmed.email) {
      this.fieldErrors.email = 'is required';
    } else if (trimmed.email.length > 254) {
      this.fieldErrors.email = 'must be 254 characters or fewer';
    } else if (!this.isValidEmail(trimmed.email)) {
      this.fieldErrors.email = 'is not a valid email address';
    }

    if (!trimmed.website) {
      this.fieldErrors.website = 'is required';
    } else if (trimmed.website.length > 2048) {
      this.fieldErrors.website = 'must be 2048 characters or fewer';
    } else if (!this.isValidURL(trimmed.website)) {
      this.fieldErrors.website = 'is not a valid URL';
    }

    if (!trimmed.address_street) {
      this.fieldErrors.address_street = 'is required';
    } else if (trimmed.address_street.length > 200) {
      this.fieldErrors.address_street = 'must be 200 characters or fewer';
    }

    if (!trimmed.address_city) {
      this.fieldErrors.address_city = 'is required';
    } else if (trimmed.address_city.length > 100) {
      this.fieldErrors.address_city = 'must be 100 characters or fewer';
    }

    if (!trimmed.address_postcode) {
      this.fieldErrors.address_postcode = 'is required';
    } else if (trimmed.address_postcode.length > 16) {
      this.fieldErrors.address_postcode = 'must be 16 characters or fewer';
    }

    return Object.keys(this.fieldErrors).length === 0;
  }

  private applyApiError(err: any): void {
    const body = err?.error;
    if (body?.code === 'validation_error' && body?.details?.field_errors) {
      for (const fe of body.details.field_errors) {
        const key = fe.field as keyof SiteProfileFormModel;
        this.fieldErrors[key] = fe.message;
      }
      return;
    }
    this.pageError = 'Failed to save site profile. Please try again.';
  }

  private isValidEmail(email: string): boolean {
    const at = email.lastIndexOf('@');
    if (at <= 0 || at >= email.length - 1) return false;
    const domain = email.substring(at + 1);
    if (domain.length < 3) return false;
    return domain.includes('.');
  }

  private isValidURL(raw: string): boolean {
    try {
      const u = new URL(raw);
      return u.protocol !== '' && u.host !== '';
    } catch {
      return false;
    }
  }
}
