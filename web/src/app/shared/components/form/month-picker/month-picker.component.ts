import { CommonModule } from '@angular/common';
import { Component, EventEmitter, forwardRef, Output } from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroChevronLeft, heroChevronRight } from '@ng-icons/heroicons/outline';

export interface MonthYear {
  month: number;
  year: number;
}

const MONTH_NAMES = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
];

@Component({
  selector: 'app-month-picker',
  imports: [CommonModule, NgIcon],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => MonthPickerComponent),
      multi: true,
    },
    provideIcons({ heroChevronLeft, heroChevronRight }),
  ],
  template: `
    <div class="flex items-center gap-2">
      <button
        type="button"
        class="flex h-8 w-8 items-center justify-center rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-400 dark:hover:bg-gray-800"
        (click)="prevMonth()"
        aria-label="Previous month"
      >
        <ng-icon name="heroChevronLeft" size="16" />
      </button>
      <div class="flex items-center gap-1">
        <span class="text-sm font-semibold text-gray-800 dark:text-white/90">
          {{ monthName }}
        </span>
        <select
          class="rounded border border-gray-200 bg-white px-1 py-0.5 text-sm text-gray-800 focus:border-brand-300 focus:outline-none dark:border-gray-700 dark:bg-gray-900 dark:text-white/90"
          [value]="value.year"
          (change)="onYearChange($event)"
          aria-label="Select year"
        >
          @for (year of yearOptions; track year) {
            <option [value]="year">{{ year }}</option>
          }
        </select>
      </div>
      <button
        type="button"
        class="flex h-8 w-8 items-center justify-center rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-400 dark:hover:bg-gray-800"
        (click)="nextMonth()"
        aria-label="Next month"
      >
        <ng-icon name="heroChevronRight" size="16" />
      </button>
    </div>
  `,
})
export class MonthPickerComponent implements ControlValueAccessor {
  @Output() valueChange = new EventEmitter<MonthYear>();

  value: MonthYear = {
    month: new Date().getMonth(),
    year: new Date().getFullYear(),
  };

  private onChange: (value: MonthYear) => void = () => { /* Set via registerOnChange */ };
  private onTouched: () => void = () => { /* Set via registerOnTouched */ };

  get monthName(): string {
    return MONTH_NAMES[this.value.month];
  }

  get yearOptions(): number[] {
    const currentYear = new Date().getFullYear();
    const years: number[] = [];
    for (let y = currentYear - 5; y <= currentYear + 5; y++) {
      years.push(y);
    }
    return years;
  }

  prevMonth(): void {
    let { month, year } = this.value;
    month--;
    if (month < 0) {
      month = 11;
      year--;
    }
    this.setValue({ month, year });
  }

  nextMonth(): void {
    let { month, year } = this.value;
    month++;
    if (month > 11) {
      month = 0;
      year++;
    }
    this.setValue({ month, year });
  }

  onYearChange(event: Event): void {
    const select = event.target as HTMLSelectElement;
    const year = parseInt(select.value, 10);
    this.setValue({ ...this.value, year });
  }

  private setValue(val: MonthYear): void {
    this.value = val;
    this.onChange(val);
    this.onTouched();
    this.valueChange.emit(val);
  }

  writeValue(value: MonthYear | null): void {
    if (value) {
      this.value = value;
    }
  }

  registerOnChange(fn: (value: MonthYear) => void): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: () => void): void {
    this.onTouched = fn;
  }

  setDisabledState(): void {
    // Could disable buttons/select if needed
  }
}
