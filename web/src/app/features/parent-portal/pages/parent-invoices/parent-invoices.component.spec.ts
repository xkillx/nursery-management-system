import { of, Subject, throwError } from 'rxjs';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { provideRouter } from '@angular/router';

import { ParentInvoicesComponent } from './parent-invoices.component';
import { ParentInvoicesApiService } from '../../data/parent-invoices-api.service';
import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { ParentInvoiceListItem, CheckoutSessionResult } from '../../models/parent-invoices.models';

function makeItem(overrides: Partial<Omit<ParentInvoiceListItem, 'invoiceId' | 'status'>> & { invoiceId: string; status?: ParentInvoiceListItem['status'] }): ParentInvoiceListItem {
  return {
    invoiceKind: 'monthly',
    invoiceNumber: null,
    invoiceNumberDisplay: `INV-${overrides.invoiceId}`,
    childId: 'child-1',
    childName: 'Ada Lovelace',
    billingMonth: '2026-05',
    period: null,
    status: 'issued',
    dueStatus: 'due',
    currencyCode: 'gbp',
    subtotalMinor: 45000,
    fundedDeductionMinor: 0,
    totalDueMinor: 45000,
    amountPaidMinor: 0,
    dueAt: '2026-06-01T00:00:00Z',
    issuedAt: '2026-05-28T00:00:00Z',
    paidAt: null,
    paymentFailedAt: null,
    paymentStatusUpdatedAt: null,
    ...overrides,
  };
}

describe('ParentInvoicesComponent', () => {
  let fixture: ComponentFixture<ParentInvoicesComponent>;
  let component: ParentInvoicesComponent;
  let apiMock: jasmine.SpyObj<ParentInvoicesApiService>;
  let errorMapperMock: jasmine.SpyObj<ApiErrorMapper>;

  beforeEach(async () => {
    apiMock = jasmine.createSpyObj('ParentInvoicesApiService', ['listInvoices', 'createCheckoutSession']);
    errorMapperMock = jasmine.createSpyObj('ApiErrorMapper', ['mapAndHandle']);

    await TestBed.configureTestingModule({
      imports: [ParentInvoicesComponent],
      providers: [
        provideRouter([]),
        { provide: ParentInvoicesApiService, useValue: apiMock },
        { provide: ApiErrorMapper, useValue: errorMapperMock },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ParentInvoicesComponent);
    component = fixture.componentInstance;
  });

  it('shows loading state on init', () => {
    apiMock.listInvoices.and.returnValue(of({ items: [], total: 0, page: 1, pageSize: 50 }));
    fixture.detectChanges();
    expect(apiMock.listInvoices).toHaveBeenCalledWith({ page: 1, pageSize: 50 });
  });

  it('shows empty state when no invoices', () => {
    apiMock.listInvoices.and.returnValue(of({ items: [], total: 0, page: 1, pageSize: 50 }));
    fixture.detectChanges();
    const emptyEl = fixture.debugElement.query(By.css('app-empty-state'));
    expect(emptyEl).toBeTruthy();
  });

  it('shows error state', () => {
    apiMock.listInvoices.and.returnValue(throwError(() => ({ error: { code: 'internal_error', message: 'fail' } })));
    errorMapperMock.mapAndHandle.and.returnValue({ code: 'internal_error', message: 'fail', requestId: null, fieldErrors: {} });
    fixture.detectChanges();
    const alertEl = fixture.debugElement.query(By.css('app-alert'));
    expect(alertEl).toBeTruthy();
  });

  it('shows attention section for overdue invoices', () => {
    const items = [
      makeItem({ invoiceId: '1', status: 'overdue', dueStatus: 'overdue' }),
      makeItem({ invoiceId: '2', status: 'paid', dueStatus: 'paid', totalDueMinor: 45000, amountPaidMinor: 45000 }),
    ];
    apiMock.listInvoices.and.returnValue(of({ items, total: items.length, page: 1, pageSize: 50 }));
    fixture.detectChanges();

    const heading = fixture.debugElement.query(By.css('[aria-labelledby="attention-heading"]'));
    expect(heading).toBeTruthy();
  });

  it('attention invoices excluded from history groups', () => {
    const items = [
      makeItem({ invoiceId: '1', status: 'overdue', dueStatus: 'overdue' }),
      makeItem({ invoiceId: '2', status: 'paid', dueStatus: 'paid', totalDueMinor: 45000, amountPaidMinor: 45000 }),
    ];
    apiMock.listInvoices.and.returnValue(of({ items, total: items.length, page: 1, pageSize: 50 }));
    fixture.detectChanges();

    const historySections = fixture.debugElement.queryAll(By.css('h2'));
    const childGroupHeadings = historySections.filter((el) => !el.nativeElement.textContent.includes('Needs attention'));
    expect(childGroupHeadings.length).toBe(1);
  });

  it('shows child-grouped history', () => {
    const items = [
      makeItem({ invoiceId: '1', childId: 'c1', childName: 'Ada', billingMonth: '2026-05', status: 'paid', dueStatus: 'paid', totalDueMinor: 45000, amountPaidMinor: 45000 }),
      makeItem({ invoiceId: '2', childId: 'c2', childName: 'Zara', billingMonth: '2026-05', status: 'paid', dueStatus: 'paid', totalDueMinor: 45000, amountPaidMinor: 45000 }),
    ];
    apiMock.listInvoices.and.returnValue(of({ items, total: items.length, page: 1, pageSize: 50 }));
    fixture.detectChanges();

    expect(component.historyGroups.length).toBe(2);
  });

  it('shows view details link', () => {
    const items = [makeItem({ invoiceId: '1', status: 'paid', dueStatus: 'paid', totalDueMinor: 45000, amountPaidMinor: 45000 })];
    apiMock.listInvoices.and.returnValue(of({ items, total: items.length, page: 1, pageSize: 50 }));
    fixture.detectChanges();

    const links = fixture.debugElement.queryAll(By.css('a'));
    const viewLinks = links.filter((el) => el.properties['href']?.includes('/parent/invoices/'));
    expect(viewLinks.length).toBeGreaterThan(0);
  });

  it('startPayment calls checkout session and redirects', () => {
    const items = [makeItem({ invoiceId: '1', status: 'overdue', dueStatus: 'overdue' })];
    apiMock.listInvoices.and.returnValue(of({ items, total: items.length, page: 1, pageSize: 50 }));
    apiMock.createCheckoutSession.and.returnValue(of({
      checkoutSessionId: 'cs-1',
      checkoutUrl: 'https://checkout.stripe.com/session',
      paymentAttemptId: 'pa-1',
    }));

    fixture.detectChanges();
    spyOn(component, 'redirectTo');
    component.startPayment('1');
    expect(apiMock.createCheckoutSession).toHaveBeenCalledWith('1');
    expect(component.redirectTo).toHaveBeenCalledWith('https://checkout.stripe.com/session');
  });

  it('handles payment error', () => {
    const items = [makeItem({ invoiceId: '1', status: 'overdue', dueStatus: 'overdue' })];
    apiMock.listInvoices.and.returnValue(of({ items, total: items.length, page: 1, pageSize: 50 }));
    apiMock.createCheckoutSession.and.returnValue(throwError(() => ({ error: { code: 'conflict', message: 'already paid' } })));
    errorMapperMock.mapAndHandle.and.returnValue({ code: 'conflict', message: 'already paid', requestId: null, fieldErrors: {} });

    fixture.detectChanges();
    component.startPayment('1');
    expect(component.payingInvoiceIds.has('1')).toBeFalse();
  });

  it('shows load more when more pages available', () => {
    const items = Array.from({ length: 50 }, (_, i) => makeItem({ invoiceId: String(i), status: 'paid', dueStatus: 'paid', totalDueMinor: 45000, amountPaidMinor: 45000 }));
    apiMock.listInvoices.and.returnValue(of({ items, total: 100, page: 1, pageSize: 50 }));
    fixture.detectChanges();

    expect(component.hasMore).toBeTrue();
  });

  it('duplicate startPayment does not create second checkout session', () => {
    const items = [makeItem({ invoiceId: '1', status: 'overdue', dueStatus: 'overdue' })];
    apiMock.listInvoices.and.returnValue(of({ items, total: items.length, page: 1, pageSize: 50 }));
    apiMock.createCheckoutSession.and.returnValue(new Subject<CheckoutSessionResult>().asObservable());

    fixture.detectChanges();
    component.startPayment('1');
    component.startPayment('1'); // second call while first is pending

    expect(apiMock.createCheckoutSession).toHaveBeenCalledTimes(1);
    expect(component.payingInvoiceIds.has('1')).toBeTrue();
  });

  it('payment error removes only that invoice ID from payingInvoiceIds', () => {
    const items = [
      makeItem({ invoiceId: '1', status: 'overdue', dueStatus: 'overdue' }),
      makeItem({ invoiceId: '2', status: 'overdue', dueStatus: 'overdue' }),
    ];
    apiMock.listInvoices.and.returnValue(of({ items, total: items.length, page: 1, pageSize: 50 }));
    apiMock.createCheckoutSession.withArgs('1').and.returnValue(new Subject<CheckoutSessionResult>().asObservable());
    apiMock.createCheckoutSession.withArgs('2').and.returnValue(throwError(() => ({ error: { code: 'conflict', message: 'fail' } })));
    errorMapperMock.mapAndHandle.and.returnValue({ code: 'conflict', message: 'fail', requestId: null, fieldErrors: {} });

    fixture.detectChanges();
    component.startPayment('1');
    component.startPayment('2');

    expect(component.payingInvoiceIds.has('1')).toBeTrue(); // still pending
    expect(component.payingInvoiceIds.has('2')).toBeFalse(); // removed on error
  });
});
