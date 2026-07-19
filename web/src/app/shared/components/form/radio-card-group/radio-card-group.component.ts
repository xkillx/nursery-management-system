import { CommonModule } from '@angular/common';
import { Component, forwardRef, input } from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';

export interface RadioCardOption {
  value: string;
  label: string;
  description?: string;
}

@Component({
  selector: 'app-radio-card-group',
  imports: [CommonModule],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => RadioCardGroupComponent),
      multi: true,
    },
  ],
  template: `
    <div class="grid gap-3" [ngClass]="gridClass()">
      @for (opt of options(); track opt.value) {
        <label
          class="relative flex cursor-pointer flex-col rounded-xl border p-4 transition-all duration-200"
          [ngClass]="value === opt.value
            ? 'border-brand-500 bg-brand-50 dark:bg-brand-500/10 ring-1 ring-brand-500'
            : 'border-gray-200 hover:border-gray-300 dark:border-gray-700 dark:hover:border-gray-600'"
        >
          <input
            type="radio"
            [name]="name()"
            [value]="opt.value"
            [checked]="value === opt.value"
            (change)="select(opt.value)"
            class="sr-only"
          />
          <span class="text-sm font-semibold text-gray-900 dark:text-white">{{ opt.label }}</span>
          @if (opt.description) {
            <span class="mt-1 text-xs text-gray-500 dark:text-gray-400 leading-normal">{{ opt.description }}</span>
          }
        </label>
      }
    </div>
  `,
})
export class RadioCardGroupComponent implements ControlValueAccessor {
  options = input.required<RadioCardOption[]>();
  name = input('radio-card');
  gridClass = input('sm:grid-cols-3');

  value = '';

  private propagateChange: (value: string) => void = () => { /* Set via registerOnChange */ };
  private propagateTouched: () => void = () => { /* Set via registerOnTouched */ };

  select(value: string): void {
    this.value = value;
    this.propagateChange(value);
    this.propagateTouched();
  }

  writeValue(value: string): void {
    this.value = value ?? '';
  }

  registerOnChange(fn: (value: string) => void): void {
    this.propagateChange = fn;
  }

  registerOnTouched(fn: () => void): void {
    this.propagateTouched = fn;
  }
}
