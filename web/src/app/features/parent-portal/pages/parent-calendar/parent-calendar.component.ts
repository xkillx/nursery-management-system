import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';

import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { ParentCalendarApiService } from '../../data/parent-calendar-api.service';
import { CalendarDayItem } from '../../models/parent-calendar.models';

@Component({
  selector: 'app-parent-calendar',
  imports: [
    CommonModule,
    PageHeaderComponent,
    EmptyStateComponent,
    LoadingStateComponent,
    AlertComponent,
  ],
  templateUrl: './parent-calendar.component.html',
})
export class ParentCalendarComponent implements OnInit {
  private readonly calendarApi = inject(ParentCalendarApiService);

  closureDays: CalendarDayItem[] = [];
  holidayPeriods: CalendarDayItem[] = [];
  isLoading = false;
  errorMessage: string | null = null;

  ngOnInit(): void {
    this.loadCalendar();
  }

  get nextClosure(): CalendarDayItem | null {
    const today = new Date().toISOString().split('T')[0];
    return this.closureDays.find((d) => d.date >= today) ?? null;
  }

  get nextHoliday(): CalendarDayItem | null {
    const today = new Date().toISOString().split('T')[0];
    return this.holidayPeriods.find((d) => d.date >= today) ?? null;
  }

  formatDate(dateStr: string): string {
    const date = new Date(dateStr + 'T00:00:00');
    return date.toLocaleDateString('en-GB', {
      weekday: 'short',
      day: 'numeric',
      month: 'short',
      year: 'numeric',
    });
  }

  private loadCalendar(): void {
    this.isLoading = true;
    this.errorMessage = null;

    const today = new Date();
    const from = new Date(today.getFullYear(), today.getMonth(), 1).toISOString().split('T')[0];
    const to = new Date(today.getFullYear(), today.getMonth() + 3, 0).toISOString().split('T')[0];

    this.calendarApi.listClosureDays(from, to).subscribe({
      next: (items) => {
        this.closureDays = items;
        this.checkLoaded();
      },
      error: (err) => {
        this.isLoading = false;
        this.errorMessage = err?.message ?? 'Failed to load closure days.';
      },
    });

    this.calendarApi.listHolidayPeriods(from, to).subscribe({
      next: (items) => {
        this.holidayPeriods = items;
        this.checkLoaded();
      },
      error: (err) => {
        this.isLoading = false;
        this.errorMessage = err?.message ?? 'Failed to load holiday periods.';
      },
    });
  }

  private checkLoaded(): void {
    if (this.closureDays.length >= 0 && this.holidayPeriods.length >= 0) {
      this.isLoading = false;
    }
  }
}
