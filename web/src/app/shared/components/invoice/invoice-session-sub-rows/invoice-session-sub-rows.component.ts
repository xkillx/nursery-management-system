import { ChangeDetectionStrategy, Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';

import { InvoiceSession } from '../../../models/invoice-session.models';
import {
  formatSessionDate,
  formatSessionBooking,
  formatSessionQuantity,
  formatUnitAmount,
} from '../invoice-session-rows/session-rows.formatters';

@Component({
  selector: 'app-invoice-session-sub-rows',
  imports: [CommonModule],
  changeDetection: ChangeDetectionStrategy.OnPush,
  templateUrl: './invoice-session-sub-rows.component.html',
})
export class InvoiceSessionSubRowsComponent {
  @Input() sessions: InvoiceSession[] = [];
  @Input() unitAmountMinor: number | null = null;
  @Input() leadingSpacers = 1;
  @Input() trailingSpacers = 1;

  readonly formatSessionDate = formatSessionDate;
  readonly formatSessionBooking = formatSessionBooking;
  readonly formatSessionQuantity = formatSessionQuantity;
  readonly formatUnitAmount = formatUnitAmount;

  spacerCells(count: number): number[] {
    return Array.from({ length: count }, (_, i) => i);
  }
}
