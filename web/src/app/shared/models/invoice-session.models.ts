export interface InvoiceSession {
  occurrenceDate: string;
  startMinutes: number | null;
  endMinutes: number | null;
  durationMinutes: number | null;
  sessionTypeName: string;
  sessionAmountMinor: number;
}

export interface InvoiceLineAggregate {
  description: string;
  sessionCount: number;
  quantityMinutes: number | null;
  totalMinor: number;
}
