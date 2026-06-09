import { HttpClientTestingModule } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, provideRouter } from '@angular/router';
import { of, throwError } from 'rxjs';
import { HttpErrorResponse } from '@angular/common/http';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { ManagerInvoiceDetailComponent } from './manager-invoice-detail.component';
import { ManagerInvoicesApiService } from '../../data/manager-invoices-api.service';
import { ManagerInvoiceDetail } from '../../models/manager-invoices.models';

const issuedDetail: ManagerInvoiceDetail = {
  invoiceId: 'inv-1',
  invoiceKind: 'monthly',
  invoiceNumber: 'INV-202605-0001',
  invoiceNumberDisplay: 'INV-202605-0001',
  childId: 'c1',
  childName: 'Ben',
  billingMonth: '2026-05',
  period: { startDate: '2026-05-01', endDate: '2026-05-31' },
  status: 'issued',
  dueStatus: 'due',
  currencyCode: 'gbp',
  subtotalMinor: 33000,
  fundedDeductionMinor: 9000,
  totalDueMinor: 24000,
  amountPaidMinor: 0,
  issuedAt: '2026-06-09T12:00:00Z',
  lockedAt: '2026-06-09T12:00:00Z',
  dueAt: '2026-06-10T00:00:00Z',
  paidAt: null,
  paymentFailedAt: null,
  paymentStatusUpdatedAt: null,
  adjustsInvoiceId: null,
  adjustmentReasonCode: null,
  adjustmentReasonNote: null,
  generatedRunId: 'run-1',
  generatedRunStatus: 'completed',
  generatedRunStartedAt: '2026-06-09T11:00:00Z',
  generatedRunCompletedAt: '2026-06-09T11:05:00Z',
  generatedRunExceptionCount: 1,
  generatedRunExceptions: [
    { childId: 'c2', childName: 'Alice', blockerCodes: ['incomplete_attendance'] },
  ],
  calculation: {
    coreHourlyRateMinor: 25,
    rawAttendedMinutes: 1315,
    roundedAttendedMinutes: 1320,
    fundedAllowanceMinutes: 360,
    fundedDeductionMinutes: 360,
    coreBillableMinutes: 960,
    includedSessionCount: 4,
    coreSubtotalMinor: 24000,
    extrasTotalMinor: 9000,
  },
  lines: [
    {
      lineId: 'l1',
      lineKind: 'core_childcare',
      description: 'Core childcare',
      sortOrder: 1,
      quantityMinutes: 1320,
      unitAmountMinor: 25,
      lineAmountMinor: 33000,
      rawAttendedMinutes: null,
      roundedAttendedMinutes: null,
      fundedAllowanceMinutes: null,
      fundedDeductionMinutes: null,
      coreBillableMinutes: null,
      sessionCount: null,
    },
    {
      lineId: 'l2',
      lineKind: 'funded_deduction',
      description: 'Funded deduction',
      sortOrder: 2,
      quantityMinutes: 360,
      unitAmountMinor: 25,
      lineAmountMinor: -9000,
      rawAttendedMinutes: null,
      roundedAttendedMinutes: null,
      fundedAllowanceMinutes: null,
      fundedDeductionMinutes: null,
      coreBillableMinutes: null,
      sessionCount: null,
    },
  ],
  createdAt: '2026-06-09T10:00:00Z',
  updatedAt: '2026-06-09T12:00:00Z',
};

const draftDetail: ManagerInvoiceDetail = {
  ...issuedDetail,
  invoiceId: 'inv-3',
  invoiceNumber: null,
  invoiceNumberDisplay: '',
  status: 'draft',
  dueStatus: 'not_due',
  lockedAt: null,
  issuedAt: null,
  dueAt: null,
  lines: [],
  generatedRunExceptions: [],
  calculation: null,
};

describe('ManagerInvoiceDetailComponent', () => {
  let fixture: ComponentFixture<ManagerInvoiceDetailComponent>;
  let apiService: jasmine.SpyObj<ManagerInvoicesApiService>;

  function createFixture(detail: ManagerInvoiceDetail = issuedDetail) {
    const spy = jasmine.createSpyObj('ManagerInvoicesApiService', ['getInvoice']);
    spy.getInvoice.and.returnValue(of(detail));

    TestBed.configureTestingModule({
      imports: [ManagerInvoiceDetailComponent, HttpClientTestingModule],
      providers: [
        provideRouter([]),
        ApiErrorMapper,
        { provide: ManagerInvoicesApiService, useValue: spy },
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { paramMap: { get: (key: string) => key === 'invoiceId' ? 'inv-1' : null } } },
        },
      ],
    });

    apiService = TestBed.inject(ManagerInvoicesApiService) as jasmine.SpyObj<ManagerInvoicesApiService>;
    fixture = TestBed.createComponent(ManagerInvoiceDetailComponent);
    fixture.detectChanges();
  }

  it('fetches detail using route param on init', () => {
    createFixture();
    expect(apiService.getInvoice).toHaveBeenCalledWith('inv-1');
  });

  it('renders header identity, child name, billing month, status, and due status', () => {
    createFixture();
    const text = fixture.nativeElement.textContent;

    expect(text).toContain('INV-202605-0001');
    expect(text).toContain('Ben');
    expect(text).toContain('May 2026');
    expect(text).toContain('Issued');
    expect(text).toContain('Due');
  });

  it('renders funded summary calculation quantities', () => {
    createFixture();
    const text = fixture.nativeElement.textContent;

    expect(text).toContain('22h');
    expect(text).toContain('6h');
    expect(text).toContain('16h');
    expect(text).toContain('4');
  });

  it('renders line items in sorted order', () => {
    createFixture();
    const table = fixture.nativeElement.querySelector('table');
    expect(table).toBeTruthy();
    const tableText = table.textContent;

    const coreIndex = tableText.indexOf('Core childcare');
    const fundedIndex = tableText.indexOf('Funded deduction');
    expect(coreIndex).toBeLessThan(fundedIndex);
  });

  it('shows locked/immutable notice for issued invoices', () => {
    createFixture();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Issued invoice locked');
    expect(text).toContain('Locked');
  });

  it('shows read-only review for draft invoices', () => {
    createFixture(draftDetail);
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Read-only review');
    expect(text).not.toContain('Issued invoice locked');
  });

  it('does not contain forbidden action labels as standalone words', () => {
    createFixture();
    const text = fixture.nativeElement.textContent;
    const forbidden = ['Edit', 'Regenerate', 'Delete', 'Adjust', 'Checkout', 'Retry payment', 'Pay'];

    for (const label of forbidden) {
      expect(text).not.toContain(label);
    }
  });

  it('shows net due and funded deduction summary', () => {
    createFixture();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('£240.00');
    expect(text).toContain('£330.00');
  });

  it('shows generation context exceptions', () => {
    createFixture();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Alice');
    expect(text).toContain('incomplete_attendance');
  });

  it('shows back link to invoices list', () => {
    createFixture();
    const links: HTMLAnchorElement[] = fixture.nativeElement.querySelectorAll('a');
    const backLink = Array.from(links).find((a) => a.textContent?.includes('Back to invoices'));
    expect(backLink).toBeTruthy();
    expect(backLink!.href).toContain('/staff/manager/invoices');
  });
});

describe('ManagerInvoiceDetailComponent error handling', () => {
  let fixture: ComponentFixture<ManagerInvoiceDetailComponent>;

  beforeEach(() => {
    const spy = jasmine.createSpyObj('ManagerInvoicesApiService', ['getInvoice']);
    spy.getInvoice.and.returnValue(throwError(() => new HttpErrorResponse({
      error: { code: 'not_found', message: 'Invoice not found', request_id: 'req-99' },
      status: 404,
    })));

    TestBed.configureTestingModule({
      imports: [ManagerInvoiceDetailComponent, HttpClientTestingModule],
      providers: [
        provideRouter([]),
        ApiErrorMapper,
        { provide: ManagerInvoicesApiService, useValue: spy },
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { paramMap: { get: (key: string) => key === 'invoiceId' ? 'inv-1' : null } } },
        },
      ],
    });

    fixture = TestBed.createComponent(ManagerInvoiceDetailComponent);
    fixture.detectChanges();
  });

  it('shows error message on API error with request ID', () => {
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Invoice not found');
    expect(text).toContain('req-99');
  });
});
