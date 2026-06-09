import { of, throwError, Subject } from 'rxjs';
import { ComponentFixture, TestBed, fakeAsync, tick, discardPeriodicTasks } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { ActivatedRoute, Router } from '@angular/router';

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
  let routerMock: jasmine.SpyObj<Router>;

  function createTestBed(options: {
    routeId?: string | null;
    queryParams?: Record<string, string>;
  } = {}) {
    const routeId = options.routeId === undefined ? 'inv-1' : options.routeId;
    const queryParams = options.queryParams ?? {};

    apiMock = jasmine.createSpyObj('ParentInvoicesApiService', ['getInvoice', 'createCheckoutSession']);
    errorMapperMock = jasmine.createSpyObj('ApiErrorMapper', ['mapAndHandle']);
    routerMock = jasmine.createSpyObj('Router', ['navigate']);

    TestBed.configureTestingModule({
      imports: [ParentInvoiceDetailComponent],
      providers: [
        { provide: ParentInvoicesApiService, useValue: apiMock },
        { provide: ApiErrorMapper, useValue: errorMapperMock },
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: {
              paramMap: { get: (key: string) => key === 'invoiceId' ? routeId : null },
              queryParamMap: {
                get: (key: string) => queryParams[key] ?? null,
              },
              queryParams,
            },
          },
        },
        { provide: Router, useValue: routerMock },
      ],
    });

    fixture = TestBed.createComponent(ParentInvoiceDetailComponent);
    component = fixture.componentInstance;
  }

  it('loads detail on init', () => {
    createTestBed();
    const detail = makeDetail();
    apiMock.getInvoice.and.returnValue(of(detail));
    fixture.detectChanges();

    expect(apiMock.getInvoice).toHaveBeenCalledWith('inv-1');
    expect(component.detail?.invoiceId).toBe('inv-1');
  });

  it('shows error for missing invoice id', () => {
    createTestBed({ routeId: null });
    fixture.detectChanges();

    expect(component.errorMessage).toBe('Invoice ID is missing.');
  });

  it('shows error for API failure', () => {
    createTestBed();
    apiMock.getInvoice.and.returnValue(throwError(() => ({ error: { code: 'not_found', message: 'Invoice not found' } })));
    errorMapperMock.mapAndHandle.and.returnValue({ code: 'not_found', message: 'Invoice not found', requestId: null, fieldErrors: {} });
    fixture.detectChanges();

    expect(component.errorMessage).toContain('Invoice not found');
  });

  it('renders parent-safe fields', () => {
    createTestBed();
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
    createTestBed();
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
    createTestBed();
    const detail = makeDetail({ status: 'issued', totalDueMinor: 40000, amountPaidMinor: 0 });
    apiMock.getInvoice.and.returnValue(of(detail));
    fixture.detectChanges();

    const payBtn = fixture.debugElement.query(By.css('button'));
    expect(payBtn).toBeTruthy();
    expect(payBtn.nativeElement.textContent.trim()).toContain('Pay now');
  });

  it('hides pay action for paid invoice', () => {
    createTestBed();
    const detail = makeDetail({ status: 'paid', dueStatus: 'paid', totalDueMinor: 40000, amountPaidMinor: 40000 });
    apiMock.getInvoice.and.returnValue(of(detail));
    fixture.detectChanges();

    const payBtn = fixture.debugElement.query(By.css('button'));
    expect(payBtn).toBeNull();
  });

  it('hides pay action for zero-balance invoice', () => {
    createTestBed();
    const detail = makeDetail({ status: 'issued', totalDueMinor: 0, amountPaidMinor: 0 });
    apiMock.getInvoice.and.returnValue(of(detail));
    fixture.detectChanges();

    const payBtn = fixture.debugElement.query(By.css('button'));
    expect(payBtn).toBeNull();
  });

  it('renders lines as received from service (already sorted)', () => {
    createTestBed();
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
    createTestBed();
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
    createTestBed();
    const detail = makeDetail({ status: 'issued' });
    apiMock.getInvoice.and.returnValue(of(detail));
    apiMock.createCheckoutSession.and.returnValue(throwError(() => ({ error: { code: 'conflict', message: 'Already paid' } })));
    errorMapperMock.mapAndHandle.and.returnValue({ code: 'conflict', message: 'Already paid', requestId: null, fieldErrors: {} });

    fixture.detectChanges();
    component.startPayment();
    expect(component.isPaying).toBeFalse();
    expect(component.errorMessage).toContain('Already paid');
  });

  describe('return state — success', () => {
    it('parses checkout=success and clears params', () => {
      createTestBed({ queryParams: { checkout: 'success', session_id: 'cs_test' } });
      const detail = makeDetail({ status: 'issued' });
      apiMock.getInvoice.and.returnValue(of(detail));
      fixture.detectChanges();

      expect(component.returnKind).toBe('success');
      expect(routerMock.navigate).toHaveBeenCalledWith(
        [],
        jasmine.objectContaining({ replaceUrl: true, relativeTo: jasmine.any(Object) }),
      );
    });

    it('shows processing state when status remains issued after success return', fakeAsync(() => {
      createTestBed({ queryParams: { checkout: 'success' } });
      const detail = makeDetail({ status: 'issued' });
      apiMock.getInvoice.and.returnValue(of(detail));
      fixture.detectChanges();

      expect(component.returnDisplayState).toBe('processing');
      expect(component.isPolling).toBeTrue();
      expect(fixture.debugElement.nativeElement.textContent).toContain('Payment is still processing');

      discardPeriodicTasks();
    }));

    it('polls and resolves to paid', fakeAsync(() => {
      createTestBed({ queryParams: { checkout: 'success' } });
      const issued = makeDetail({ status: 'issued' });
      const paid = makeDetail({ status: 'paid', dueStatus: 'paid', amountPaidMinor: 40000 });

      apiMock.getInvoice.and.returnValue(of(issued));
      fixture.detectChanges();

      expect(component.returnDisplayState).toBe('processing');

      apiMock.getInvoice.and.returnValue(of(paid));
      tick(2000);
      fixture.detectChanges();

      expect(component.returnDisplayState).toBe('paid');
      expect(component.isPolling).toBeFalse();
      expect(fixture.debugElement.nativeElement.textContent).toContain('Payment received');

      discardPeriodicTasks();
    }));

    it('polls and resolves to payment_failed', fakeAsync(() => {
      createTestBed({ queryParams: { checkout: 'success' } });
      const issued = makeDetail({ status: 'issued' });
      const failed = makeDetail({ status: 'payment_failed' });

      apiMock.getInvoice.and.returnValue(of(issued));
      fixture.detectChanges();

      apiMock.getInvoice.and.returnValue(of(failed));
      tick(2000);
      fixture.detectChanges();

      expect(component.returnDisplayState).toBe('failed');
      expect(component.isPolling).toBeFalse();
      expect(fixture.debugElement.nativeElement.textContent).toContain('Payment was not completed');

      discardPeriodicTasks();
    }));

    it('stops polling after max duration', fakeAsync(() => {
      createTestBed({ queryParams: { checkout: 'success' } });
      const issued = makeDetail({ status: 'issued' });
      apiMock.getInvoice.and.returnValue(of(issued));
      fixture.detectChanges();

      tick(20000);
      fixture.detectChanges();

      expect(component.isPolling).toBeFalse();
      expect(component.returnDisplayState).toBe('processing');

      discardPeriodicTasks();
    }));

    it('immediately shows paid when first load is paid after success return', () => {
      createTestBed({ queryParams: { checkout: 'success' } });
      const paid = makeDetail({ status: 'paid', dueStatus: 'paid', amountPaidMinor: 40000 });
      apiMock.getInvoice.and.returnValue(of(paid));
      fixture.detectChanges();

      expect(component.returnDisplayState).toBe('paid');
      expect(component.isPolling).toBeFalse();
    });
  });

  describe('return state — cancelled', () => {
    it('parses checkout=cancelled and shows cancelled state', () => {
      createTestBed({ queryParams: { checkout: 'cancelled' } });
      const detail = makeDetail({ status: 'issued' });
      apiMock.getInvoice.and.returnValue(of(detail));
      fixture.detectChanges();

      expect(component.returnKind).toBe('cancelled');
      expect(component.returnDisplayState).toBe('cancelled');
      expect(component.isPolling).toBeFalse();
      expect(fixture.debugElement.nativeElement.textContent).toContain('Payment canceled');
    });

    it('parses checkout=canceled (alternative spelling)', () => {
      createTestBed({ queryParams: { checkout: 'canceled' } });
      const detail = makeDetail({ status: 'issued' });
      apiMock.getInvoice.and.returnValue(of(detail));
      fixture.detectChanges();

      expect(component.returnKind).toBe('cancelled');
      expect(component.returnDisplayState).toBe('cancelled');
    });

    it('shows pay action for payable cancelled return', () => {
      createTestBed({ queryParams: { checkout: 'cancelled' } });
      const detail = makeDetail({ status: 'issued', totalDueMinor: 40000, amountPaidMinor: 0 });
      apiMock.getInvoice.and.returnValue(of(detail));
      fixture.detectChanges();

      const payBtn = fixture.debugElement.query(By.css('button'));
      expect(payBtn).toBeTruthy();
      expect(payBtn.nativeElement.textContent.trim()).toContain('Pay now');
    });

    it('does not poll after cancelled return', () => {
      createTestBed({ queryParams: { checkout: 'cancelled' } });
      const detail = makeDetail({ status: 'issued' });
      apiMock.getInvoice.and.returnValue(of(detail));
      fixture.detectChanges();

      expect(component.isPolling).toBeFalse();
      expect(apiMock.getInvoice).toHaveBeenCalledTimes(1);
    });
  });

  describe('return state — normal load', () => {
    it('has no return banner when no query params', () => {
      createTestBed();
      const detail = makeDetail({ status: 'issued' });
      apiMock.getInvoice.and.returnValue(of(detail));
      fixture.detectChanges();

      expect(component.returnKind).toBe('none');
      expect(component.returnDisplayState).toBeNull();
      expect(routerMock.navigate).not.toHaveBeenCalled();
    });

    it('ignores unknown checkout value', () => {
      createTestBed({ queryParams: { checkout: 'unknown' } });
      const detail = makeDetail({ status: 'issued' });
      apiMock.getInvoice.and.returnValue(of(detail));
      fixture.detectChanges();

      expect(component.returnKind).toBe('none');
      expect(component.returnDisplayState).toBeNull();
    });
  });

  describe('return state — no sensitive data', () => {
    it('never exposes stripe session id in rendered content', () => {
      createTestBed({ queryParams: { checkout: 'success', session_id: 'cs_test_secret_123' } });
      const paid = makeDetail({ status: 'paid', dueStatus: 'paid', amountPaidMinor: 40000 });
      apiMock.getInvoice.and.returnValue(of(paid));
      fixture.detectChanges();

      const text = fixture.debugElement.nativeElement.textContent;
      expect(text).not.toContain('cs_test_secret_123');
      expect(text).not.toContain('session_id');
    });

    it('return alert never contains webhook or stripe diagnostic language', () => {
      createTestBed({ queryParams: { checkout: 'success' } });
      const detail = makeDetail({ status: 'issued' });
      apiMock.getInvoice.and.returnValue(of(detail));
      fixture.detectChanges();

      const text = fixture.debugElement.nativeElement.textContent.toLowerCase();
      expect(text).not.toContain('stripe');
      expect(text).not.toContain('webhook');
      expect(text).not.toContain('session');
    });
  });

  describe('polling cleanup', () => {
    it('stops polling on destroy', fakeAsync(() => {
      createTestBed({ queryParams: { checkout: 'success' } });
      const issued = makeDetail({ status: 'issued' });
      apiMock.getInvoice.and.returnValue(of(issued));
      fixture.detectChanges();

      expect(component.isPolling).toBeTrue();

      component.ngOnDestroy();
      expect(component.isPolling).toBeFalse();

      discardPeriodicTasks();
    }));

    it('stops polling on API error during poll', fakeAsync(() => {
      createTestBed({ queryParams: { checkout: 'success' } });
      const issued = makeDetail({ status: 'issued' });
      apiMock.getInvoice.and.returnValue(of(issued));
      fixture.detectChanges();

      errorMapperMock.mapAndHandle.and.returnValue({ code: 'internal', message: 'Server error', requestId: null, fieldErrors: {} });
      apiMock.getInvoice.and.returnValue(throwError(() => ({ error: { code: 'internal', message: 'Server error' } })));
      tick(2000);
      fixture.detectChanges();

      expect(component.isPolling).toBeFalse();
      expect(component.errorMessage).toContain('Server error');

      discardPeriodicTasks();
    }));
  });
});
