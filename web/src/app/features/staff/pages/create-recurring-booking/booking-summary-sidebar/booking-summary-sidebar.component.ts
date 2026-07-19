import { CommonModule } from '@angular/common';
import { Component, Input, Output, EventEmitter, OnChanges } from '@angular/core';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroCalendarDays, heroClock, heroCurrencyPound, heroAcademicCap, heroCheck } from '@ng-icons/heroicons/outline';

import { SessionEntry } from '../../../models/booking.models';
import { StaffSessionType } from '../../../data/session-types-api.service';

interface SessionDisplay {
  dayName: string;
  sessionName: string;
  durationHours: number;
}

const DAY_NAMES = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'];

@Component({
  selector: 'app-booking-summary-sidebar',
  imports: [CommonModule, NgIcon],
  host: { class: 'block' },
  providers: [
    provideIcons({ heroCalendarDays, heroClock, heroCurrencyPound, heroAcademicCap, heroCheck }),
  ],
  template: `
    <div class="rounded-xl border border-gray-100 bg-white shadow-theme-sm overflow-hidden dark:border-gray-800 dark:bg-gray-900/20">
        <!-- Header -->
        <div class="bg-brand-600 px-5 py-4.5 text-white dark:bg-brand-700">
          <h3 class="text-lg font-bold tracking-tight">Booking Summary</h3>
        </div>

        <div class="p-5 space-y-4">
          <!-- Selected Sessions List -->
          <div>
            <div class="text-[11px] font-bold text-gray-400 dark:text-gray-500 tracking-wider uppercase mb-3">
              Selected Sessions
            </div>
            
            @if (sessions.length > 0) {
              <div class="space-y-2.5">
                @for (s of sessions; track $index) {
                  <div class="flex items-center justify-between p-3.5 bg-brand-50/20 border border-brand-100/50 rounded-xl text-sm dark:bg-brand-500/5 dark:border-brand-500/10">
                    <div class="flex items-center gap-3">
                      <span class="flex h-5 w-5 shrink-0 items-center justify-center rounded-full border border-brand-500 text-brand-500 bg-white dark:bg-gray-800 dark:border-brand-400 dark:text-brand-400">
                        <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="4">
                          <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                        </svg>
                      </span>
                      <span class="text-gray-700 font-semibold dark:text-gray-300">{{ s.dayName }}: {{ s.sessionName }}</span>
                    </div>
                    <span class="text-sm font-semibold text-gray-500 dark:text-gray-400">{{ s.durationHours | number:'1.0-2' }} hrs</span>
                  </div>
                }
              </div>
            } @else {
              <div class="py-8 text-center text-sm text-gray-400 dark:text-gray-500">
                Select sessions on the grid to see a summary.
              </div>
            }
          </div>

          @if (sessions.length > 0) {
            <!-- Divider -->
            <div class="border-t border-gray-100 dark:border-gray-800 my-4"></div>

            <!-- Two Cards -->
            <div class="grid grid-cols-2 gap-4">
              <div class="rounded-xl bg-brand-50/20 p-4 text-center border border-brand-100/20 dark:bg-brand-500/5 dark:border-brand-500/10">
                <div class="text-xs font-bold text-gray-400 dark:text-gray-500 tracking-wide uppercase">Weekly Sessions</div>
                <div class="text-3xl font-black text-brand-600 dark:text-brand-400 mt-1.5">{{ sessions.length }}</div>
              </div>
              <div class="rounded-xl bg-brand-50/20 p-4 text-center border border-brand-100/20 dark:bg-brand-500/5 dark:border-brand-500/10">
                <div class="text-xs font-bold text-gray-400 dark:text-gray-500 tracking-wide uppercase">Weekly Hours</div>
                <div class="text-3xl font-black text-brand-600 dark:text-brand-400 mt-1.5">{{ totalWeeklyHours | number:'1.0-2' }}</div>
              </div>
            </div>

            <!-- Middle Stats -->
            <div class="space-y-3 pt-2">
              @if (fundingType && fundingHours && fundingType !== 'none') {
                <div class="flex items-center justify-between text-sm">
                  <span class="text-gray-600 dark:text-gray-400 font-semibold">Funded Hours{{ getFundingTypeLabel() }}</span>
                  <span class="text-emerald-600 dark:text-emerald-400 font-bold">- {{ fundingHours | number:'1.2-2' }}</span>
                </div>
              }

              <div class="flex items-center justify-between text-sm">
                <span class="text-gray-600 dark:text-gray-400 font-semibold">Chargeable Hours</span>
                <span class="font-bold text-gray-900 dark:text-white">{{ chargeableHours | number:'1.2-2' }}</span>
              </div>
            </div>

            <!-- Dashed Divider -->
            <div class="border-t border-dashed border-gray-200 dark:border-gray-800 my-4"></div>

            <!-- Weekly Total -->
            <div class="flex items-center justify-between my-4">
              <span class="text-base font-bold text-gray-900 dark:text-white">Weekly Total</span>
              <span class="text-2xl font-black text-brand-600 dark:text-brand-400">
                {{ (hourlyRateMinor !== null && hourlyRateMinor > 0) ? formatGbp(weeklyCostMinor) : '£0.00' }}
              </span>
            </div>
            @if (hourlyRateMinor === null || hourlyRateMinor <= 0) {
              <div class="rounded-lg bg-amber-50 p-3 text-xs text-amber-700 dark:bg-amber-500/10 dark:text-amber-400">
                Set up billing rate to see cost estimate.
              </div>
            }

            <!-- Solid Divider -->
            <div class="border-t border-gray-100 dark:border-gray-800 my-4"></div>

            <!-- Action Buttons -->
            <div class="space-y-3 pt-1">
              <button
                type="button"
                class="flex w-full min-h-12 items-center justify-center gap-2 rounded-xl bg-brand-600 text-white font-bold text-sm shadow-xs transition-all hover:bg-brand-700 focus:outline-hidden focus:ring-2 focus:ring-brand-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed dark:bg-brand-500 dark:hover:bg-brand-600"
                [disabled]="!canSubmit || isSaving"
                (click)="onSave()"
              >
                @if (isSaving) {
                  <svg class="size-4 animate-spin" viewBox="0 0 24 24" fill="none">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                  </svg>
                  Saving...
                } @else {
                  <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"></path>
                    <polyline points="17 21 17 13 7 13 7 21"></polyline>
                    <polyline points="7 3 7 8 15 8"></polyline>
                  </svg>
                  Save Booking
                }
              </button>

              <button
                type="button"
                class="flex w-full min-h-12 items-center justify-center rounded-xl border border-gray-200 bg-white text-gray-700 font-bold text-sm transition-all hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-white/[0.05]"
                (click)="onCancel()"
              >
                Cancel
              </button>
            </div>
          }
        </div>
      </div>
  `,
})
export class BookingSummarySidebarComponent implements OnChanges {
  @Input() childName: string | null = null;
  @Input() sessionEntries: SessionEntry[] = [];
  @Input() sessionTypes: StaffSessionType[] = [];
  @Input() fundingType: string | null = null;
  @Input() fundingHours: number | null = null;
  @Input() hourlyRateMinor: number | null = null;
  @Input() canSubmit = false;
  @Input() isSaving = false;

  @Output() save = new EventEmitter<void>();
  @Output() cancel = new EventEmitter<void>();

  sessions: SessionDisplay[] = [];
  totalWeeklyHours = 0;
  chargeableHours = 0;
  weeklyCostMinor = 0;

  ngOnChanges(): void {
    this.recompute();
  }

  getFundingTypeLabel(): string {
    if (this.fundingType === 'fifteen_hours') {
      return ' (Universal)';
    } else if (this.fundingType === 'thirty_hours') {
      return ' (Extended)';
    }
    return '';
  }

  onSave(): void {
    this.save.emit();
  }

  onCancel(): void {
    this.cancel.emit();
  }

  private recompute(): void {
    const typeMap = new Map<string, StaffSessionType>();
    for (const st of this.sessionTypes) {
      typeMap.set(st.id, st);
    }

    this.sessions = this.sessionEntries
      .map((e) => {
        const st = typeMap.get(e.session_type_id);
        if (!st) return null;
        const durationHours = this.computeDurationHours(st.startTime, st.endTime);
        return {
          dayName: DAY_NAMES[e.day_of_week] ?? `Day ${e.day_of_week}`,
          sessionName: st.name,
          durationHours,
        };
      })
      .filter((s): s is SessionDisplay => s !== null);

    this.totalWeeklyHours = this.sessions.reduce((sum, s) => sum + s.durationHours, 0);

    const fundedHours = this.fundingType && this.fundingHours ? this.fundingHours : 0;
    this.chargeableHours = Math.max(0, this.totalWeeklyHours - fundedHours);

    if (this.hourlyRateMinor !== null && this.hourlyRateMinor > 0) {
      this.weeklyCostMinor = Math.round(this.chargeableHours * this.hourlyRateMinor);
    } else {
      this.weeklyCostMinor = 0;
    }
  }

  private computeDurationHours(startTime: string, endTime: string): number {
    if (!startTime || !endTime) return 0;
    const [sh, sm] = startTime.split(':').map(Number);
    const [eh, em] = endTime.split(':').map(Number);
    return ((eh * 60 + em) - (sh * 60 + sm)) / 60;
  }

  formatGbp(minor: number): string {
    return `£${(minor / 100).toFixed(2)}`;
  }
}
