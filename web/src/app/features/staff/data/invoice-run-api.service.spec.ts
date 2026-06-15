import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { InvoiceRunApiService } from './invoice-run-api.service';

describe('InvoiceRunApiService', () => {
  let service: InvoiceRunApiService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [InvoiceRunApiService],
    });
    service = TestBed.inject(InvoiceRunApiService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  describe('loadPreflight', () => {
    it('sends GET with billing_month query param', () => {
      const response = {
        billing_month: '2026-05',
        summary: {
          total_children_count: 3,
          eligible_children_count: 2,
          blocked_children_count: 1,
          included_session_count: 22,
          rounded_attended_minutes: 2640,
          funded_deduction_minor: 9000,
          total_due_minor: 42000,
        },
        eligible_children: [
          { child_id: 'c1',child_first_name: 'Ben',
child_middle_name: null,
child_last_name: null, rounded_attended_minutes: 1320, funded_deduction_minor: 4500, total_due_minor: 21000 },
        ],
        blocked_children: [
          {
            child_id: 'c2',child_first_name: 'Alice',
child_middle_name: null,
child_last_name: null,
            blockers: [{ code: 'incomplete_attendance', message: '3 sessions missing check-out' }],
          },
        ],
      };

      service.loadPreflight('2026-05').subscribe((preflight) => {
        expect(preflight.billingMonth).toBe('2026-05');
        expect(preflight.summary.eligibleChildren).toBe(2);
        expect(preflight.summary.blockedChildren).toBe(1);
        expect(preflight.eligibleChildren.length).toBe(1);
        expect(preflight.blockedChildren.length).toBe(1);
        expect(preflight.blockedChildren[0].blockers[0].code).toBe('incomplete_attendance');
        expect(preflight.blockedChildren[0].blockers[0].detail).toBe('3 sessions missing check-out');
      });

      const req = httpMock.expectOne((r) => r.url === '/api/v1/invoices/drafts/preflight');
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('billing_month')).toBe('2026-05');
      req.flush(response);
    });
  });

  describe('generateDrafts', () => {
    it('sends POST with billing_month and maps created/updated counts', () => {
      const response = {
        run_id: 'run-1',
        billing_month: '2026-05',
        status: 'completed_with_exceptions',
        summary: { eligible_count: 3, success_count: 2, blocked_count: 1, total_due_minor: 42000 },
        generated: [
          { child_id: 'c1',child_first_name: 'Ben',
child_middle_name: null,
child_last_name: null, action: 'created', invoice_id: 'inv-1' },
          { child_id: 'c3',child_first_name: 'Chloe',
child_middle_name: null,
child_last_name: null, action: 'updated', invoice_id: 'inv-3' },
        ],
        blocked: [
          {
            child_id: 'c2',child_first_name: 'Alice',
child_middle_name: null,
child_last_name: null,
            blockers: [{ code: 'incomplete_attendance', message: 'Missing sessions' }],
          },
        ],
      };

      service.generateDrafts('2026-05').subscribe((result) => {
        expect(result.generatedCount).toBe(1);
        expect(result.updatedCount).toBe(1);
        expect(result.blockedCount).toBe(1);
        expect(result.blockedChildren.length).toBe(1);
        expect(result.blockedChildren[0].childName).toBe('Alice');
      });

      const req = httpMock.expectOne('/api/v1/invoice-runs/drafts');
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({ billing_month: '2026-05' });
      req.flush(response);
    });

    it('includes child_ids when provided', () => {
      service.generateDrafts('2026-05', ['c1', 'c3']).subscribe();

      const req = httpMock.expectOne('/api/v1/invoice-runs/drafts');
      expect(req.request.body).toEqual({ billing_month: '2026-05', child_ids: ['c1', 'c3'] });
      req.flush({
        run_id: 'run-2', billing_month: '2026-05', status: 'completed',
        summary: { eligible_count: 2, success_count: 2, blocked_count: 0, total_due_minor: 0 },
        generated: [], blocked: [],
      });
    });
  });

  describe('listDrafts', () => {
    it('performs list then detail calls and maps lines/calculation', () => {
      const listResponse = {
        items: [
          { invoice_id: 'inv-1', child_id: 'c1',child_first_name: 'Ben',
child_middle_name: null,
child_last_name: null, billing_month: '2026-05', status: 'draft', subtotal_minor: 33000, funded_deduction_minor: 9000, total_due_minor: 24000, invoice_number: null, issued_at: null },
        ],
        limit: 200,
        offset: 0,
      };

      const detailResponse = {
        invoice_id: 'inv-1',
        child_id: 'c1',child_first_name: 'Ben',
child_middle_name: null,
child_last_name: null,
        billing_month: '2026-05',
        status: 'draft',
        subtotal_minor: 33000,
        funded_deduction_minor: 9000,
        total_due_minor: 24000,
        invoice_number: null,
        issued_at: null,
        calculation: { rounded_attended_minutes: 1320, funded_deduction_minutes: 360, extras_total_minor: 0 },
        lines: [
          { line_id: 'l1', line_kind: 'core_childcare', description: 'Core childcare', sort_order: 1, quantity_minutes: 1320, unit_amount_minor: 25, line_amount_minor: 33000 },
          { line_id: 'l2', line_kind: 'funded_deduction', description: 'Funded deduction', sort_order: 2, quantity_minutes: 360, unit_amount_minor: 25, line_amount_minor: -9000 },
        ],
      };

      service.listDrafts('2026-05').subscribe((drafts) => {
        expect(drafts.length).toBe(1);
        expect(drafts[0].invoiceId).toBe('inv-1');
        expect(drafts[0].childName).toBe('Ben');
        expect(drafts[0].attendedMinutes).toBe(1320);
        expect(drafts[0].extrasMinor).toBe(0);
        expect(drafts[0].lines.length).toBe(2);
        expect(drafts[0].lines[0].lineAmountMinor).toBe(33000);
        expect(drafts[0].lines[1].lineAmountMinor).toBe(-9000);
      });

      const listReq = httpMock.expectOne((r) => r.url === '/api/v1/invoices' && r.params.get('status') === 'draft');
      expect(listReq.request.params.get('billing_month')).toBe('2026-05');
      expect(listReq.request.params.get('limit')).toBe('200');
      listReq.flush(listResponse);

      const detailReq = httpMock.expectOne('/api/v1/invoices/inv-1');
      detailReq.flush(detailResponse);
    });

    it('returns empty array without detail calls when list is empty', () => {
      service.listDrafts('2026-05').subscribe((drafts) => {
        expect(drafts.length).toBe(0);
      });

      const req = httpMock.expectOne((r) => r.url === '/api/v1/invoices');
      req.flush({ items: [], limit: 200, offset: 0 });
    });
  });

  describe('bulkIssue', () => {
    it('sends selected invoice IDs and confirm true', () => {
      const response = {
        run_id: 'run-3',
        billing_month: '2026-05',
        status: 'completed',
        summary: { issued_count: 2, blocked_count: 0, total_due_minor: 48000 },
        issued: [
          { invoice_id: 'inv-1', invoice_number: 'INV-202605-0001', child_id: 'c1',child_first_name: 'Ben',
child_middle_name: null,
child_last_name: null, issued_at: '2026-06-09T12:00:00Z', total_due_minor: 24000 },
          { invoice_id: 'inv-2', invoice_number: 'INV-202605-0002', child_id: 'c3',child_first_name: 'Chloe',
child_middle_name: null,
child_last_name: null, issued_at: '2026-06-09T12:00:00Z', total_due_minor: 24000 },
        ],
        blocked: [],
      };

      service.bulkIssue('2026-05', ['inv-1', 'inv-2']).subscribe((result) => {
        expect(result.issuedCount).toBe(2);
        expect(result.totalIssuedMinor).toBe(48000);
        expect(result.issued[0].invoiceNumber).toBe('INV-202605-0001');
        expect(result.skipped.length).toBe(0);
      });

      const req = httpMock.expectOne('/api/v1/invoices/bulk-issue');
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({
        billing_month: '2026-05',
        invoice_ids: ['inv-1', 'inv-2'],
        confirm: true,
      });
      req.flush(response);
    });

    it('maps blocked invoices to skipped entries', () => {
      const response = {
        run_id: 'run-4',
        billing_month: '2026-05',
        status: 'completed_with_exceptions',
        summary: { issued_count: 1, blocked_count: 1, total_due_minor: 24000 },
        issued: [
          { invoice_id: 'inv-1', invoice_number: 'INV-202605-0001', child_id: 'c1',child_first_name: 'Ben',
child_middle_name: null,
child_last_name: null, issued_at: '2026-06-09T12:00:00Z', total_due_minor: 24000 },
        ],
        blocked: [
          { invoice_id: 'inv-2', child_id: 'c2',child_first_name: 'Alice',
child_middle_name: null,
child_last_name: null, blockers: [{ code: 'invoice_not_draft', message: 'Invoice is not in draft status' }] },
        ],
      };

      service.bulkIssue('2026-05', ['inv-1', 'inv-2']).subscribe((result) => {
        expect(result.issuedCount).toBe(1);
        expect(result.skipped.length).toBe(1);
        expect(result.skipped[0].childName).toBe('Alice');
        expect(result.skipped[0].reason).toContain('Invoice is not in draft status');
      });

      const req = httpMock.expectOne('/api/v1/invoices/bulk-issue');
      req.flush(response);
    });
  });

  describe('loadPreflight — multiple blockers', () => {
    it('maps blocked child with multiple blocker codes and messages', () => {
      const response = {
        billing_month: '2026-05',
        summary: {
          total_children_count: 2,
          eligible_children_count: 1,
          blocked_children_count: 1,
          included_session_count: 10,
          rounded_attended_minutes: 1320,
          funded_deduction_minor: 4500,
          total_due_minor: 21000,
        },
        eligible_children: [],
        blocked_children: [
          {
            child_id: 'c2',child_first_name: 'Alice',
child_middle_name: null,
child_last_name: null,
            blockers: [
              { code: 'incomplete_attendance', message: '3 sessions missing check-out' },
              { code: 'missing_funding_profile', message: 'No funding profile configured' },
            ],
          },
        ],
      };

      service.loadPreflight('2026-05').subscribe((preflight) => {
        expect(preflight.blockedChildren.length).toBe(1);
        const blocked = preflight.blockedChildren[0];
        expect(blocked.childId).toBe('c2');
        expect(blocked.childName).toBe('Alice');
        expect(blocked.blockers.length).toBe(2);
        expect(blocked.blockers[0].code).toBe('incomplete_attendance');
        expect(blocked.blockers[0].detail).toBe('3 sessions missing check-out');
        expect(blocked.blockers[1].code).toBe('missing_funding_profile');
        expect(blocked.blockers[1].detail).toBe('No funding profile configured');
      });

      const req = httpMock.expectOne((r) => r.url === '/api/v1/invoices/drafts/preflight');
      req.flush(response);
    });
  });

  describe('generateDrafts — mixed actions', () => {
    it('maps generated array with both created and updated actions', () => {
      const response = {
        run_id: 'run-5',
        billing_month: '2026-05',
        status: 'completed',
        summary: { eligible_count: 4, success_count: 4, blocked_count: 0, total_due_minor: 96000 },
        generated: [
          { child_id: 'c1',child_first_name: 'Ben',
child_middle_name: null,
child_last_name: null, action: 'created', invoice_id: 'inv-1' },
          { child_id: 'c3',child_first_name: 'Chloe',
child_middle_name: null,
child_last_name: null, action: 'updated', invoice_id: 'inv-3' },
          { child_id: 'c4',child_first_name: 'Dan',
child_middle_name: null,
child_last_name: null, action: 'created', invoice_id: 'inv-4' },
          { child_id: 'c5',child_first_name: 'Eve',
child_middle_name: null,
child_last_name: null, action: 'updated', invoice_id: 'inv-5' },
        ],
        blocked: [],
      };

      service.generateDrafts('2026-05').subscribe((result) => {
        expect(result.generatedCount).toBe(2);
        expect(result.updatedCount).toBe(2);
        expect(result.blockedCount).toBe(0);
        expect(result.blockedChildren).toEqual([]);
      });

      const req = httpMock.expectOne('/api/v1/invoice-runs/drafts');
      req.flush(response);
    });
  });

  describe('bulkIssue — edge cases', () => {
    it('maps blocked invoice with empty blockers to fallback skipped reason', () => {
      const response = {
        run_id: 'run-6',
        billing_month: '2026-05',
        status: 'completed_with_exceptions',
        summary: { issued_count: 1, blocked_count: 1, total_due_minor: 24000 },
        issued: [
          { invoice_id: 'inv-1', invoice_number: 'INV-001', child_id: 'c1',child_first_name: 'Ben',
child_middle_name: null,
child_last_name: null, issued_at: '2026-06-09T12:00:00Z', total_due_minor: 24000 },
        ],
        blocked: [
          { invoice_id: 'inv-2', child_id: 'c2',child_first_name: 'Alice',
child_middle_name: null,
child_last_name: null, blockers: [] },
        ],
      };

      service.bulkIssue('2026-05', ['inv-1', 'inv-2']).subscribe((result) => {
        expect(result.issuedCount).toBe(1);
        expect(result.skipped.length).toBe(1);
        expect(result.skipped[0].invoiceId).toBe('inv-2');
        expect(result.skipped[0].childName).toBe('Alice');
        expect(result.skipped[0].reason).toBe('Blocked');
      });

      const req = httpMock.expectOne('/api/v1/invoices/bulk-issue');
      req.flush(response);
    });
  });

  describe('issueOne', () => {
    it('sends confirm true and returns result with invoice number and child context', () => {
      const issueResponse = {
        invoice_id: 'inv-1',
        invoice_number: 'INV-202605-0001',
        issued_at: '2026-06-09T12:00:00Z',
        total_due_minor: 24000,
      };

      const detailResponse = {
        invoice_id: 'inv-1',
        child_id: 'c1',child_first_name: 'Ben',
child_middle_name: null,
child_last_name: null,
        billing_month: '2026-05',
        status: 'issued',
        subtotal_minor: 33000,
        funded_deduction_minor: 9000,
        total_due_minor: 24000,
        invoice_number: 'INV-202605-0001',
        issued_at: '2026-06-09T12:00:00Z',
        calculation: { rounded_attended_minutes: 1320, funded_deduction_minutes: 360, extras_total_minor: 0 },
        lines: [],
      };

      service.issueOne('inv-1').subscribe((result) => {
        expect(result.issuedCount).toBe(1);
        expect(result.issued[0].invoiceNumber).toBe('INV-202605-0001');
        expect(result.issued[0].childName).toBe('Ben');
        expect(result.totalIssuedMinor).toBe(24000);
      });

      const issueReq = httpMock.expectOne('/api/v1/invoices/inv-1/issue');
      expect(issueReq.request.method).toBe('POST');
      expect(issueReq.request.body).toEqual({ confirm: true });
      issueReq.flush(issueResponse);

      const detailReq = httpMock.expectOne('/api/v1/invoices/inv-1');
      detailReq.flush(detailResponse);
    });
  });
});
