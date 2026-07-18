import { CommonModule } from '@angular/common';
import { Component, EventEmitter, forwardRef, Output } from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';
import { DropdownComponent } from '../dropdown.component';

interface AttendanceStatus {
  value: string;
  label: string;
  icon: string;
  colorClass: string;
}

const ATTENDANCE_STATUSES: AttendanceStatus[] = [
  { value: 'booked', label: 'Booked', icon: '📋', colorClass: 'text-gray-600 bg-gray-100 dark:text-gray-400 dark:bg-gray-800' },
  { value: 'present', label: 'Present', icon: '✓', colorClass: 'text-success-700 bg-success-50 dark:text-success-400 dark:bg-success-500/10' },
  { value: 'absent', label: 'Absent', icon: '✗', colorClass: 'text-error-700 bg-error-50 dark:text-error-400 dark:bg-error-500/10' },
  { value: 'late_arrival', label: 'Late Arrival', icon: '⏰', colorClass: 'text-warning-700 bg-warning-50 dark:text-warning-400 dark:bg-warning-500/10' },
  { value: 'early_pickup', label: 'Early Pickup', icon: '🏃', colorClass: 'text-blue-light-700 bg-blue-light-50 dark:text-blue-light-400 dark:bg-blue-light-500/10' },
  { value: 'no_show', label: 'No Show', icon: '⊘', colorClass: 'text-error-800 bg-error-100 dark:text-error-300 dark:bg-error-500/20' },
];

@Component({
  selector: 'app-status-dropdown',
  imports: [CommonModule, DropdownComponent],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => StatusDropdownComponent),
      multi: true,
    },
  ],
  template: `
    <div class="relative">
      <button
        type="button"
        class="dropdown-toggle flex items-center gap-2 rounded-lg border border-gray-200 px-3 py-2 text-sm font-medium transition-colors hover:bg-gray-50 dark:border-gray-700 dark:hover:bg-gray-800"
        [ngClass]="selectedStatus?.colorClass ?? ''"
        (click)="toggle()"
        (keydown.enter)="toggle()"
        (keydown.space)="toggle(); $event.preventDefault()"
        [attr.aria-expanded]="isOpen"
        aria-haspopup="listbox"
        aria-label="Attendance status"
      >
        <span>{{ selectedStatus?.icon ?? '—' }}</span>
        <span>{{ selectedStatus?.label ?? 'Select status' }}</span>
      </button>
      <app-dropdown [isOpen]="isOpen" (closed)="close()">
        <div
          role="listbox"
          aria-label="Attendance status"
          class="py-1"
        >
          @for (status of statuses; track status.value) {
            <button
              type="button"
              role="option"
              class="flex w-full items-center gap-3 px-4 py-2 text-sm text-left transition-colors hover:bg-gray-50 dark:hover:bg-white/5"
              [ngClass]="status.colorClass"
              [attr.aria-selected]="status.value === selectedValue"
              (click)="select(status.value)"
              (keydown.arrowdown)="focusNext($event)"
              (keydown.arrowup)="focusPrev($event)"
              (keydown.enter)="select(status.value)"
              (keydown.escape)="close()"
            >
              <span class="w-5 text-center">{{ status.icon }}</span>
              <span>{{ status.label }}</span>
            </button>
          }
        </div>
      </app-dropdown>
    </div>
  `,
})
export class StatusDropdownComponent implements ControlValueAccessor {
  @Output() valueChange = new EventEmitter<string>();

  readonly statuses = ATTENDANCE_STATUSES;

  isOpen = false;
  selectedValue: string | null = null;

  private onChange: (value: string) => void = () => { /* Set via registerOnChange */ };
  private onTouched: () => void = () => { /* Set via registerOnTouched */ };

  get selectedStatus(): AttendanceStatus | undefined {
    return this.statuses.find((s) => s.value === this.selectedValue);
  }

  toggle(): void {
    this.isOpen = !this.isOpen;
  }

  close(): void {
    this.isOpen = false;
    this.onTouched();
  }

  select(value: string): void {
    this.selectedValue = value;
    this.isOpen = false;
    this.onChange(value);
    this.onTouched();
    this.valueChange.emit(value);
  }

  focusNext(event: Event): void {
    event.preventDefault();
    const buttons = (event.target as HTMLElement)
      .closest('[role="listbox"]')
      ?.querySelectorAll('button[role="option"]') as NodeListOf<HTMLButtonElement> | undefined;
    if (!buttons) return;
    const currentIndex = Array.from(buttons).indexOf(event.target as HTMLButtonElement);
    const nextIndex = Math.min(currentIndex + 1, buttons.length - 1);
    buttons[nextIndex]?.focus();
  }

  focusPrev(event: Event): void {
    event.preventDefault();
    const buttons = (event.target as HTMLElement)
      .closest('[role="listbox"]')
      ?.querySelectorAll('button[role="option"]') as NodeListOf<HTMLButtonElement> | undefined;
    if (!buttons) return;
    const currentIndex = Array.from(buttons).indexOf(event.target as HTMLButtonElement);
    const prevIndex = Math.max(currentIndex - 1, 0);
    buttons[prevIndex]?.focus();
  }

  writeValue(value: string | null): void {
    this.selectedValue = value;
  }

  registerOnChange(fn: (value: string) => void): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: () => void): void {
    this.onTouched = fn;
  }

  setDisabledState(): void {
    // Could disable the button if needed
  }
}
