import { CommonModule } from '@angular/common';
import { Component, OnInit, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroArrowLeft,
  heroClock,
  heroInformationCircle,
  heroPencilSquare,
  heroPlus,
  heroShieldCheck,
} from '@ng-icons/heroicons/outline';

import { ROLES, ROLE_ROUTES } from '../../../../core/constants/roles';
import { AuthService } from '../../../../core/services/auth.service';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { TimePickerComponent } from '../../../../shared/components/form/time-picker/time-picker.component';
import {
  StaffSessionTypeInput,
  StaffSessionTypesApiService,
} from '../../../staff/data/session-types-api.service';

@Component({
  selector: 'app-owner-session-type-form',
  imports: [
    CommonModule,
    FormsModule,
    RouterModule,
    NgIcon,
    LoadingStateComponent,
    AlertComponent,
    TimePickerComponent,
  ],
  templateUrl: './owner-session-type-form.component.html',
  providers: [
    provideIcons({
      heroArrowLeft,
      heroClock,
      heroInformationCircle,
      heroPencilSquare,
      heroPlus,
      heroShieldCheck,
    }),
  ],
})
export class OwnerSessionTypeFormComponent implements OnInit {
  private readonly api = inject(StaffSessionTypesApiService);
  private readonly auth = inject(AuthService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);

  mode: 'create' | 'edit' = 'create';
  sessionTypeId: string | null = null;
  siteId: string | null = null;
  siteName = '';
  role: string | null = null;
  isOwner = false;

  get listRoute(): string {
    return this.isOwner ? ROLE_ROUTES.ownerSessionTypes : ROLE_ROUTES.managerSessionTypes;
  }

  loading = false;
  saving = false;
  pageError: string | null = null;
  fieldErrors: { name?: string; startTime?: string; endTime?: string } = {};

  form = {
    name: '',
    startTime: '08:00',
    endTime: '13:00',
  };

  ngOnInit(): void {
    const membership = this.auth.activeMembership();
    this.role = membership?.role ?? null;
    this.isOwner = this.role === ROLES.owner;
    if (membership?.branch_id) {
      this.siteId = membership.branch_id;
      this.siteName = membership.branch_name ?? 'Assigned site';
    }

    const id = this.route.snapshot.paramMap.get('sessionTypeId');
    if (id) {
      this.mode = 'edit';
      this.sessionTypeId = id;
      this.loadExisting(id);
    }
  }

  private loadExisting(id: string): void {
    if (!this.siteId) {
      this.pageError = 'No site available.';
      return;
    }
    this.loading = true;
    this.api.listSessionTypes(this.siteId, { includeArchived: true }).subscribe({
      next: (types) => {
        const t = types.find((x) => x.id === id);
        if (!t) {
          this.pageError = 'Session type not found.';
        } else {
          this.form.name = t.name;
          this.form.startTime = t.startTime;
          this.form.endTime = t.endTime;
        }
        this.loading = false;
      },
      error: (err) => {
        this.loading = false;
        this.pageError = err?.error?.message ?? 'Failed to load session type.';
      },
    });
  }

  onSubmit(): void {
    this.fieldErrors = {};
    if (!this.form.name.trim()) {
      this.fieldErrors.name = 'Name is required.';
    }
    if (!this.isValidTime(this.form.startTime)) {
      this.fieldErrors.startTime = 'Use HH:MM 24-hour format.';
    }
    if (!this.isValidTime(this.form.endTime)) {
      this.fieldErrors.endTime = 'Use HH:MM 24-hour format.';
    }
    if (
      this.fieldErrors.startTime === undefined &&
      this.fieldErrors.endTime === undefined &&
      this.toMinutes(this.form.startTime) >= this.toMinutes(this.form.endTime)
    ) {
      this.fieldErrors.endTime = 'End time must be after start time.';
    }
    if (Object.keys(this.fieldErrors).length > 0) {
      return;
    }
    if (!this.siteId) {
      this.pageError = 'No site available.';
      return;
    }
    this.saving = true;
    this.pageError = null;
    const payload: StaffSessionTypeInput = {
      name: this.form.name.trim(),
      start_time: this.form.startTime,
      end_time: this.form.endTime,
    };
    const op =
      this.mode === 'create'
        ? this.api.createSessionType(this.siteId, payload)
        : this.api.updateSessionType(this.siteId, this.sessionTypeId!, payload);
    op.subscribe({
      next: () => {
        this.saving = false;
        this.router.navigateByUrl(this.listRoute);
      },
      error: (err) => {
        this.saving = false;
        const code = err?.error?.code;
        if (code === 'session_type_name_duplicate') {
          this.fieldErrors.name = 'An active session type with this name already exists in this site.';
        } else if (code === 'session_type_invalid_time_order') {
          this.fieldErrors.endTime = 'End time must be after start time.';
        } else {
          this.pageError = err?.error?.message ?? 'Failed to save session type.';
        }
      },
    });
  }

  private isValidTime(s: string): boolean {
    if (!/^([01]?\d|2[0-3]):[0-5]\d$/.test(s.trim())) return false;
    return true;
  }

  private toMinutes(s: string): number {
    const [hh, mm] = s.split(':').map((p) => parseInt(p, 10));
    return hh * 60 + mm;
  }
}
