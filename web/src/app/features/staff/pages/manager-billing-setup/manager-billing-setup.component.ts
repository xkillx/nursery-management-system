import { CommonModule } from '@angular/common';
import { Component, OnInit, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { StaffApiService } from '../../data/staff-api.service';

@Component({
  selector: 'app-manager-billing-setup',
  standalone: true,
  imports: [CommonModule, FormsModule, AlertComponent, LoadingStateComponent],
  templateUrl: './manager-billing-setup.component.html',
})
export class ManagerBillingSetupComponent implements OnInit {
  private readonly api = inject(StaffApiService);

  loading = true;
  error: string | null = null;
  rateMinor: number | null = null;
  hasRate = false;

  editing = false;
  editValue = '';
  editError: string | null = null;
  saving = false;

  ngOnInit(): void {
    this.loadRate();
  }

  get displayRate(): string {
    if (!this.hasRate || this.rateMinor === null) {
      return 'Not set';
    }
    return '£' + (this.rateMinor / 100).toFixed(2) + '/hr';
  }

  startEdit(): void {
    this.editing = true;
    this.editValue = this.rateMinor !== null ? (this.rateMinor / 100).toFixed(2) : '';
    this.editError = null;
  }

  cancelEdit(): void {
    this.editing = false;
    this.editValue = '';
    this.editError = null;
  }

  save(): void {
    const pounds = parseFloat(this.editValue);
    if (isNaN(pounds) || pounds <= 0) {
      this.editError = 'Enter a positive rate.';
      return;
    }
    const minor = Math.round(pounds * 100);
    this.editError = null;
    this.saving = true;

    this.api.updateSiteRate(minor).subscribe({
      next: () => {
        this.saving = false;
        this.editing = false;
        this.editValue = '';
        this.loadRate();
      },
      error: () => {
        this.saving = false;
        this.editError = 'Failed to save rate. Please try again.';
      },
    });
  }

  private loadRate(): void {
    this.loading = true;
    this.error = null;

    this.api.getSiteRate().subscribe({
      next: (res) => {
        this.rateMinor = res.core_hourly_rate_minor;
        this.hasRate = res.has_rate;
        this.loading = false;
      },
      error: () => {
        this.error = 'Failed to load billing setup. Please try again.';
        this.loading = false;
      },
    });
  }
}
