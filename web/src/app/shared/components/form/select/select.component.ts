import { Component, ElementRef, EventEmitter, forwardRef, Input, OnInit, Output, ViewChild } from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';

export interface Option {
  value: string;
  label: string;
}

@Component({
  selector: 'app-select',
  imports: [],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => SelectComponent),
      multi: true,
    },
  ],
  templateUrl: './select.component.html',
})
export class SelectComponent implements ControlValueAccessor, OnInit {
  @ViewChild('selectRef', { static: false }) selectRef!: ElementRef<HTMLSelectElement>;

  @Input() options: Option[] = [];
  @Input() placeholder = 'Select an option';
  @Input() className = '';
  @Input() defaultValue = '';
  @Input() value = '';
  @Input() id?: string = '';
  @Input() name?: string = '';
  @Input() disabled = false;
  @Input() error = false;
  @Input() describedBy?: string;
  @Input() placeholderDisabled = true;

  @Output() valueChange = new EventEmitter<string>();
  @Output() blurred = new EventEmitter<void>();

  focus(): void {
    if (this.selectRef) {
      this.selectRef.nativeElement.focus();
    }
  }

  private propagateChange: (value: string) => void = () => { /* Set via registerOnChange */ };
  private propagateTouched: () => void = () => { /* Set via registerOnTouched */ };

  ngOnInit() {
    if (!this.value && this.defaultValue) {
      this.value = this.defaultValue;
    }
  }

  onChange(event: Event) {
    const value = (event.target as HTMLSelectElement).value;
    this.value = value;
    this.propagateChange(value);
    this.valueChange.emit(value);
  }

  markTouched() {
    this.propagateTouched();
    this.blurred.emit();
  }

  writeValue(value: string | null): void {
    this.value = value ?? '';
  }

  registerOnChange(fn: (value: string) => void): void {
    this.propagateChange = fn;
  }

  registerOnTouched(fn: () => void): void {
    this.propagateTouched = fn;
  }

  setDisabledState(isDisabled: boolean): void {
    this.disabled = isDisabled;
  }
}
