import { CommonModule } from '@angular/common';
import { Component, EventEmitter, forwardRef, Input, Output } from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroPlus, heroCheck } from '@ng-icons/heroicons/outline';

import { StaffSessionType } from '../../../../features/staff/data/session-types-api.service';
import { SessionEntry } from '../../../../features/staff/models/booking.models';

const DEFAULT_DAYS = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'];
const WEEKDAY_INDICES = [0, 1, 2, 3, 4];

@Component({
  selector: 'app-session-grid',
  imports: [CommonModule, NgIcon],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => SessionGridComponent),
      multi: true,
    },
    provideIcons({ heroPlus, heroCheck }),
  ],
  template: `
    <div class="space-y-4">
      <div class="flex gap-2.5">
        <button
          type="button"
          class="rounded-xl border border-gray-200 bg-white px-3 py-1.5 text-xs font-semibold text-gray-600 hover:bg-gray-50 transition-all dark:border-gray-700 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-white/[0.05]"
          (click)="selectAllWeekdays()"
        >
          Select All Weekdays
        </button>
        <button
          type="button"
          class="rounded-xl border border-gray-200 bg-white px-3 py-1.5 text-xs font-semibold text-gray-600 hover:bg-gray-50 transition-all dark:border-gray-700 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-white/[0.05]"
          (click)="clearAll()"
        >
          Clear All
        </button>
      </div>

      <div class="overflow-x-auto rounded-xl border border-gray-100 dark:border-gray-800">
        <table class="w-full text-sm border-collapse">
          <thead>
            <tr class="border-b border-gray-100 bg-gray-50/55 dark:border-gray-800 dark:bg-gray-800/30">
              <th class="px-4 py-3 text-left text-xs font-bold text-gray-400 dark:text-gray-500 uppercase tracking-wider w-32">Day</th>
              @for (st of sessionTypes; track st.id) {
                <th class="px-3 py-3 text-center text-xs font-bold text-gray-400 dark:text-gray-500 uppercase tracking-wider">
                  <div class="text-gray-700 dark:text-gray-300 font-semibold">{{ st.name }}</div>
                  <div class="mt-0.5 font-normal text-gray-400 dark:text-gray-500 lowercase tracking-normal">({{ formatTime(st.startTime) }}–{{ formatTime(st.endTime) }})</div>
                </th>
              }
            </tr>
          </thead>
          <tbody>
            @for (day of days; track day.index; let i = $index) {
              <tr class="border-b border-gray-100 dark:border-gray-800/50 last:border-b-0 hover:bg-gray-50/40 dark:hover:bg-gray-800/10 transition-colors">
                <td class="px-4 py-3 text-sm font-bold text-gray-900 dark:text-white">{{ day.label }}</td>
                @for (st of sessionTypes; track st.id) {
                  <td class="px-3 py-2 text-center">
                    <button
                      type="button"
                      class="mx-auto flex h-12 w-full min-w-[72px] max-w-[120px] items-center justify-center rounded-xl border-2 transition-all duration-200 cursor-pointer"
                      [ngClass]="isSelected(day.index, st.id)
                        ? 'border-brand-500 bg-brand-500 text-white shadow-theme-xs'
                        : 'border-dashed border-gray-200 text-gray-300 hover:border-brand-500 hover:text-brand-500 dark:border-gray-800 dark:text-gray-700 dark:hover:border-brand-500 dark:hover:text-brand-500'"
                      (click)="toggleCell(day.index, st.id)"
                    >
                      @if (isSelected(day.index, st.id)) {
                        <ng-icon name="heroCheck" size="16" />
                      } @else {
                        <ng-icon name="heroPlus" size="16" />
                      }
                    </button>
                  </td>
                }
              </tr>
            }
          </tbody>
        </table>
      </div>
    </div>
  `,
})
export class SessionGridComponent implements ControlValueAccessor {
  @Input() sessionTypes: StaffSessionType[] = [];
  @Input() days: { label: string; index: number }[] = DEFAULT_DAYS.map((label, index) => ({ label, index }));
  @Output() valueChange = new EventEmitter<SessionEntry[]>();

  selected: Record<number, Set<string>> = {};

  private onChange: (value: SessionEntry[]) => void = () => { /* Set via registerOnChange */ };
  private onTouched: () => void = () => { /* Set via registerOnTouched */ };

  isSelected(dayIndex: number, sessionTypeId: string): boolean {
    return this.selected[dayIndex]?.has(sessionTypeId) ?? false;
  }

  toggleCell(dayIndex: number, sessionTypeId: string): void {
    if (!this.selected[dayIndex]) {
      this.selected[dayIndex] = new Set();
    }
    if (this.selected[dayIndex].has(sessionTypeId)) {
      this.selected[dayIndex].delete(sessionTypeId);
    } else {
      this.selected[dayIndex].add(sessionTypeId);
    }
    this.emit();
  }

  selectAllWeekdays(): void {
    for (const idx of WEEKDAY_INDICES) {
      if (!this.selected[idx]) {
        this.selected[idx] = new Set();
      }
      for (const st of this.sessionTypes) {
        this.selected[idx].add(st.id);
      }
    }
    this.emit();
  }

  clearAll(): void {
    this.selected = {};
    this.emit();
  }

  formatTime(time: string): string {
    if (!time) return '';
    const parts = time.split(':');
    return `${parts[0]}:${parts[1]}`;
  }

  writeValue(value: SessionEntry[] | null): void {
    this.selected = {};
    if (value) {
      for (const entry of value) {
        if (!this.selected[entry.day_of_week]) {
          this.selected[entry.day_of_week] = new Set();
        }
        this.selected[entry.day_of_week].add(entry.session_type_id);
      }
    }
  }

  registerOnChange(fn: (value: SessionEntry[]) => void): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: () => void): void {
    this.onTouched = fn;
  }

  setDisabledState(): void {
    // Not implemented for this component
  }

  private emit(): void {
    const entries: SessionEntry[] = [];
    for (const [dayStr, typeIds] of Object.entries(this.selected)) {
      for (const typeId of typeIds) {
        entries.push({ day_of_week: +dayStr, session_type_id: typeId });
      }
    }
    this.onChange(entries);
    this.onTouched();
    this.valueChange.emit(entries);
  }
}
