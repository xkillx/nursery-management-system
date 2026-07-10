import { HttpClientTestingModule } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { ManagerInvoicesComponent } from './manager-invoices.component';
import { ManagerInvoicesApiService } from '../../data/manager-invoices-api.service';
import { ManagerInvoiceListItem } from '../../models/manager-invoices.models';
import { of, throwError } from 'rxjs';
import { HttpErrorResponse } from '@angular/common/http';

const mockItems: ManagerInvoiceListItem[] = [
  {
    invoiceId: 'inv-1',
    invoiceKind: 'monthly',
    invoiceNumber: 'INV-202605-0001',
    invoiceNumberDisplay: 'INV-202605-0001',
    childId: 'c1',
    childName: 'Ben',
    billingMonth: '2026-05',
    period: null,
    status: 'issued',
    dueStatus: 'due',
    currencyCode: 'gbp',
    subtotalMinor: 33000,
    fundedDeductionMinor: 9000,
    totalDueMinor: 24000,
    amountPaidMinor: 0,
    dueAt: '2026-06-10T00:00:00Z',
    issuedAt: '2026-06-09T12:00:00Z',
    paidAt: null,
    paymentFailedAt: null,
    paymentStatusUpdatedAt: null,
    generatedRunId: null,
    generatedRunStatus: null,
    generatedRunStartedAt: null,
    generatedRunCompletedAt: null,
    generatedRunExceptionCount: null,
    photoUrl: null,
    createdAt: '2026-06-09T10:00:00Z',
    updatedAt: '2026-06-09T12:00:00Z',
  },
  {
    invoiceId: 'inv-2',
    invoiceKind: 'monthly',
    invoiceNumber: null,
    invoiceNumberDisplay: '',
    childId: 'c2',
    childName: 'Alice',
    billingMonth: '2026-05',
    period: null,
    status: 'draft',
    dueStatus: 'not_due',
    currencyCode: 'gbp',
    subtotalMinor: 10000,
    fundedDeductionMinor: 0,
    totalDueMinor: 10000,
    amountPaidMinor: 0,
    dueAt: null,
    issuedAt: null,
    paidAt: null,
    paymentFailedAt: null,
    paymentStatusUpdatedAt: null,
    generatedRunId: null,
    generatedRunStatus: null,
    generatedRunStartedAt: null,
    generatedRunCompletedAt: null,
    generatedRunExceptionCount: null,
    photoUrl: null,
    createdAt: '2026-06-09T10:00:00Z',
    updatedAt: '2026-06-09T10:00:00Z',
  },
];

describe('ManagerInvoicesComponent', () => {
  let fixture: ComponentFixture<ManagerInvoicesComponent>;
  let apiService: jasmine.SpyObj<ManagerInvoicesApiService>;

  beforeEach(async () => {
    const spy = jasmine.createSpyObj('ManagerInvoicesApiService', ['listInvoices']);
    spy.listInvoices.and.returnValue(of({ items: mockItems, total: 2, page: 1, page_size: 50 }));

    await TestBed.configureTestingModule({
      imports: [ManagerInvoicesComponent, HttpClientTestingModule],
      providers: [
        provideRouter([]),
        ApiErrorMapper,
        { provide: ManagerInvoicesApiService, useValue: spy },
      ],
    }).compileComponents();

    apiService = TestBed.inject(ManagerInvoicesApiService) as jasmine.SpyObj<ManagerInvoicesApiService>;
    fixture = TestBed.createComponent(ManagerInvoicesComponent);
  });

  it('loads list on init with default range and all status', () => {
    fixture.detectChanges();

    expect(apiService.listInvoices).toHaveBeenCalledTimes(1);
    const call = apiService.listInvoices.calls.argsFor(0)[0];
    expect(call.status).toBe('all');
    expect(call.billingMonthFrom).toMatch(/^\d{4}-\d{2}$/);
    expect(call.billingMonthTo).toMatch(/^\d{4}-\d{2}$/);
    expect(call.limit).toBe(50);
    expect(call.offset).toBe(0);
  });

  it('renders invoice identity, child name, status, due status, and net due', () => {
    fixture.detectChanges();
    const text = fixture.nativeElement.textContent;

    expect(text).toContain('INV-202605-0001');
    expect(text).toContain('Ben');
    expect(text).toContain('Draft invoice');
    expect(text).toContain('Alice');
    expect(text).toContain('Issued');
    expect(text).toContain('Due');
    expect(text).toContain('Not due');
    expect(text).toContain('Draft');
    expect(text).toContain('£240.00');
  });

  it('renders View links pointing to detail routes', () => {
    fixture.detectChanges();
    const links: HTMLAnchorElement[] = fixture.nativeElement.querySelectorAll('a');

    const viewLinks = Array.from(links).filter((a) => a.textContent?.trim() === 'View');
    expect(viewLinks.length).toBeGreaterThanOrEqual(1);
    expect(viewLinks[0].href).toContain('/manager/invoices/');
  });

  it('reloads and resets offset when status changes', () => {
    fixture.detectChanges();
    apiService.listInvoices.calls.reset();

    fixture.componentInstance.onStatusChange('issued');
    expect(apiService.listInvoices).toHaveBeenCalledTimes(1);
    const call = apiService.listInvoices.calls.argsFor(0)[0];
    expect(call.status).toBe('issued');
    expect(call.offset).toBe(0);
  });

  it('reloads and resets offset when range preset changes', () => {
    fixture.detectChanges();
    apiService.listInvoices.calls.reset();

    fixture.componentInstance.onRangePreset('this');
    expect(apiService.listInvoices).toHaveBeenCalledTimes(1);
    const call = apiService.listInvoices.calls.argsFor(0)[0];
    expect(call.billingMonthFrom).toBeDefined();
    expect(call.billingMonthTo).toBeDefined();
    expect(call.offset).toBe(0);
  });

  it('switches to custom range and reloads', () => {
    fixture.detectChanges();
    apiService.listInvoices.calls.reset();

    fixture.componentInstance.selectedBillingMonthFrom = '2026-01';
    fixture.componentInstance.selectedBillingMonthTo = '2026-06';
    fixture.componentInstance.onCustomRangeChange();
    expect(apiService.listInvoices).toHaveBeenCalledTimes(1);
    const call = apiService.listInvoices.calls.argsFor(0)[0];
    expect(call.billingMonthFrom).toBe('2026-01');
    expect(call.billingMonthTo).toBe('2026-06');
    expect(call.offset).toBe(0);
  });

  it('displays funded deduction as subtraction when positive', () => {
    fixture.detectChanges();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('£90.00');
  });

  it('shows empty state when no invoices', () => {
    apiService.listInvoices.and.returnValue(of({ items: [], total: 0, page: 1, page_size: 50 }));
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('No invoices found');
  });

  it('shows error message on API error', () => {
    apiService.listInvoices.and.returnValue(throwError(() => new HttpErrorResponse({
      error: { code: 'internal_error', message: 'Server error', request_id: 'req-1' },
      status: 500,
    })));

    fixture.detectChanges();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Something went wrong. Try again.');
    expect(text).toContain('req-1');
  });

  it('does not contain forbidden action labels', () => {
    fixture.detectChanges();
    const text = fixture.nativeElement.textContent;
    const forbidden = ['Edit', 'Regenerate', 'Delete', 'Adjust', 'Checkout', 'Retry payment'];

    for (const label of forbidden) {
      expect(text).not.toContain(label);
    }
  });

  it('does not send status param when filter is all', () => {
    fixture.detectChanges();
    const call = apiService.listInvoices.calls.argsFor(0)[0];
    expect(call.status).toBe('all');
  });

  it('shows Unpaid cue for issued invoice with zero paid', () => {
    const unpaidItem: ManagerInvoiceListItem = {
      ...mockItems[0],
      invoiceId: 'inv-u1',
      status: 'issued',
      dueStatus: 'due',
      amountPaidMinor: 0,
      paidAt: null,
    };
    apiService.listInvoices.and.returnValue(of({ items: [unpaidItem], total: 1, page: 1, page_size: 50 }));
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Unpaid');
  });

  it('shows Unpaid cue with Overdue for overdue invoice with zero paid', () => {
    const overdueItem: ManagerInvoiceListItem = {
      ...mockItems[0],
      invoiceId: 'inv-o1',
      status: 'overdue',
      dueStatus: 'overdue',
      amountPaidMinor: 0,
      paidAt: null,
    };
    apiService.listInvoices.and.returnValue(of({ items: [overdueItem], total: 1, page: 1, page_size: 50 }));
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Unpaid');
    expect(text).toContain('Overdue');
  });

  it('shows Paid cue for paid invoice with paid timestamp', () => {
    const paidItem: ManagerInvoiceListItem = {
      ...mockItems[0],
      invoiceId: 'inv-p1',
      status: 'paid',
      dueStatus: 'paid',
      amountPaidMinor: 24000,
      paidAt: '2026-06-09T15:00:00Z',
    };
    apiService.listInvoices.and.returnValue(of({ items: [paidItem], total: 1, page: 1, page_size: 50 }));
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Paid');
  });

  it('shows Payment failed cue for payment_failed invoice', () => {
    const failedItem: ManagerInvoiceListItem = {
      ...mockItems[0],
      invoiceId: 'inv-f1',
      status: 'payment_failed',
      dueStatus: 'due',
      amountPaidMinor: 0,
      paymentFailedAt: '2026-06-09T16:00:00Z',
    };
    apiService.listInvoices.and.returnValue(of({ items: [failedItem], total: 1, page: 1, page_size: 50 }));
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Payment failed');
  });

  it('shows Not issued for draft invoices', () => {
    fixture.detectChanges();
    const rows = fixture.nativeElement.querySelectorAll('tbody tr');
    const draftRow = rows[1];
    expect(draftRow.textContent).toContain('Not issued');
  });

  it('does not contain Checkout or Retry payment labels, and no Pay action button', () => {
    fixture.detectChanges();
    const text = fixture.nativeElement.textContent;
    expect(text).not.toContain('Checkout');
    expect(text).not.toContain('Retry payment');

    const buttons: HTMLButtonElement[] = Array.from(fixture.nativeElement.querySelectorAll('button'));
    const anchors: HTMLAnchorElement[] = Array.from(fixture.nativeElement.querySelectorAll('a'));
    const allElements: HTMLElement[] = [...buttons, ...anchors];
    const payActions = allElements.filter((el) => el.textContent?.trim() === 'Pay' || el.textContent?.trim() === 'Retry payment');
    expect(payActions.length).toBe(0);
  });
});
