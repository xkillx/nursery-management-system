
import { Component, Input, Output, EventEmitter, ElementRef, ViewChild, forwardRef } from '@angular/core';
import { NG_VALUE_ACCESSOR, ControlValueAccessor } from '@angular/forms';
import flatpickr from 'flatpickr';
import { LabelComponent } from '../label/label.component';
import "flatpickr/dist/flatpickr.css";

@Component({
  selector: 'app-date-picker',
  imports: [LabelComponent],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => DatePickerComponent),
      multi: true,
    },
  ],
  templateUrl: './date-picker.component.html',
  styles: ``
})
export class DatePickerComponent implements ControlValueAccessor {

  @Input() id!: string;
  @Input() name: string = '';
  @Input() mode: 'single' | 'multiple' | 'range' | 'time' = 'single';
  @Input() defaultDate?: string | Date | string[] | Date[];
  @Input() label?: string;
  @Input() placeholder: string = 'Select date';
  @Input() error: boolean = false;
  @Input() className: string = '';
  @Input() min?: string;
  @Input() max?: string;
  @Input() disabled: boolean = false;
  @Input() describedBy?: string;
  @Output() dateChange = new EventEmitter<any>();
  @Output() blurred = new EventEmitter<void>();

  @ViewChild('dateInput', { static: false }) dateInput!: ElementRef<HTMLInputElement>;

  private flatpickrInstance: flatpickr.Instance | undefined;
  private onChange: (value: string) => void = () => {};
  private onTouched: () => void = () => {};
  private pendingValue: string | null = null;

  get inputClasses(): string {
    let classes = `h-11 w-full rounded-lg border appearance-none px-4 py-2.5 text-sm shadow-theme-xs placeholder:text-gray-400 focus:outline-hidden focus:ring-3 dark:bg-gray-900 dark:text-white/90 dark:placeholder:text-white/30 ${this.className}`;

    if (this.disabled) {
      classes += ` text-gray-500 border-gray-300 opacity-40 bg-gray-100 cursor-not-allowed dark:bg-gray-800 dark:text-gray-400 dark:border-gray-700`;
    } else if (this.error) {
      classes += ` bg-transparent text-gray-800 border-error-500 focus:border-error-300 focus:ring-error-500/20 dark:text-error-400 dark:border-error-500 dark:focus:border-error-800`;
    } else {
      classes += ` bg-transparent text-gray-800 border-gray-300 focus:border-brand-300 focus:ring-brand-500/20 dark:border-gray-700 dark:text-white/90 dark:focus:border-brand-800`;
    }
    return classes;
  }

  ngAfterViewInit() {
    this.flatpickrInstance = flatpickr(this.dateInput.nativeElement, {
      mode: this.mode,
      static: true,
      monthSelectorType: 'static',
      dateFormat: 'Y-m-d',
      defaultDate: this.defaultDate,
      minDate: this.min,
      maxDate: this.max,
      onChange: (selectedDates, dateStr, instance) => {
        this.dateChange.emit({ selectedDates, dateStr, instance });
        this.onChange(dateStr);
      },
      onClose: () => {
        this.onTouched();
        this.blurred.emit();
      }
    });

    if (this.pendingValue !== null) {
      this.flatpickrInstance.setDate(this.pendingValue, true);
      this.pendingValue = null;
    }
  }

  ngOnDestroy() {
    if (this.flatpickrInstance) {
      this.flatpickrInstance.destroy();
    }
  }

  writeValue(value: string | null): void {
    if (this.flatpickrInstance) {
      if (value) {
        this.flatpickrInstance.setDate(value, true);
      } else {
        this.flatpickrInstance.clear();
      }
    } else {
      this.pendingValue = value;
    }
  }

  registerOnChange(fn: (value: string) => void): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: () => void): void {
    this.onTouched = fn;
  }

  focus(): void {
    if (this.dateInput) {
      this.dateInput.nativeElement.focus();
    }
  }

  setDisabledState(isDisabled: boolean): void {
    this.disabled = isDisabled;
    if (this.flatpickrInstance) {
      this.flatpickrInstance.set('clickOpens', !isDisabled);
    }
  }
}
