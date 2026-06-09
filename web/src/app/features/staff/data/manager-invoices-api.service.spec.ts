import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { ManagerInvoicesApiService } from './manager-invoices-api.service';

describe('ManagerInvoicesApiService', () => {
  let service: ManagerInvoicesApiService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [ManagerInvoicesApiService],
    });
    service = TestBed.inject(ManagerInvoicesApiService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  describe('listInvoices', () => {
    it('omits status query param when filter is all', () => {
      const response = {
        items: [
          {
            invoice_id: 'inv-1',
            invoice_kind: 'monthly',
            invoice_number: null,
            invoice_number_display: 'Draft invoice',
            child_id: 'c1',
            child_name: 'Ben',
            billing_month: '2026-05',
            status: 'draft',
            due_status: 'not_due',
            currency_code: 'gbp',
            subtotal_minor: 33000,
            funded_deduction_minor: 9000,
            total_due_minor: 24000,
            amount_paid_minor: 0,
            due_at: null,
            issued_at: null,
            paid_at: null,
            payment_failed_at: null,
            payment_status_updated_at: null,
            generated_run_id: null,
            generated_run_status: null,
            generated_run_started_at: null,
            generated_run_completed_at: null,
            generated_run_exception_count: null,
            created_at: '2026-06-09T10:00:00Z',
            updated_at: '2026-06-09T10:00:00Z',
          },
        ],
        limit: 50,
        offset: 0,
      };

      service.listInvoices({ billingMonth: '2026-05', status: 'all', limit: 50, offset: 0 }).subscribe((result) => {
        expect(result.items.length).toBe(1);
        expect(result.items[0].invoiceId).toBe('inv-1');
        expect(result.items[0].invoiceNumberDisplay).toBe('Draft invoice');
        expect(result.items[0].dueStatus).toBe('not_due');
        expect(result.limit).toBe(50);
        expect(result.offset).toBe(0);
      });

      const req = httpMock.expectOne((r) => r.url === '/api/v1/invoices');
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('billing_month')).toBe('2026-05');
      expect(req.request.params.get('status')).toBeNull();
      expect(req.request.params.get('limit')).toBe('50');
      expect(req.request.params.get('offset')).toBe('0');
      req.flush(response);
    });

    it('includes status query param for concrete filter', () => {
      service.listInvoices({ billingMonth: '2026-05', status: 'issued', limit: 50, offset: 0 }).subscribe();

      const req = httpMock.expectOne((r) => r.url === '/api/v1/invoices');
      expect(req.request.params.get('status')).toBe('issued');
      req.flush({ items: [], limit: 50, offset: 0 });
    });

    it('sends limit and offset for pagination', () => {
      service.listInvoices({ billingMonth: '2026-05', status: 'all', limit: 50, offset: 50 }).subscribe();

      const req = httpMock.expectOne((r) => r.url === '/api/v1/invoices');
      expect(req.request.params.get('limit')).toBe('50');
      expect(req.request.params.get('offset')).toBe('50');
      req.flush({ items: [], limit: 50, offset: 50 });
    });

    it('maps camelCase fields from snake_case response', () => {
      const response = {
        items: [
          {
            invoice_id: 'inv-2',
            invoice_kind: 'monthly',
            invoice_number: 'INV-202605-0001',
            invoice_number_display: 'INV-202605-0001',
            child_id: 'c2',
            child_name: 'Alice',
            billing_month: '2026-05',
            period: { start_date: '2026-05-01', end_date: '2026-05-31' },
            status: 'issued',
            due_status: 'due',
            currency_code: 'gbp',
            subtotal_minor: 33000,
            funded_deduction_minor: 9000,
            total_due_minor: 24000,
            amount_paid_minor: 0,
            due_at: '2026-06-10T00:00:00Z',
            issued_at: '2026-06-09T12:00:00Z',
            paid_at: null,
            payment_failed_at: null,
            payment_status_updated_at: null,
            generated_run_id: 'run-1',
            generated_run_status: 'completed',
            generated_run_started_at: '2026-06-09T11:00:00Z',
            generated_run_completed_at: '2026-06-09T11:05:00Z',
            generated_run_exception_count: 0,
            created_at: '2026-06-09T10:00:00Z',
            updated_at: '2026-06-09T12:00:00Z',
          },
        ],
        limit: 50,
        offset: 0,
      };

      service.listInvoices({ billingMonth: '2026-05', status: 'all', limit: 50, offset: 0 }).subscribe((result) => {
        const item = result.items[0];
        expect(item.childName).toBe('Alice');
        expect(item.invoiceNumber).toBe('INV-202605-0001');
        expect(item.period?.startDate).toBe('2026-05-01');
        expect(item.period?.endDate).toBe('2026-05-31');
        expect(item.generatedRunId).toBe('run-1');
        expect(item.amountPaidMinor).toBe(0);
      });

      const req = httpMock.expectOne((r) => r.url === '/api/v1/invoices');
      req.flush(response);
    });

    it('uses null-safe defaults for optional fields', () => {
      const minimalItem = {
        invoice_id: 'inv-3',
        invoice_kind: 'monthly',
        invoice_number: null,
        child_id: 'c3',
        child_name: 'Chloe',
        billing_month: '2026-05',
        status: 'draft',
        subtotal_minor: 10000,
        funded_deduction_minor: 0,
        total_due_minor: 10000,
        issued_at: null,
        created_at: '2026-06-09T10:00:00Z',
        updated_at: '2026-06-09T10:00:00Z',
      };

      service.listInvoices({ billingMonth: '2026-05', status: 'all', limit: 50, offset: 0 }).subscribe((result) => {
        const item = result.items[0];
        expect(item.dueStatus).toBe('not_due');
        expect(item.amountPaidMinor).toBe(0);
        expect(item.invoiceNumberDisplay).toBe('');
        expect(item.period).toBeNull();
      });

      const req = httpMock.expectOne((r) => r.url === '/api/v1/invoices');
      req.flush({ items: [minimalItem], limit: 50, offset: 0 });
    });
  });

  describe('getInvoice', () => {
    it('maps detail with due status, locked timestamp, calculation, and line sorting', () => {
      const detailResponse = {
        invoice_id: 'inv-1',
        invoice_kind: 'monthly',
        invoice_number: 'INV-202605-0001',
        invoice_number_display: 'INV-202605-0001',
        child_id: 'c1',
        child_name: 'Ben',
        billing_month: '2026-05',
        status: 'issued',
        due_status: 'due',
        currency_code: 'gbp',
        subtotal_minor: 33000,
        funded_deduction_minor: 9000,
        total_due_minor: 24000,
        amount_paid_minor: 0,
        issued_at: '2026-06-09T12:00:00Z',
        locked_at: '2026-06-09T12:00:00Z',
        due_at: '2026-06-10T00:00:00Z',
        paid_at: null,
        payment_failed_at: null,
        payment_status_updated_at: null,
        adjusts_invoice_id: null,
        adjustment_reason_code: null,
        adjustment_reason_note: null,
        generated_run_id: 'run-1',
        generated_run_status: 'completed',
        generated_run_exception_count: 1,
        generated_run_exceptions: [
          { child_id: 'c2', child_name: 'Alice', blocker_codes: ['incomplete_attendance'] },
        ],
        calculation: {
          core_hourly_rate_minor: 25,
          raw_attended_minutes: 1315,
          rounded_attended_minutes: 1320,
          funded_allowance_minutes: 360,
          funded_deduction_minutes: 360,
          core_billable_minutes: 960,
          included_session_count: 4,
          core_subtotal_minor: 24000,
          extras_total_minor: 9000,
        },
        lines: [
          { line_id: 'l2', line_kind: 'funded_deduction', description: 'Funded deduction', sort_order: 2, quantity_minutes: 360, unit_amount_minor: 25, line_amount_minor: -9000 },
          { line_id: 'l1', line_kind: 'core_childcare', description: 'Core childcare', sort_order: 1, quantity_minutes: 1320, unit_amount_minor: 25, line_amount_minor: 33000 },
        ],
        created_at: '2026-06-09T10:00:00Z',
        updated_at: '2026-06-09T12:00:00Z',
      };

      service.getInvoice('inv-1').subscribe((detail) => {
        expect(detail.invoiceId).toBe('inv-1');
        expect(detail.dueStatus).toBe('due');
        expect(detail.lockedAt).toBe('2026-06-09T12:00:00Z');
        expect(detail.calculation?.coreHourlyRateMinor).toBe(25);
        expect(detail.calculation?.roundedAttendedMinutes).toBe(1320);
        expect(detail.calculation?.coreBillableMinutes).toBe(960);
        expect(detail.generatedRunExceptions.length).toBe(1);
        expect(detail.generatedRunExceptions[0].childName).toBe('Alice');
        expect(detail.generatedRunExceptions[0].blockerCodes).toEqual(['incomplete_attendance']);
        // Lines sorted by sort_order
        expect(detail.lines[0].lineId).toBe('l1');
        expect(detail.lines[1].lineId).toBe('l2');
        expect(detail.lines[0].lineKind).toBe('core_childcare');
      });

      const req = httpMock.expectOne('/api/v1/invoices/inv-1');
      expect(req.request.method).toBe('GET');
      req.flush(detailResponse);
    });

    it('handles null calculation and empty lines gracefully', () => {
      const detailResponse = {
        invoice_id: 'inv-2',
        invoice_kind: 'monthly',
        invoice_number: null,
        child_id: 'c1',
        child_name: 'Ben',
        billing_month: '2026-05',
        status: 'draft',
        subtotal_minor: 10000,
        funded_deduction_minor: 0,
        total_due_minor: 10000,
        issued_at: null,
        created_at: '2026-06-09T10:00:00Z',
        updated_at: '2026-06-09T10:00:00Z',
      };

      service.getInvoice('inv-2').subscribe((detail) => {
        expect(detail.calculation).toBeNull();
        expect(detail.lines).toEqual([]);
        expect(detail.generatedRunExceptions).toEqual([]);
      });

      const req = httpMock.expectOne('/api/v1/invoices/inv-2');
      req.flush(detailResponse);
    });
  });
});
