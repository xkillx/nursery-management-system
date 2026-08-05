import { ChangeDetectionStrategy, Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';

import { InvoiceSession, InvoiceLineAggregate } from '../../../models/invoice-session.models';
import {
  formatSessionDate,
  formatSessionBooking,
  formatSessionQuantity,
  formatUnitAmount,
} from './session-rows.formatters';

@Component({
  selector: 'app-invoice-session-rows',
  imports: [CommonModule],
  changeDetection: ChangeDetectionStrategy.OnPush,
  templateUrl: './invoice-session-rows.component.html',
})
export class InvoiceSessionRowsComponent {
  @Input() sessions: InvoiceSession[] = [];
  @Input() aggregate: InvoiceLineAggregate | null = null;
  @Input() unitAmountMinor: number | null = null;

  readonly formatSessionDate = formatSessionDate;
  readonly formatSessionBooking = formatSessionBooking;
  readonly formatSessionQuantity = formatSessionQuantity;
  readonly formatUnitAmount = formatUnitAmount;
}
