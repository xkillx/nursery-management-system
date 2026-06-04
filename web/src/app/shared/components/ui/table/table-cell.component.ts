import { CommonModule } from '@angular/common';
import { Component, Input } from '@angular/core';

@Component({
  selector: 'app-table-cell',
  imports: [CommonModule],
  template: `
    @if (isHeader) {
      <th [ngClass]="'py-2 pr-3 ' + className"><ng-content></ng-content></th>
    } @else {
      <td [ngClass]="'py-3 pr-3 ' + className"><ng-content></ng-content></td>
    }
  `,
  styles: ``,
})
export class TableCellComponent {
  @Input() isHeader = false;
  @Input() className = '';
}
