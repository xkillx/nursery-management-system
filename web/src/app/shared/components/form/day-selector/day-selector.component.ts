import { Component, EventEmitter, forwardRef, Input, Output } from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';
import { CheckboxComponent } from '../input/checkbox.component';

const DAY_LABELS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];

@Component({
  selector: 'app-day-selector',
  imports: [CheckboxComponent],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => DaySelectorComponent),
      multi: true,
    },
  ],
  template: `
    <div class="space-y-2">
      <div class="grid grid-cols-4 gap-2 sm:grid-cols-7">
        @for (day of days; track day.index) {
          <app-checkbox
            [id]="idPrefix + '-' + day.index"
            [name]="namePrefix + '-' + day.index"
            [label]="day.label"
            [checked]="isSelected(day.index)"
            (checkedChange)="toggleDay(day.index, $event)"
          />
        }
      </div>
      <app-checkbox
        [id]="idPrefix + '-select-all'"
        [name]="namePrefix + '-select-all'"
        label="Select All"
        [checked]="allSelected"
        (checkedChange)="toggleAll($event)"
      />
    </div>
  `,
})
export class DaySelectorComponent implements ControlValueAccessor {
  @Input() idPrefix = 'day';
  @Input() namePrefix = 'day';
  @Output() valueChange = new EventEmitter<number[]>();

  readonly days = DAY_LABELS.map((label, index) => ({ label, index }));

  selectedDays: number[] = [];

  private onChange: (value: number[]) => void = () => { /* Set via registerOnChange */ };
  private onTouched: () => void = () => { /* Set via registerOnTouched */ };

  get allSelected(): boolean {
    return this.selectedDays.length === 7;
  }

  isSelected(index: number): boolean {
    return this.selectedDays.includes(index);
  }

  toggleDay(index: number, checked: boolean): void {
    if (checked) {
      if (!this.selectedDays.includes(index)) {
        this.selectedDays = [...this.selectedDays, index].sort();
      }
    } else {
      this.selectedDays = this.selectedDays.filter((d) => d !== index);
    }
    this.emit();
  }

  toggleAll(checked: boolean): void {
    if (checked) {
      this.selectedDays = [0, 1, 2, 3, 4, 5, 6];
    } else {
      this.selectedDays = [];
    }
    this.emit();
  }

  private emit(): void {
    this.onChange(this.selectedDays);
    this.onTouched();
    this.valueChange.emit(this.selectedDays);
  }

  writeValue(value: number[] | null): void {
    this.selectedDays = value ?? [];
  }

  registerOnChange(fn: (value: number[]) => void): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: () => void): void {
    this.onTouched = fn;
  }

  setDisabledState(): void {
    // Handled by individual checkboxes
  }
}
