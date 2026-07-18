import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';

export interface InvoiceLineItem {
  description: string;
  category: 'session_fees' | 'meals' | 'funding_deductions';
  quantity: number;
  unitPrice: number;
  total: number;
}

export interface InvoicePreviewData {
  nurseryName: string;
  nurseryAddress: string;
  invoiceNumber: string;
  date: string;
  dueDate: string;
  childName: string;
  parentName: string;
  lineItems: InvoiceLineItem[];
  totalPayable: number;
}

@Component({
  selector: 'app-invoice-preview',
  imports: [CommonModule],
  templateUrl: './invoice-preview.component.html',
  styles: `
    @media print {
      :host {
        display: block;
      }
      .invoice-no-print {
        display: none !important;
      }
    }
  `,
})
export class InvoicePreviewComponent {
  @Input() invoice!: InvoicePreviewData;

  get sessionFees(): InvoiceLineItem[] {
    return (this.invoice?.lineItems ?? []).filter((i) => i.category === 'session_fees');
  }

  get meals(): InvoiceLineItem[] {
    return (this.invoice?.lineItems ?? []).filter((i) => i.category === 'meals');
  }

  get fundingDeductions(): InvoiceLineItem[] {
    return (this.invoice?.lineItems ?? []).filter((i) => i.category === 'funding_deductions');
  }

  formatCurrency(amount: number): string {
    return `£${amount.toFixed(2)}`;
  }
}
