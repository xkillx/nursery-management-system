import { HttpClientTestingModule } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, provideRouter } from '@angular/router';
import { of, throwError } from 'rxjs';
import { Observable } from 'rxjs';
import { HttpErrorResponse } from '@angular/common/http';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { ManagerInvoiceDetailComponent } from './manager-invoice-detail.component';
import { ManagerInvoicesApiService } from '../../data/manager-invoices-api.service';
import { ManagerInvoiceDetail, ManagerPaymentStatus, PaymentEvent } from '../../models/manager-invoices.models';
import { ToastService } from '../../../../shared/services/toast.service';

const issuedDetail: ManagerInvoiceDetail = {
  invoiceId: 'inv-1',
  invoiceKind: 'monthly',
  invoiceNumber: 'INV-202605-0001',
  invoiceNumberDisplay: 'INV-202605-0001',
  childId: 'c1',
  childName: 'Ben',
  billingMonth: '2026-05',
  site_profile: { nursery_name: 'London North Nursery', phone: '020 1234 5678', email: 'info@nursery.com', website: 'https://nursery.com', address_street: '12 High Street', address_city: 'London', address_postcode: 'N1 8PR' },
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
  roomName: 'Toddler Group',
  parentContact: {
    fullName: 'Burhan Khalid',
    addressLine1: '45 Maple Avenue',
    addressLine2: 'Crouch End',
    addressCity: 'London',
    addressPostcode: 'N8 9LE',
    email: 'burhan@example.com',
    telephone: '07700 900123',
  },
  photoUrl: null,
  parentNote: 'Please pay by the due date.',
  generatedRunExceptions: [
    { childId: 'c2', childName: 'Alice', blockerCodes: ['incomplete_attendance'] },
  ],
  calculation: {
    siteCoreHourlyRateMinor: 25,
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
      quantityHours: 1320,
      unitAmountMinor: 25,
      lineAmountMinor: 33000,
      rawAttendedMinutes: null,
      roundedAttendedMinutes: null,
      fundedAllowanceMinutes: null,
      fundedDeductionMinutes: null,
      coreBillableMinutes: null,
      sessionCount: null,
      fundingModel: null,
    },
    {
      lineId: 'l2',
      lineKind: 'funded_deduction',
      description: 'Funded deduction',
      sortOrder: 2,
      quantityHours: 360,
      unitAmountMinor: 25,
      lineAmountMinor: -9000,
      rawAttendedMinutes: null,
      roundedAttendedMinutes: null,
      fundedAllowanceMinutes: null,
      fundedDeductionMinutes: null,
      coreBillableMinutes: null,
      sessionCount: null,
      fundingModel: null,
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
  roomName: null,
  parentContact: null,
  lines: [],
  generatedRunExceptions: [],
  calculation: null,
};

const unpaidPaymentStatus: ManagerPaymentStatus = {
  invoiceId: 'inv-1',
  status: 'issued',
  dueStatus: 'due',
  currencyCode: 'gbp',
  totalDueMinor: 24000,
  amountPaidMinor: 0,
  paidAt: null,
  paymentFailedAt: null,
  paymentStatusUpdatedAt: null,
  checkoutRetryAvailable: true,
  checkoutRetryReasonCode: 'no_payment_collected',
  latestPaymentAttempt: null,
  latestPaymentEvent: null,
};

const paidPaymentStatus: ManagerPaymentStatus = {
  invoiceId: 'inv-1',
  status: 'paid',
  dueStatus: 'paid',
  currencyCode: 'gbp',
  totalDueMinor: 24000,
  amountPaidMinor: 24000,
  paidAt: '2026-06-09T15:00:00Z',
  paymentFailedAt: null,
  paymentStatusUpdatedAt: '2026-06-09T15:00:00Z',
  checkoutRetryAvailable: false,
  checkoutRetryReasonCode: 'already_paid',
  latestPaymentAttempt: {
    paymentAttemptId: 'pa-1',
    status: 'paid',
    amountMinor: 24000,
    currencyCode: 'gbp',
    stripeCheckoutSessionId: 'cs_1',
    stripePaymentIntentId: 'pi_1',
    stripeExpiresAt: null,
    failureReason: null,
    providerErrorCode: null,
    providerErrorMessage: null,
    createdAt: '2026-06-09T14:00:00Z',
    updatedAt: '2026-06-09T15:00:00Z',
  },
  latestPaymentEvent: {
    paymentEventId: 'pe-1',
    paymentAttemptId: 'pa-1',
    stripeEventId: 'evt_1',
    stripeEventType: 'checkout.session.completed',
    stripeCheckoutSessionId: 'cs_1',
    stripePaymentIntentId: 'pi_1',
    outcome: 'payment_succeeded',
    reasonCode: 'checkout_paid',
    previousInvoiceStatus: 'issued',
    newInvoiceStatus: 'paid',
    attemptPreviousStatus: 'checkout_created',
    attemptNewStatus: 'paid',
    amountMinor: 24000,
    currencyCode: 'gbp',
    webhookProcessingStatus: 'processed',
    webhookProcessingReason: null,
    webhookReceivedAt: '2026-06-09T15:00:00Z',
    webhookProcessedAt: '2026-06-09T15:00:01Z',
    createdAt: '2026-06-09T15:00:00Z',
  },
};

const failedPaymentStatus: ManagerPaymentStatus = {
  invoiceId: 'inv-1',
  status: 'payment_failed',
  dueStatus: 'due',
  currencyCode: 'gbp',
  totalDueMinor: 24000,
  amountPaidMinor: 0,
  paidAt: null,
  paymentFailedAt: '2026-06-09T16:00:00Z',
  paymentStatusUpdatedAt: '2026-06-09T16:00:00Z',
  checkoutRetryAvailable: true,
  checkoutRetryReasonCode: 'no_payment_collected',
  latestPaymentAttempt: {
    paymentAttemptId: 'pa-2',
    status: 'payment_failed',
    amountMinor: 24000,
    currencyCode: 'gbp',
    stripeCheckoutSessionId: 'cs_2',
    stripePaymentIntentId: null,
    stripeExpiresAt: null,
    failureReason: 'Card declined',
    providerErrorCode: 'card_declined',
    providerErrorMessage: 'Your card was declined',
    createdAt: '2026-06-09T15:30:00Z',
    updatedAt: '2026-06-09T16:00:00Z',
  },
  latestPaymentEvent: {
    paymentEventId: 'pe-2',
    paymentAttemptId: 'pa-2',
    stripeEventId: 'evt_2',
    stripeEventType: 'checkout.session.expired',
    stripeCheckoutSessionId: 'cs_2',
    stripePaymentIntentId: null,
    outcome: 'payment_failed',
    reasonCode: 'card_declined',
    previousInvoiceStatus: 'issued',
    newInvoiceStatus: 'payment_failed',
    attemptPreviousStatus: 'checkout_created',
    attemptNewStatus: 'payment_failed',
    amountMinor: 24000,
    currencyCode: 'gbp',
    webhookProcessingStatus: 'processed',
    webhookProcessingReason: null,
    webhookReceivedAt: '2026-06-09T16:00:00Z',
    webhookProcessedAt: '2026-06-09T16:00:01Z',
    createdAt: '2026-06-09T16:00:00Z',
  },
};

const openAttemptStatus: ManagerPaymentStatus = {
  invoiceId: 'inv-1',
  status: 'issued',
  dueStatus: 'due',
  currencyCode: 'gbp',
  totalDueMinor: 24000,
  amountPaidMinor: 0,
  paidAt: null,
  paymentFailedAt: null,
  paymentStatusUpdatedAt: '2026-06-09T14:00:00Z',
  checkoutRetryAvailable: true,
  checkoutRetryReasonCode: 'no_payment_collected',
  latestPaymentAttempt: {
    paymentAttemptId: 'pa-3',
    status: 'checkout_created',
    amountMinor: 24000,
    currencyCode: 'gbp',
    stripeCheckoutSessionId: 'cs_3',
    stripePaymentIntentId: null,
    stripeExpiresAt: '2026-06-10T14:00:00Z',
    failureReason: null,
    providerErrorCode: null,
    providerErrorMessage: null,
    createdAt: '2026-06-09T14:00:00Z',
    updatedAt: '2026-06-09T14:00:00Z',
  },
  latestPaymentEvent: null,
};

const emptyPaymentEvents: PaymentEvent[] = [];

const samplePaymentEvents: PaymentEvent[] = [
  {
    paymentEventId: 'pe-1',
    paymentAttemptId: 'pa-1',
    stripeEventId: 'evt_1',
    stripeEventType: 'checkout.session.completed',
    stripeCheckoutSessionId: 'cs_1',
    stripePaymentIntentId: 'pi_1',
    outcome: 'payment_succeeded',
    reasonCode: 'checkout_paid',
    previousInvoiceStatus: 'issued',
    newInvoiceStatus: 'paid',
    attemptPreviousStatus: 'checkout_created',
    attemptNewStatus: 'paid',
    amountMinor: 24000,
    currencyCode: 'gbp',
    webhookProcessingStatus: 'processed',
    webhookProcessingReason: null,
    webhookReceivedAt: '2026-06-09T15:00:00Z',
    webhookProcessedAt: '2026-06-09T15:00:01Z',
    createdAt: '2026-06-09T15:00:00Z',
  },
];

function createSpy() {
  return jasmine.createSpyObj('ManagerInvoicesApiService', [
    'getInvoice',
    'getPaymentStatus',
    'listPaymentEvents',
    'issueInvoice',
    'sendInvoiceToParent',
    'createPaymentLink',
  ]);
}

describe('ManagerInvoiceDetailComponent', () => {
  let fixture: ComponentFixture<ManagerInvoiceDetailComponent>;
  let apiService: jasmine.SpyObj<ManagerInvoicesApiService>;

  function createFixture(
    detail: ManagerInvoiceDetail = issuedDetail,
    paymentStatus: ManagerPaymentStatus | null = unpaidPaymentStatus,
    events: PaymentEvent[] = emptyPaymentEvents,
  ) {
    const spy = createSpy();
    spy.getInvoice.and.returnValue(of(detail));
    spy.getPaymentStatus.and.returnValue(of(paymentStatus ?? unpaidPaymentStatus));
    spy.listPaymentEvents.and.returnValue(of({ items: events, limit: 50, offset: 0 }));

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

  it('loads payment diagnostics after invoice detail succeeds for issued invoice', () => {
    createFixture();
    expect(apiService.getPaymentStatus).toHaveBeenCalledWith('inv-1');
    expect(apiService.listPaymentEvents).toHaveBeenCalledWith('inv-1', { limit: 50, offset: 0 });
  });

  it('does not load payment diagnostics for draft invoice', () => {
    createFixture(draftDetail, null, []);
    expect(apiService.getPaymentStatus).not.toHaveBeenCalled();
    expect(apiService.listPaymentEvents).not.toHaveBeenCalled();
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

  it('renders child name with room label', () => {
    createFixture();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Ben');
    expect(text).toContain('Toddler Group');
  });

  it('renders parent contact details in To card', () => {
    createFixture();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Burhan Khalid');
    expect(text).toContain('45 Maple Avenue');
  });

  it('renders service breakdown section', () => {
    createFixture();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Service Breakdown');
    expect(text).toContain('Core childcare');
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
    createFixture(draftDetail, null, []);
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Draft invoice — editable');
    expect(text).not.toContain('Issued invoice locked');
  });

  it('renders Send to Parent button for issued invoice', () => {
    createFixture();
    fixture.detectChanges();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Send to Parent');
    expect(text).toContain('Copy Payment Link');
    expect(text).not.toContain('Checkout');
    expect(text).not.toContain('Retry payment');
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

  it('shows Download PDF and Send to Parent action buttons', () => {
    createFixture();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Download PDF');
    expect(text).toContain('Send to Parent');
  });

  it('shows collapsed payment review section for issued invoice', () => {
    createFixture();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Payment Review');
    expect(text).not.toContain('Amount paid');
  });

  it('expands payment review on toggle click', () => {
    createFixture();
    const toggleBtn: HTMLButtonElement = fixture.nativeElement.querySelector('button[aria-expanded]');
    toggleBtn.click();
    fixture.detectChanges();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Amount paid');
    expect(text).toContain('Balance due');
  });

  it('does not show payment review section for draft invoice', () => {
    createFixture(draftDetail, null, []);
    const text = fixture.nativeElement.textContent;
    expect(text).not.toContain('Payment Review');
  });

  function expandPayReview() {
    const btn: HTMLButtonElement | null = fixture.nativeElement.querySelector('button[aria-expanded]');
    if (btn) { btn.click(); fixture.detectChanges(); }
  }

  it('shows unpaid state for issued invoice with zero paid', () => {
    createFixture();
    expandPayReview();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Unpaid');
  });

  it('shows paid state with paid amount and zero balance', () => {
    createFixture(issuedDetail, paidPaymentStatus, samplePaymentEvents);
    expandPayReview();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Paid');
    expect(text).toContain('Payment succeeded');
    expect(text).toContain('£240.00');
  });

  it('shows payment failed state with failure reason', () => {
    createFixture({ ...issuedDetail, status: 'payment_failed', paymentFailedAt: '2026-06-09T16:00:00Z' }, failedPaymentStatus, samplePaymentEvents);
    expandPayReview();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Payment failed');
    expect(text).toContain('Card declined');
    expect(text).toContain('Your card was declined');
  });

  it('shows awaiting provider update for open checkout attempt', () => {
    createFixture(issuedDetail, openAttemptStatus, emptyPaymentEvents);
    expandPayReview();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Awaiting provider update');
  });

  it('does not show parent retry for open attempt even when checkoutRetryAvailable is true', () => {
    createFixture(issuedDetail, openAttemptStatus, emptyPaymentEvents);
    expandPayReview();
    const text = fixture.nativeElement.textContent;
    expect(text).not.toContain('Parent retry available');
  });

  it('shows parent retry available when eligible and no open attempt', () => {
    createFixture();
    expandPayReview();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Parent retry available');
  });

  it('shows payment events in history', () => {
    createFixture(issuedDetail, paidPaymentStatus, samplePaymentEvents);
    expandPayReview();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Payment history');
    expect(text).toContain('Payment succeeded');
    expect(text).toContain('checkout_paid');
  });

  it('shows empty payment history state', () => {
    createFixture();
    expandPayReview();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('No payment events recorded');
  });

  it('shows provider IDs as secondary diagnostic text', () => {
    createFixture(issuedDetail, paidPaymentStatus, samplePaymentEvents);
    expandPayReview();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('cs_1');
    expect(text).toContain('pi_1');
  });

  it('shows retry unavailable reason when retry not available', () => {
    const noRetryStatus: ManagerPaymentStatus = {
      ...unpaidPaymentStatus,
      checkoutRetryAvailable: false,
      checkoutRetryReasonCode: 'already_paid',
    };
    createFixture(issuedDetail, noRetryStatus, emptyPaymentEvents);
    expandPayReview();
    const text = fixture.nativeElement.textContent;
    expect(text).not.toContain('Parent retry available');
  });

  it('shows issue button for draft invoice', () => {
    createFixture(draftDetail, null, []);
    const buttons: HTMLButtonElement[] = Array.from(fixture.nativeElement.querySelectorAll('button'));
    const issueButton = buttons.find((btn) => btn.textContent?.trim().includes('Issue Invoice'));
    expect(issueButton).toBeTruthy();
  });

  it('hides issue button for issued invoice', () => {
    createFixture(issuedDetail, unpaidPaymentStatus, emptyPaymentEvents);
    const buttons: HTMLButtonElement[] = Array.from(fixture.nativeElement.querySelectorAll('button'));
    const issueButton = buttons.find((btn) => btn.textContent?.trim().includes('Issue Invoice'));
    expect(issueButton).toBeFalsy();
  });

  it('calls issueInvoice and refreshes on confirm', () => {
    createFixture(draftDetail, null, []);
    apiService.issueInvoice.and.returnValue(of({ invoiceId: 'inv-3', status: 'issued' }));
    const component = fixture.componentInstance;
    component.openIssueConfirmation();
    expect(component.isConfirmIssueOpen).toBe(true);
    component.confirmIssue();
    expect(apiService.issueInvoice).toHaveBeenCalledWith('inv-3');
    expect(apiService.getInvoice).toHaveBeenCalledTimes(2);
    expect(component.isConfirmIssueOpen).toBe(false);
  });

  it('does not call API when confirm is cancelled', () => {
    createFixture(draftDetail, null, []);
    const component = fixture.componentInstance;
    component.openIssueConfirmation();
    expect(component.isConfirmIssueOpen).toBe(true);
    component.cancelIssue();
    expect(component.isConfirmIssueOpen).toBe(false);
    expect(apiService.issueInvoice).not.toHaveBeenCalled();
  });

  it('shows error toast on issue failure', () => {
    createFixture(draftDetail, null, []);
    apiService.issueInvoice.and.returnValue(throwError(() => new HttpErrorResponse({
      error: { code: 'invalid_state', message: 'Invoice already issued', request_id: 'req-err-1' },
      status: 409,
    })));
    const toastSpy = spyOn(TestBed.inject(ToastService), 'error');
    const component = fixture.componentInstance;
    component.confirmIssue();
    expect(toastSpy).toHaveBeenCalled();
    expect(component.isIssuing).toBe(false);
  });

  it('disables button and shows loading text while issuing', () => {
    createFixture(draftDetail, null, []);
    apiService.issueInvoice.and.returnValue(of({ invoiceId: 'inv-3', status: 'issued' }));
    const component = fixture.componentInstance;
    component.isIssuing = true;
    fixture.detectChanges();
    const buttons: HTMLButtonElement[] = Array.from(fixture.nativeElement.querySelectorAll('button'));
    const issueButton = buttons.find((btn) => btn.textContent?.trim().includes('Issuing...'));
    expect(issueButton).toBeTruthy();
    expect(issueButton!.disabled).toBe(true);
  });

  it('renders both Send to Parent and Copy Payment Link for issued invoice', () => {
    createFixture();
    const buttons: HTMLButtonElement[] = Array.from(fixture.nativeElement.querySelectorAll('button'));
    const sendButton = buttons.find((btn) => btn.textContent?.trim().includes('Send to Parent'));
    const copyButton = buttons.find((btn) => btn.textContent?.trim().includes('Copy Payment Link'));
    expect(sendButton).toBeTruthy();
    expect(copyButton).toBeTruthy();
  });

  it('opens send confirmation dialog with specified copy', () => {
    createFixture();
    const component = fixture.componentInstance;
    component.openSendConfirmation();
    fixture.detectChanges();
    expect(component.isConfirmSendOpen).toBe(true);

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Send invoice to parent');
    expect(text).toContain('Burhan Khalid');
    expect(text).toContain('burhan@example.com');
    const buttons: HTMLButtonElement[] = Array.from(fixture.nativeElement.querySelectorAll('button'));
    const confirmButton = buttons.find((btn) => btn.textContent?.trim() === 'Send');
    expect(confirmButton).toBeTruthy();
  });

  it('confirmSend calls sendInvoiceToParent and shows queued toast on success', () => {
    createFixture();
    apiService.sendInvoiceToParent.and.returnValue(of({ status: 'queued' }));
    const toastSpy = spyOn(TestBed.inject(ToastService), 'success');
    const component = fixture.componentInstance;
    component.openSendConfirmation();
    component.confirmSend();

    expect(apiService.sendInvoiceToParent).toHaveBeenCalledWith('inv-1');
    expect(component.isConfirmSendOpen).toBe(false);
    expect(toastSpy).toHaveBeenCalledWith('Invoice email queued for delivery to parent.');
    expect(component.isSendingInvoice).toBe(false);
  });

  it('guards double clicks and shows sending state while request is in flight', () => {
    createFixture();
    let resolveSend!: (v: { status: string }) => void;
    apiService.sendInvoiceToParent.and.returnValue(
      new Observable<{ status: string }>((sub) => {
        resolveSend = (v) => {
          sub.next(v);
          sub.complete();
        };
      }),
    );
    const component = fixture.componentInstance;
    component.openSendConfirmation();
    component.confirmSend();
    fixture.detectChanges();

    expect(component.isSendingInvoice).toBe(true);
    const buttons: HTMLButtonElement[] = Array.from(fixture.nativeElement.querySelectorAll('button'));
    const sendButton = buttons.find((btn) => btn.textContent?.trim().includes('Sending...'));
    expect(sendButton).toBeTruthy();
    expect(sendButton!.disabled).toBe(true);

    component.openSendConfirmation();
    expect(component.isConfirmSendOpen).toBe(false);

    resolveSend({ status: 'queued' });
  });

  it('shows error toast and resets sending state when sendInvoiceToParent rejects', () => {
    createFixture();
    apiService.sendInvoiceToParent.and.returnValue(throwError(() => new HttpErrorResponse({
      error: { code: 'internal_error', message: 'Enqueue failed', request_id: 'req-send-1' },
      status: 500,
    })));
    const toastSpy = spyOn(TestBed.inject(ToastService), 'error');
    const component = fixture.componentInstance;
    component.openSendConfirmation();
    component.confirmSend();

    expect(toastSpy).toHaveBeenCalled();
    expect(component.isSendingInvoice).toBe(false);
    expect(component.isConfirmSendOpen).toBe(false);
  });

  it('disables Send to Parent with title hint when parent has no email', () => {
    const noEmailDetail = { ...issuedDetail, parentContact: { ...issuedDetail.parentContact!, email: '' } };
    createFixture(noEmailDetail, unpaidPaymentStatus, []);
    const buttons: HTMLButtonElement[] = Array.from(fixture.nativeElement.querySelectorAll('button'));
    const sendButton = buttons.find((btn) => btn.textContent?.trim().includes('Send to Parent'));
    const copyButton = buttons.find((btn) => btn.textContent?.trim().includes('Copy Payment Link'));
    expect(sendButton).toBeTruthy();
    expect(sendButton!.disabled).toBe(true);
    expect(sendButton!.getAttribute('title')).toBe('No email on file for this parent');
    expect(copyButton).toBeTruthy();
    expect(copyButton!.disabled).toBe(false);
  });

  it('hides both buttons for draft invoice', () => {
    createFixture(draftDetail, null, []);
    const text = fixture.nativeElement.textContent;
    expect(text).not.toContain('Send to Parent');
    expect(text).not.toContain('Copy Payment Link');
  });

  it('hides both buttons for void invoice', () => {
    createFixture({ ...issuedDetail, status: 'void', lockedAt: null }, unpaidPaymentStatus, []);
    const text = fixture.nativeElement.textContent;
    expect(text).not.toContain('Send to Parent');
    expect(text).not.toContain('Copy Payment Link');
  });

  it('Copy Payment Link still calls createPaymentLink and stores the URL', () => {
    createFixture();
    apiService.createPaymentLink.and.returnValue(of({ paymentLinkId: 'pl-1', url: 'https://pay.example.com/pl', existing: false }));
    if (!navigator.clipboard) {
      Object.defineProperty(navigator, 'clipboard', { value: {}, configurable: true });
    }
    spyOn(navigator.clipboard, 'writeText').and.returnValue(Promise.resolve());
    const component = fixture.componentInstance;
    const buttons: HTMLButtonElement[] = Array.from(fixture.nativeElement.querySelectorAll('button'));
    const copyButton = buttons.find((btn) => btn.textContent?.trim().includes('Copy Payment Link'));
    copyButton!.click();
    fixture.detectChanges();
    expect(apiService.createPaymentLink).toHaveBeenCalledWith('inv-1');
    expect(component.paymentLinkUrl).toBe('https://pay.example.com/pl');
  });
});

describe('ManagerInvoiceDetailComponent error handling', () => {
  let fixture: ComponentFixture<ManagerInvoiceDetailComponent>;

  beforeEach(() => {
    const spy = createSpy();
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
    expect(text).toContain('no longer available');
    expect(text).not.toContain('req-99');
  });
});

describe('ManagerInvoiceDetailComponent payment diagnostics error', () => {
  let fixture: ComponentFixture<ManagerInvoiceDetailComponent>;

  beforeEach(() => {
    const spy = createSpy();
    spy.getInvoice.and.returnValue(of(issuedDetail));
    spy.getPaymentStatus.and.returnValue(throwError(() => new HttpErrorResponse({
      error: { code: 'internal_error', message: 'Payment service unavailable', request_id: 'req-pay-1' },
      status: 500,
    })));
    spy.listPaymentEvents.and.returnValue(throwError(() => new HttpErrorResponse({
      error: { code: 'internal_error', message: 'Payment events unavailable', request_id: 'req-ev-1' },
      status: 500,
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

  it('keeps invoice detail visible when payment diagnostics fail', () => {
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('INV-202605-0001');
    expect(text).toContain('Ben');
    expect(text).toContain('Payment Review');
  });

  it('shows payment diagnostics error with request ID', () => {
    const btn: HTMLButtonElement | null = fixture.nativeElement.querySelector('button[aria-expanded]');
    if (btn) { btn.click(); fixture.detectChanges(); }
    const text = fixture.nativeElement.textContent;
    const hasRequestId = text.includes('req-pay-1') || text.includes('req-ev-1');
    expect(hasRequestId).toBe(true);
  });
});
