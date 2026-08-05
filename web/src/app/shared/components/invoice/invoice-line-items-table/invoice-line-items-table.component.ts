import { ChangeDetectionStrategy, Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';

import { InvoiceSession, InvoiceLineAggregate } from '../../../models/invoice-session.models';
import { InvoiceSessionRowsComponent } from '../invoice-session-rows/invoice-session-rows.component';
import { formatUnitAmount } from '../invoice-session-rows/session-rows.formatters';

export interface InvoiceLineItemsTableLine {
  lineKind: string;
  description: string;
  quantityMinutes: number | null;
  unitAmountMinor: number | null;
  lineAmountMinor: number;
  fundingModel: string | null;
  sessions: InvoiceSession[];
  aggregate: InvoiceLineAggregate | null;
}

export function lineKindLabel(kind: string): string {
  return kind
    .split('_')
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(' ');
}

export function fundingModelLabel(model: string | null): string {
  if (model === 'term_time_only') return 'Term-time funding';
  if (model === 'stretched') return 'Stretched funding';
  return 'Funded hours';
}

@Component({
  selector: 'app-invoice-line-items-table',
  imports: [CommonModule, InvoiceSessionRowsComponent],
  changeDetection: ChangeDetectionStrategy.OnPush,
  templateUrl: './invoice-line-items-table.component.html',
})
export class InvoiceLineItemsTableComponent {
  @Input() lines: InvoiceLineItemsTableLine[] = [];
  @Input() emptyMessage = 'No line items have been generated yet.';

  readonly formatUnitAmount = formatUnitAmount;
  readonly lineKindLabel = lineKindLabel;
  readonly fundingModelLabel = fundingModelLabel;
}
