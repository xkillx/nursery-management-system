import { CommonModule } from '@angular/common';
import { Component, Input } from '@angular/core';

@Component({
  selector: 'app-form-field',
  imports: [CommonModule],
  template: `
    <div [class]="className">
      @if (label) {
        <label
          class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-400"
          [attr.for]="labelFor"
        >
          {{ label }}
          @if (required) {
            <span class="text-error-500" aria-hidden="true"> *</span>
          }
        </label>
      }
      <ng-content></ng-content>
      @if (hint && !error) {
        <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">{{ hint }}</p>
      }
      @if (error) {
        <p class="mt-1.5 text-xs text-error-500">{{ error }}</p>
      }
    </div>
  `,
})
export class FormFieldComponent {
  @Input() label = '';
  @Input() labelFor = '';
  @Input() required = false;
  @Input() hint = '';
  @Input() error = '';
  @Input() className = '';
}
