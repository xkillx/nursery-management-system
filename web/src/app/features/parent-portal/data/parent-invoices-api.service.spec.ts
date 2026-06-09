import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { ParentInvoicesApiService } from './parent-invoices-api.service';

describe('ParentInvoicesApiService', () => {
  let service: ParentInvoicesApiService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        ParentInvoicesApiService,
        provideHttpClient(),
        provideHttpClientTesting(),
      ],
    });

    service = TestBed.inject(ParentInvoicesApiService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  describe('listInvoices', () => {
    it('sends limit and offset params', () => {
      service.listInvoices({ limit: 200, offset: 0 }).subscribe();

      const req = httpMock.expectOne((r) => r.url.includes('/parent/invoices'));
      expect(req.request.params.get('limit')).toBe('200');
      expect(req.request.params.get('offset')).toBe('0');
      req.flush({ items: [], limit: 200, offset: 0 });
    });

    it('maps list fields from snake_case to camelCase', (done) => {
      const apiResponse = {
        items: [{
          invoice_id: 'inv-1',
          invoice_kind: 'monthly',
          invoice_number: 'INV-001',
          invoice_number_display: 'INV-001',
          child_id: 'c-1',
          child_name: 'Ada',
          billing_month: '2026-05',
          period: { start_date: '2026-05-01', end_date: '2026-05-31' },
          status: 'issued',
          due_status: 'due',
          currency_code: 'gbp',
          subtotal_minor: 45000,
          funded_deduction_minor: 5000,
          total_due_minor: 40000,
          amount_paid_minor: 0,
          due_at: '2026-06-01T00:00:00Z',
          issued_at: '2026-05-28T00:00:00Z',
          paid_at: null,
          payment_failed_at: null,
          payment_status_updated_at: null,
        }],
        limit: 200,
        offset: 0,
      };

      service.listInvoices({ limit: 200, offset: 0 }).subscribe((result) => {
        expect(result.items.length).toBe(1);
        const item = result.items[0];
        expect(item.invoiceId).toBe('inv-1');
        expect(item.invoiceKind).toBe('monthly');
        expect(item.invoiceNumberDisplay).toBe('INV-001');
        expect(item.childName).toBe('Ada');
        expect(item.billingMonth).toBe('2026-05');
        expect(item.period?.startDate).toBe('2026-05-01');
        expect(item.status).toBe('issued');
        expect(item.dueStatus).toBe('due');
        expect(item.subtotalMinor).toBe(45000);
        expect(item.fundedDeductionMinor).toBe(5000);
        expect(item.totalDueMinor).toBe(40000);
        expect(item.amountPaidMinor).toBe(0);
        expect(result.limit).toBe(200);
        expect(result.offset).toBe(0);
        done();
      });

      const req = httpMock.expectOne((r) => r.url.includes('/parent/invoices'));
      req.flush(apiResponse);
    });
  });

  describe('getInvoice', () => {
    it('maps detail fields including calculation and lines', (done) => {
      const apiResponse = {
        invoice_id: 'inv-1',
        invoice_kind: 'monthly',
        invoice_number: 'INV-001',
        invoice_number_display: 'INV-001',
        child_id: 'c-1',
        child_name: 'Ada',
        billing_month: '2026-05',
        period: { start_date: '2026-05-01', end_date: '2026-05-31' },
        status: 'issued',
        due_status: 'due',
        currency_code: 'gbp',
        subtotal_minor: 45000,
        funded_deduction_minor: 5000,
        total_due_minor: 40000,
        amount_paid_minor: 0,
        issued_at: '2026-05-28T00:00:00Z',
        due_at: '2026-06-01T00:00:00Z',
        paid_at: null,
        payment_failed_at: null,
        payment_status_updated_at: null,
        calculation: {
          raw_attended_minutes: 600,
          rounded_attended_minutes: 600,
          funded_allowance_minutes: 300,
          funded_deduction_minutes: 150,
          core_billable_minutes: 450,
          included_session_count: 20,
          core_hourly_rate_minor: 6000,
          core_subtotal_minor: 45000,
          extras_total_minor: 5000,
        },
        lines: [
          { line_kind: 'core', description: 'Core sessions', sort_order: 2, quantity_minutes: 450, unit_amount_minor: 100, line_amount_minor: 45000 },
          { line_kind: 'extras', description: 'Extra sessions', sort_order: 1, quantity_minutes: 50, unit_amount_minor: 100, line_amount_minor: 5000 },
        ],
      };

      service.getInvoice('inv-1').subscribe((detail) => {
        expect(detail.invoiceId).toBe('inv-1');
        expect(detail.childName).toBe('Ada');
        expect(detail.calculation?.roundedAttendedMinutes).toBe(600);
        expect(detail.calculation?.coreHourlyRateMinor).toBe(6000);
        expect(detail.lines.length).toBe(2);
        expect(detail.lines[0].sortOrder).toBe(1);
        expect(detail.lines[0].lineKind).toBe('extras');
        expect(detail.lines[1].sortOrder).toBe(2);
        done();
      });

      const req = httpMock.expectOne((r) => r.url.includes('/parent/invoices/inv-1'));
      req.flush(apiResponse);
    });
  });

  describe('createCheckoutSession', () => {
    it('posts and maps checkout_url', (done) => {
      const apiResponse = {
        checkout_session_id: 'cs-123',
        checkout_url: 'https://checkout.stripe.com/session',
        payment_attempt_id: 'pa-456',
      };

      service.createCheckoutSession('inv-1').subscribe((result) => {
        expect(result.checkoutSessionId).toBe('cs-123');
        expect(result.checkoutUrl).toBe('https://checkout.stripe.com/session');
        expect(result.paymentAttemptId).toBe('pa-456');
        done();
      });

      const req = httpMock.expectOne((r) => r.url.includes('/parent/invoices/inv-1/checkout-sessions'));
      expect(req.request.method).toBe('POST');
      req.flush(apiResponse);
    });
  });
});
