import { CommonModule } from '@angular/common';
import { Component, EventEmitter, forwardRef, Input, OnInit, Output } from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';

@Component({
  selector: 'app-switch',
  imports: [CommonModule],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => SwitchComponent),
      multi: true,
    },
  ],
  template: `
    <label
      class="inline-flex cursor-pointer select-none items-center gap-3 text-sm font-medium"
      [ngClass]="disabled ? 'cursor-not-allowed opacity-60' : ''"
      [attr.aria-label]="ariaLabel || null"
      [attr.aria-describedby]="describedBy || null"
      (click)="handleClick($event)"
    >
      <span class="relative inline-block">
        <span
          class="block transition duration-150 ease-linear h-6 w-11 rounded-full"
          [ngClass]="disabled ? 'bg-gray-100 dark:bg-gray-800' : switchColors.background"
        ></span>
        <span
          class="absolute left-0.5 top-0.5 h-5 w-5 rounded-full shadow-theme-sm duration-150 ease-linear transform"
          [ngClass]="switchColors.knob"
        ></span>
      </span>
      @if (label) {
        <span [ngClass]="disabled ? 'text-gray-400' : 'text-gray-700 dark:text-gray-400'">{{ label }}</span>
      }
    </label>
  `,
})
export class SwitchComponent implements OnInit, ControlValueAccessor {
  @Input() label?: string;
  @Input() checked: boolean = false;
  @Input() defaultChecked: boolean = false;
  @Input() disabled: boolean = false;
  @Input() color: 'blue' | 'gray' = 'blue';
  @Input() name?: string;
  @Input() id?: string;
  @Input() ariaLabel?: string;
  @Input() describedBy?: string;
  @Input() className?: string;
  @Output() valueChange = new EventEmitter<boolean>();
  @Output() checkedChange = new EventEmitter<boolean>();

  isChecked: boolean = false;

  private propagateChange: (value: boolean) => void = () => {};
  private propagateTouched: () => void = () => {};

  ngOnInit(): void {
    this.isChecked = this.checked || this.defaultChecked;
  }

  handleClick(event: Event): void {
    if (this.disabled) return;
    event.preventDefault();
    this.isChecked = !this.isChecked;
    this.propagateChange(this.isChecked);
    this.propagateTouched();
    this.valueChange.emit(this.isChecked);
    this.checkedChange.emit(this.isChecked);
  }

  writeValue(value: boolean | null): void {
    this.isChecked = !!value;
  }

  registerOnChange(fn: (value: boolean) => void): void {
    this.propagateChange = fn;
  }

  registerOnTouched(fn: () => void): void {
    this.propagateTouched = fn;
  }

  setDisabledState(isDisabled: boolean): void {
    this.disabled = isDisabled;
  }

  get switchColors(): { background: string; knob: string } {
    if (this.color === 'gray') {
      return {
        background: this.isChecked
          ? 'bg-gray-800 dark:bg-white/10'
          : 'bg-gray-200 dark:bg-white/10',
        knob: this.isChecked
          ? 'translate-x-full bg-white'
          : 'translate-x-0 bg-white',
      };
    }
    return {
      background: this.isChecked
        ? 'bg-brand-500'
        : 'bg-gray-200 dark:bg-white/10',
      knob: this.isChecked
        ? 'translate-x-full bg-white'
        : 'translate-x-0 bg-white',
    };
  }
}
