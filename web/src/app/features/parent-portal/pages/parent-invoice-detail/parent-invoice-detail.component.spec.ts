import { of, throwError, Subject } from 'rxjs';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { ActivatedRoute } from '@angular/router';

import { ParentInvoiceDetailComponent } from './parent-invoice-detail.component';
import { ParentInvoicesApiService } from '../../data/parent-invoices-api.service';
import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { ParentInvoiceDetail } from '../../models/parent-invoices.models';

function makeDetail(overrides: Partial<ParentInvoiceDetail> = {}): ParentInvoiceDetail {
  return {
    invoiceId: 'inv-1',
    invoiceKind: 'monthly',
    invoiceNumber: 'INV-001',
    invoiceNumberDisplay: 'INV-001',
    childId: 'c-1',
    childName: 'Ada Lovelace',
    billingMonth: '2026-05',
    period: { startDate: '2026-05-01', endDate: '2026-05-31' },
    status: 'issued',
    dueStatus: 'due',
    currencyCode: 'gbp',
    subtotalMinor: 45000,
    fundedDeductionMinor: 5000,
    totalDueMinor: 40000,
    amountPaidMinor: 0,
    issuedAt: '2026-05-28T00:00:00Z',
    dueAt: '2026-06-01T00:00:00Z',
    paidAt: null,
    paymentFailedAt: null,
    paymentStatusUpdatedAt: null,
    calculation: null,
    lines: [],
    ...overrides,
  };
}

describe('ParentInvoiceDetailComponent', () => {
  let fixture: ComponentFixture<ParentInvoiceDetailComponent>;
  let component: ParentInvoiceDetailComponent;
  let apiMock: jasmine.SpyObj<ParentInvoicesApiService>;
  let errorMapperMock: jasmine.SpyObj<ApiErrorMapper>;

  function createTestBed(routeId: string | null = 'inv-1') {
    apiMock = jasmine.createSpyObj('ParentInvoicesApiService', ['getInvoice', 'createCheckoutSession']);
    errorMapperMock = jasmine.createSpyObj('ApiErrorMapper', ['mapAndHandle']);

    TestBed.configureTestingModule({
      imports: [ParentInvoiceDetailComponent],
      providers: [
        { provide: ParentInvoicesApiService, useValue: apiMock },
        { provide: ApiErrorMapper, useValue: errorMapperMock },
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { paramMap: { get: (key: string) => key === 'invoiceId' ? routeId : null } } },
        },
      ],
    });

    fixture = TestBed.createComponent(ParentInvoiceDetailComponent);
    component = fixture.componentInstance;
  }

  it('loads detail on init', () => {
    createTestBed('inv-1');
    const detail = makeDetail();
    apiMock.getInvoice.and.returnValue(of(detail));
    fixture.detectChanges();

    expect(apiMock.getInvoice).toHaveBeenCalledWith('inv-1');
    expect(component.detail?.invoiceId).toBe('inv-1');
  });

  it('shows error for missing invoice id', () => {
    createTestBed(null);
    fixture.detectChanges();

    expect(component.errorMessage).toBe('Invoice ID is missing.');
  });

  it('shows error for API failure', () => {
    createTestBed('inv-1');
    apiMock.getInvoice.and.returnValue(throwError(() => ({ error: { code: 'not_found', message: 'Invoice not found' } })));
    errorMapperMock.mapAndHandle.and.returnValue({ code: 'not_found', message: 'Invoice not found', requestId: null, fieldErrors: {} });
    fixture.detectChanges();

    expect(component.errorMessage).toContain('Invoice not found');
  });

  it('renders parent-safe fields', () => {
    createTestBed('inv-1');
    const detail = makeDetail({
      lines: [
        { lineKind: 'core', description: 'Core sessions', sortOrder: 1, quantityMinutes: 450, unitAmountMinor: 100, lineAmountMinor: 45000 },
      ],
    });
    apiMock.getInvoice.and.returnValue(of(detail));
    fixture.detectChanges();

    const text = fixture.debugElement.nativeElement.textContent;
    expect(text).toContain('INV-001');
    expect(text).toContain('Ada Lovelace');
    expect(text).toContain('May 2026');
    expect(text).toContain('Core sessions');
  });

  it('does not render manager-only fields', () => {
    createTestBed('inv-1');
    const detail = makeDetail();
    apiMock.getInvoice.and.returnValue(of(detail));
    fixture.detectChanges();

    const text = fixture.debugElement.nativeElement.textContent.toLowerCase();
    expect(text).not.toContain('stripe');
    expect(text).not.toContain('webhook');
    expect(text).not.toContain('generated run');
    expect(text).not.toContain('adjustment reason');
  });

  it('shows pay action for payable invoice', () => {
    createTestBed('inv-1');
    const detail = makeDetail({ status: 'issued', totalDueMinor: 40000, amountPaidMinor: 0 });
    apiMock.getInvoice.and.returnValue(of(detail));
    fixture.detectChanges();

    const payBtn = fixture.debugElement.query(By.css('button'));
    expect(payBtn).toBeTruthy();
    expect(payBtn.nativeElement.textContent.trim()).toContain('Pay now');
  });

  it('hides pay action for paid invoice', () => {
    createTestBed('inv-1');
    const detail = makeDetail({ status: 'paid', dueStatus: 'paid', totalDueMinor: 40000, amountPaidMinor: 40000 });
    apiMock.getInvoice.and.returnValue(of(detail));
    fixture.detectChanges();

    const payBtn = fixture.debugElement.query(By.css('button'));
    expect(payBtn).toBeNull();
  });

  it('hides pay action for zero-balance invoice', () => {
    createTestBed('inv-1');
    const detail = makeDetail({ status: 'issued', totalDueMinor: 0, amountPaidMinor: 0 });
    apiMock.getInvoice.and.returnValue(of(detail));
    fixture.detectChanges();

    const payBtn = fixture.debugElement.query(By.css('button'));
    expect(payBtn).toBeNull();
  });

  it('renders lines as received from service (already sorted)', () => {
    createTestBed('inv-1');
    const detail = makeDetail({
      lines: [
        { lineKind: 'core', description: 'Core', sortOrder: 1, quantityMinutes: null, unitAmountMinor: null, lineAmountMinor: 45000 },
        { lineKind: 'extras', description: 'Extras', sortOrder: 2, quantityMinutes: null, unitAmountMinor: null, lineAmountMinor: 5000 },
      ],
    });
    apiMock.getInvoice.and.returnValue(of(detail));
    fixture.detectChanges();

    expect(component.detail?.lines[0].description).toBe('Core');
    expect(component.detail?.lines[1].description).toBe('Extras');
  });

  it('handles checkout creation and redirects', () => {
    createTestBed('inv-1');
    const detail = makeDetail({ status: 'issued' });
    apiMock.getInvoice.and.returnValue(of(detail));
    apiMock.createCheckoutSession.and.returnValue(of({
      checkoutSessionId: 'cs-1',
      checkoutUrl: 'https://checkout.stripe.com/session',
      paymentAttemptId: 'pa-1',
    }));

    fixture.detectChanges();
    spyOn(component, 'redirectTo');
    component.startPayment();
    expect(apiMock.createCheckoutSession).toHaveBeenCalledWith('inv-1');
    expect(component.redirectTo).toHaveBeenCalledWith('https://checkout.stripe.com/session');
  });

  it('handles checkout error', () => {
    createTestBed('inv-1');
    const detail = makeDetail({ status: 'issued' });
    apiMock.getInvoice.and.returnValue(of(detail));
    apiMock.createCheckoutSession.and.returnValue(throwError(() => ({ error: { code: 'conflict', message: 'Already paid' } })));
    errorMapperMock.mapAndHandle.and.returnValue({ code: 'conflict', message: 'Already paid', requestId: null, fieldErrors: {} });

    fixture.detectChanges();
    component.startPayment();
    expect(component.isPaying).toBeFalse();
    expect(component.errorMessage).toContain('Already paid');
  });
});
