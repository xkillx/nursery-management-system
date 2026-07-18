import { CommonModule } from '@angular/common';
import { Component, Input, OnChanges } from '@angular/core';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroCalendarDays, heroClock, heroCurrencyPound, heroAcademicCap } from '@ng-icons/heroicons/outline';

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
  providers: [
    provideIcons({ heroCalendarDays, heroClock, heroCurrencyPound, heroAcademicCap }),
  ],
  template: `
    <div class="sticky top-24 space-y-4">
      <div class="rounded-xl border border-gray-100 bg-white shadow-theme-sm overflow-hidden dark:border-gray-800 dark:bg-gray-900/20">
        <div class="bg-brand-500 px-5 py-4 text-white">
          <div class="flex items-center gap-2">
            <ng-icon name="heroCalendarDays" size="18" aria-hidden="true" />
            <h3 class="text-sm font-bold tracking-wide">
              {{ childName || 'New Recurring Booking' }}
            </h3>
          </div>
          <p class="text-xs mt-1 opacity-80 font-medium">Booking Summary</p>
        </div>

        <div class="p-5 space-y-4">
          @if (sessions.length > 0) {
            <div class="space-y-2">
              @for (s of sessions; track $index) {
                <div class="flex items-center justify-between p-2.5 bg-gray-50 rounded-lg border border-gray-100 text-sm dark:bg-gray-800/30 dark:border-gray-800">
                  <div class="flex items-center gap-2">
                    <ng-icon name="heroCheck" size="14" class="text-brand-500 font-bold" />
                    <span class="text-gray-700 font-medium dark:text-gray-300">{{ s.dayName }}: {{ s.sessionName }}</span>
                  </div>
                  <span class="text-xs text-gray-400">{{ s.durationHours | number:'1.1-1' }}h</span>
                </div>
              }
            </div>

            <div class="border-t border-gray-100 pt-3 dark:border-gray-800">
              <div class="grid grid-cols-2 gap-3">
                <div class="rounded-lg bg-gray-50 p-3 dark:bg-gray-800/50">
                  <div class="text-xs text-gray-500 dark:text-gray-400">Sessions</div>
                  <div class="text-lg font-semibold text-gray-900 dark:text-white">{{ sessions.length }}</div>
                </div>
                <div class="rounded-lg bg-gray-50 p-3 dark:bg-gray-800/50">
                  <div class="text-xs text-gray-500 dark:text-gray-400">Weekly Hours</div>
                  <div class="text-lg font-semibold text-gray-900 dark:text-white">{{ totalWeeklyHours | number:'1.1-1' }}h</div>
                </div>
              </div>
            </div>

            @if (fundingType && fundingHours) {
              <div class="border-t border-gray-100 pt-3 dark:border-gray-800">
                <div class="flex items-center justify-between text-sm">
                  <span class="flex items-center gap-1.5 text-gray-600 dark:text-gray-400">
                    <ng-icon name="heroAcademicCap" size="14" />
                    Funded Hours
                  </span>
                  <span class="text-green-600 dark:text-green-400 font-semibold">-{{ fundingHours | number:'1.1-1' }}h</span>
                </div>
              </div>
            }

            <div class="border-t border-gray-100 pt-3 dark:border-gray-800">
              <div class="flex items-center justify-between text-sm">
                <span class="text-gray-600 dark:text-gray-400">Chargeable Hours</span>
                <span class="font-semibold text-gray-950 dark:text-white">{{ chargeableHours | number:'1.1-1' }}h</span>
              </div>
            </div>

            <div class="border-t border-gray-100 pt-3 dark:border-gray-800">
              @if (hourlyRateMinor !== null && hourlyRateMinor > 0) {
                <div class="flex items-center justify-between">
                  <span class="flex items-center gap-1.5 text-sm font-bold text-gray-900 dark:text-white">
                    <ng-icon name="heroCurrencyPound" size="16" />
                    Estimated Weekly Total
                  </span>
                  <span class="text-lg font-black text-brand-600 dark:text-brand-400">{{ formatGbp(weeklyCostMinor) }}</span>
                </div>
              } @else {
                <div class="rounded-lg bg-amber-50 p-3 text-xs text-amber-700 dark:bg-amber-500/10 dark:text-amber-400">
                  Set up billing rate to see cost estimate.
                </div>
              }
            </div>
          } @else {
            <div class="py-8 text-center text-sm text-gray-400 dark:text-gray-500">
              Select sessions on the grid to see a summary.
            </div>
          }
        </div>
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

  sessions: SessionDisplay[] = [];
  totalWeeklyHours = 0;
  chargeableHours = 0;
  weeklyCostMinor = 0;

  ngOnChanges(): void {
    this.recompute();
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
