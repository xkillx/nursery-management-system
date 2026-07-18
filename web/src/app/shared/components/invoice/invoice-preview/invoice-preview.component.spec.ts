import { InvoicePreviewComponent, InvoicePreviewData } from './invoice-preview.component';

describe('InvoicePreviewComponent', () => {
  let component: InvoicePreviewComponent;
  const mockInvoice: InvoicePreviewData = {
    nurseryName: 'Sunshine Nursery',
    nurseryAddress: '123 High Street\nLondon\nSW1A 1AA',
    invoiceNumber: 'INV-2026-001',
    date: '2026-07-18',
    dueDate: '2026-08-01',
    childName: 'James Smith',
    parentName: 'Sarah Smith',
    lineItems: [
      { description: 'Full Day Session (Mon)', category: 'session_fees', quantity: 4, unitPrice: 55.00, total: 220.00 },
      { description: 'Half Day Session (Fri)', category: 'session_fees', quantity: 4, unitPrice: 35.00, total: 140.00 },
      { description: 'Lunch', category: 'meals', quantity: 4, unitPrice: 3.50, total: 14.00 },
      { description: 'Snacks', category: 'meals', quantity: 8, unitPrice: 1.00, total: 8.00 },
      { description: '15hr Funding Deduction', category: 'funding_deductions', quantity: 1, unitPrice: -93.75, total: -93.75 },
    ],
    totalPayable: 288.25,
  };

  beforeEach(() => {
    component = new InvoicePreviewComponent();
    component.invoice = mockInvoice;
  });

  it('renders invoice header with nursery info', () => {
    expect(component.invoice.nurseryName).toBe('Sunshine Nursery');
    expect(component.invoice.nurseryAddress).toContain('London');
  });

  it('renders child and parent details', () => {
    expect(component.invoice.childName).toBe('James Smith');
    expect(component.invoice.parentName).toBe('Sarah Smith');
  });

  it('groups line items by session fees', () => {
    expect(component.sessionFees.length).toBe(2);
    expect(component.sessionFees[0].description).toContain('Full Day');
  });

  it('groups line items by meals', () => {
    expect(component.meals.length).toBe(2);
    expect(component.meals[0].description).toBe('Lunch');
  });

  it('groups line items by funding deductions', () => {
    expect(component.fundingDeductions.length).toBe(1);
    expect(component.fundingDeductions[0].description).toContain('15hr');
  });

  it('displays total payable', () => {
    expect(component.invoice.totalPayable).toBe(288.25);
  });

  it('formats currency correctly', () => {
    expect(component.formatCurrency(100)).toBe('£100.00');
    expect(component.formatCurrency(0)).toBe('£0.00');
    expect(component.formatCurrency(93.75)).toBe('£93.75');
  });

  it('handles empty line items gracefully', () => {
    component.invoice = { ...mockInvoice, lineItems: [] };
    expect(component.sessionFees).toEqual([]);
    expect(component.meals).toEqual([]);
    expect(component.fundingDeductions).toEqual([]);
  });

  it('handles missing invoice without error', () => {
    (component as { invoice: unknown }).invoice = undefined;
    expect(component.sessionFees).toEqual([]);
    expect(component.meals).toEqual([]);
    expect(component.fundingDeductions).toEqual([]);
  });
});
