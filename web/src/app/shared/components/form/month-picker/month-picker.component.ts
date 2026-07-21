import { CommonModule } from '@angular/common';
import { Component, EventEmitter, Input, forwardRef, Output, ElementRef, HostListener, inject } from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroChevronLeft, heroChevronRight, heroCalendar } from '@ng-icons/heroicons/outline';

export interface MonthYear {
  month: number;
  year: number;
}

const MONTH_NAMES = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
];

const SHORT_MONTH_NAMES = [
  'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
  'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec',
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
    provideIcons({ heroChevronLeft, heroChevronRight, heroCalendar }),
  ],
  template: `
    <!-- Inline mode (original behavior) -->
    @if (mode === 'inline') {
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
            [value]="value?.year ?? dropdownYear"
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
    }

    <!-- Dropdown mode (datepicker-like) -->
    @if (mode === 'dropdown') {
      <div class="relative" #dropdownContainer>
        @if (label) {
          <label [for]="id" class="text-xs font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
            {{ label }}
          </label>
        }
        <button
          type="button"
          [id]="id"
          class="h-10 w-full rounded-lg border appearance-none px-3 text-sm text-left shadow-theme-xs transition focus:outline-hidden focus:ring-3 flex items-center gap-2"
          [class]="inputClasses"
          [disabled]="disabled"
          (click)="toggleDropdown()"
          [attr.aria-expanded]="isOpen"
          aria-haspopup="dialog"
        >
          <ng-icon name="heroCalendar" size="16" class="text-gray-400 shrink-0" />
          @if (value) {
            <span>{{ displayLabel }}</span>
          } @else {
            <span class="text-gray-400">{{ placeholder }}</span>
          }
        </button>

        @if (isOpen) {
          <div
            class="absolute z-50 mt-1 w-full min-w-[280px] rounded-xl border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-900"
            role="dialog"
            aria-label="Select month"
          >
            <!-- Year navigation -->
            <div class="flex items-center justify-between px-4 pt-3 pb-2">
              <button
                type="button"
                class="flex h-7 w-7 items-center justify-center rounded-lg text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800"
                (click)="prevDropdownYear()"
                aria-label="Previous year"
              >
                <ng-icon name="heroChevronLeft" size="16" />
              </button>
              <span class="text-sm font-bold text-gray-900 dark:text-white/90">{{ dropdownYear }}</span>
              <button
                type="button"
                class="flex h-7 w-7 items-center justify-center rounded-lg text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800"
                (click)="nextDropdownYear()"
                aria-label="Next year"
              >
                <ng-icon name="heroChevronRight" size="16" />
              </button>
            </div>

            <!-- Month grid -->
            <div class="grid grid-cols-4 gap-1 px-3 pb-3">
              @for (m of monthGrid; track m.index) {
                <button
                  type="button"
                  class="flex h-9 items-center justify-center rounded-lg text-xs font-semibold transition"
                  [class]="monthCellClasses(m.index)"
                  (click)="selectMonth(m.index)"
                >
                  {{ m.short }}
                </button>
              }
            </div>
          </div>
        }
      </div>
    }
  `,
})
export class MonthPickerComponent implements ControlValueAccessor {
  @Input() mode: 'inline' | 'dropdown' = 'inline';
  @Input() id = '';
  @Input() label = '';
  @Input() placeholder = 'Select month';
  @Input() disabled = false;
  @Input() error = false;
  @Input() className = '';

  @Output() valueChange = new EventEmitter<MonthYear>();

  value: MonthYear | null = null;

  isOpen = false;
  dropdownYear = new Date().getFullYear();

  private onChange: (value: MonthYear | string) => void = () => { /* Set via registerOnChange */ };
  private onTouched: () => void = () => { /* Set via registerOnTouched */ };

  private readonly elRef = inject(ElementRef);

  @HostListener('document:click', ['$event'])
  onDocumentClick(event: MouseEvent): void {
    if (!this.isOpen) return;
    const target = event.target as HTMLElement;
    if (!this.elRef.nativeElement.contains(target)) {
      this.isOpen = false;
    }
  }

  get monthName(): string {
    return this.value ? MONTH_NAMES[this.value.month] : '';
  }

  get displayLabel(): string {
    if (!this.value) return '';
    return `${MONTH_NAMES[this.value.month]} ${this.value.year}`;
  }

  get monthGrid(): { index: number; short: string }[] {
    return SHORT_MONTH_NAMES.map((short, index) => ({ index, short }));
  }

  get yearOptions(): number[] {
    const currentYear = new Date().getFullYear();
    const years: number[] = [];
    for (let y = currentYear - 5; y <= currentYear + 5; y++) {
      years.push(y);
    }
    return years;
  }

  get inputClasses(): string {
    let classes = `bg-transparent border-gray-300 focus:border-brand-300 focus:ring-brand-500/20 dark:border-gray-700 dark:bg-gray-900 dark:text-white/90 ${this.className}`;
    if (this.disabled) {
      classes = `text-gray-500 border-gray-300 opacity-40 bg-gray-100 cursor-not-allowed dark:bg-gray-800 dark:text-gray-400 dark:border-gray-700 ${this.className}`;
    } else if (this.error) {
      classes = `bg-transparent text-gray-800 border-error-500 focus:border-error-300 focus:ring-error-500/20 dark:text-error-400 dark:border-error-500 ${this.className}`;
    }
    return classes;
  }

  monthCellClasses(month: number): string {
    const isSelected = this.value && this.value.month === month && this.value.year === this.dropdownYear;
    const isCurrent = month === new Date().getMonth() && this.dropdownYear === new Date().getFullYear();

    if (isSelected) {
      return 'bg-brand-600 text-white hover:bg-brand-700';
    }
    if (isCurrent) {
      return 'bg-brand-50 text-brand-700 hover:bg-brand-100 dark:bg-brand-950/20 dark:text-brand-400';
    }
    return 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800';
  }

  toggleDropdown(): void {
    if (this.disabled) return;
    this.isOpen = !this.isOpen;
    if (this.isOpen && this.value) {
      this.dropdownYear = this.value.year;
    }
  }

  selectMonth(month: number): void {
    this.setValue({ month, year: this.dropdownYear });
    this.isOpen = false;
  }

  prevDropdownYear(): void {
    this.dropdownYear--;
  }

  nextDropdownYear(): void {
    this.dropdownYear++;
  }

  prevMonth(): void {
    let { month, year } = this.value || { month: new Date().getMonth(), year: new Date().getFullYear() };
    month--;
    if (month < 0) {
      month = 11;
      year--;
    }
    this.setValue({ month, year });
  }

  nextMonth(): void {
    let { month, year } = this.value || { month: new Date().getMonth(), year: new Date().getFullYear() };
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
    this.setValue({ ...(this.value || { month: 0 }), year });
  }

  private setValue(val: MonthYear): void {
    this.value = val;
    this.onChange(val);
    this.onTouched();
    this.valueChange.emit(val);
  }

  writeValue(value: MonthYear | string | null): void {
    if (!value) return;
    if (typeof value === 'string') {
      const [year, month] = value.split('-').map(Number);
      if (!isNaN(year) && !isNaN(month) && month >= 1 && month <= 12) {
        this.value = { month: month - 1, year };
      }
    } else {
      this.value = value;
    }
  }

  registerOnChange(fn: (value: MonthYear | string) => void): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: () => void): void {
    this.onTouched = fn;
  }

  setDisabledState(isDisabled: boolean): void {
    this.disabled = isDisabled;
  }
}
