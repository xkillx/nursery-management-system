import { CommonModule } from '@angular/common';
import { Component, Input } from '@angular/core';

@Component({
  selector: 'app-table-header',
  imports: [CommonModule],
  template: `
    <thead [ngClass]="'border-b border-gray-200 text-gray-500 dark:border-gray-800 dark:text-gray-400 ' + className"><ng-content></ng-content></thead>
  `,
  styles: ``
})
export class TableHeaderComponent {
  @Input() className = '';
}
