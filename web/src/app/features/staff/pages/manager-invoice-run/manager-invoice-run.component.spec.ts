import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';

import { InvoiceRunMockService } from '../../data/invoice-run-mock.service';
import { InvoiceDraftReviewItem, IssueResultSummary } from '../../models/invoice-run.models';
import { ManagerInvoiceRunComponent } from './manager-invoice-run.component';

describe('ManagerInvoiceRunComponent', () => {
  let fixture: ComponentFixture<ManagerInvoiceRunComponent>;
  let mockService: jasmine.SpyObj<InvoiceRunMockService>;
  let native: HTMLElement;

  const preflight2026_05 = {
    billingMonth: '2026-05',
    summary: {
      totalChildren: 6, eligibleChildren: 4, blockedChildren: 2,
      sessionsIncluded: 44, attendedMinutes: 7920, fundedDeductionMinor: 21000, totalDueMinor: 111600,
    },
    eligibleChildren: [
      { childId: 'c1', childName: 'Amira Hassan', attendedMinutes: 2760, fundedDeductionMinor: 12000, totalDueMinor: 33600 },
      { childId: 'c2', childName: 'Arjun Patel', attendedMinutes: 2640, fundedDeductionMinor: 9000, totalDueMinor: 36000 },
      { childId: 'c3', childName: 'Emma Chen', attendedMinutes: 2520, fundedDeductionMinor: 0, totalDueMinor: 42000 },
      { childId: 'c4', childName: 'Noah Williams', attendedMinutes: 0, fundedDeductionMinor: 0, totalDueMinor: 0 },
    ],
    blockedChildren: [
      { childId: 'cb1', childName: 'Yusuf Ali', blockers: [{ code: 'incomplete_attendance' as const, detail: '3 sessions missing check-out' }] },
    ],
  };

  const draftAmira: InvoiceDraftReviewItem = {
    invoiceId: 'inv-c1-2026-05', childId: 'c1', childName: 'Amira Hassan',
    billingMonth: '2026-05', status: 'draft', attendedMinutes: 2760,
    fundedDeductionMinor: 12000, extrasMinor: 0, subtotalMinor: 33600, netDueMinor: 33600,
    lines: [], invoiceNumber: null, issuedAt: null, generationAction: 'created',
  };

  const draftArjun: InvoiceDraftReviewItem = {
    invoiceId: 'inv-c2-2026-05', childId: 'c2', childName: 'Arjun Patel',
    billingMonth: '2026-05', status: 'draft', attendedMinutes: 2640,
    fundedDeductionMinor: 9000, extrasMinor: 0, subtotalMinor: 36000, netDueMinor: 36000,
    lines: [], invoiceNumber: null, issuedAt: null, generationAction: 'created',
  };

  beforeEach(async () => {
    const spy = jasmine.createSpyObj<InvoiceRunMockService>('InvoiceRunMockService', [
      'loadPreflight', 'generateDrafts', 'listDrafts', 'bulkIssue', 'issueOne', 'resetMonth',
    ]);

    await TestBed.configureTestingModule({
      imports: [ManagerInvoiceRunComponent],
      providers: [
        provideRouter([]),
        { provide: InvoiceRunMockService, useValue: spy },
      ],
    }).compileComponents();

    mockService = TestBed.inject(InvoiceRunMockService) as jasmine.SpyObj<InvoiceRunMockService>;
    fixture = TestBed.createComponent(ManagerInvoiceRunComponent);
    native = fixture.nativeElement;
  });

  it('renders page header', () => {
    mockService.loadPreflight.and.returnValue(of(preflight2026_05));
    fixture.detectChanges();
    expect(native.textContent).toContain('Invoice run');
  });

  it('does not contain ecommerce terminology', () => {
    mockService.loadPreflight.and.returnValue(of(preflight2026_05));
    fixture.detectChanges();
    const banned = ['Ecommerce', 'Sales', 'Orders', 'Customers', 'Products', 'Revenue'];
    for (const term of banned) {
      expect(native.textContent).not.toContain(term);
    }
  });

  it('shows preflight summary metrics', () => {
    mockService.loadPreflight.and.returnValue(of(preflight2026_05));
    fixture.detectChanges();

    expect(native.textContent).toContain('Children in month');
    expect(native.textContent).toContain('Ready for draft');
    expect(native.textContent).toContain('Exceptions');
    expect(native.textContent).toContain('Sessions included');
    expect(native.textContent).toContain('Attended time');
    expect(native.textContent).toContain('Funded deduction');
    expect(native.textContent).toContain('Estimated total');
  });

  it('shows exception table with next actions', () => {
    mockService.loadPreflight.and.returnValue(of(preflight2026_05));
    fixture.detectChanges();

    expect(native.textContent).toContain('Yusuf Ali');
    expect(native.textContent).toContain('Incomplete attendance');
    const link = native.querySelector('a[href*="attendance-corrections"]');
    expect(link).toBeTruthy();
  });

  it('shows generate button when eligible children exist', () => {
    mockService.loadPreflight.and.returnValue(of(preflight2026_05));
    fixture.detectChanges();

    const btn = native.querySelector('button');
    expect(btn?.textContent).toContain('Generate draft invoices');
  });

  it('disables generate when no eligible children', () => {
    const noEligible = {
      ...preflight2026_05,
      summary: { ...preflight2026_05.summary, eligibleChildren: 0 },
      eligibleChildren: [],
    };
    mockService.loadPreflight.and.returnValue(of(noEligible));
    fixture.detectChanges();

    expect(native.textContent).toContain('No eligible children');
    const genBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Generate'));
    expect(genBtn).toBeFalsy();
  });

  it('generates drafts and shows review', () => {
    mockService.loadPreflight.and.returnValue(of(preflight2026_05));
    mockService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05', generatedCount: 4, updatedCount: 0, blockedCount: 1, blockedChildren: [],
    }));
    mockService.listDrafts.and.returnValue(of([draftAmira, draftArjun]));

    fixture.detectChanges();

    const genBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Generate'));
    genBtn?.click();
    fixture.detectChanges();

    expect(mockService.generateDrafts).toHaveBeenCalledWith('2026-05');
    expect(native.textContent).toContain('Amira Hassan');
    expect(native.textContent).toContain('Arjun Patel');
  });

  it('shows draft review without edit controls', () => {
    mockService.loadPreflight.and.returnValue(of(preflight2026_05));
    mockService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05', generatedCount: 2, updatedCount: 0, blockedCount: 0, blockedChildren: [],
    }));
    mockService.listDrafts.and.returnValue(of([draftAmira, draftArjun]));

    fixture.detectChanges();
    const genBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Generate'));
    genBtn?.click();
    fixture.detectChanges();

    const inputs = native.querySelectorAll('input[type="text"], input[type="number"], textarea');
    expect(inputs.length).toBe(0);
  });

  it('selects all ready drafts by default after generation', () => {
    mockService.loadPreflight.and.returnValue(of(preflight2026_05));
    mockService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05', generatedCount: 2, updatedCount: 0, blockedCount: 0, blockedChildren: [],
    }));
    mockService.listDrafts.and.returnValue(of([draftAmira, draftArjun]));

    fixture.detectChanges();
    const genBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Generate'));
    genBtn?.click();
    fixture.detectChanges();

    const checkboxes = native.querySelectorAll('tbody input[type="checkbox"]') as NodeListOf<HTMLInputElement>;
    for (const cb of Array.from(checkboxes)) {
      expect(cb.checked).toBe(true);
    }
  });

  it('deselecting a draft changes selected count', () => {
    mockService.loadPreflight.and.returnValue(of(preflight2026_05));
    mockService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05', generatedCount: 2, updatedCount: 0, blockedCount: 0, blockedChildren: [],
    }));
    mockService.listDrafts.and.returnValue(of([draftAmira, draftArjun]));

    fixture.detectChanges();
    const genBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Generate'));
    genBtn?.click();
    fixture.detectChanges();

    fixture.componentInstance.toggleDraftSelection(draftAmira.invoiceId);
    fixture.detectChanges();

    const issueBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Issue selected'));
    expect(issueBtn?.textContent).toContain('1');
  });

  it('opens bulk confirmation with count and total', () => {
    mockService.loadPreflight.and.returnValue(of(preflight2026_05));
    mockService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05', generatedCount: 2, updatedCount: 0, blockedCount: 0, blockedChildren: [],
    }));
    mockService.listDrafts.and.returnValue(of([draftAmira, draftArjun]));

    fixture.detectChanges();
    const genBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Generate'));
    genBtn?.click();
    fixture.detectChanges();

    const issueBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Issue selected'));
    issueBtn?.click();
    fixture.detectChanges();

    expect(native.textContent).toContain('2 invoices');
    expect(native.textContent).toContain('immutable');
  });

  it('confirming bulk issue shows assigned invoice numbers', () => {
    mockService.loadPreflight.and.returnValue(of(preflight2026_05));
    mockService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05', generatedCount: 2, updatedCount: 0, blockedCount: 0, blockedChildren: [],
    }));
    mockService.listDrafts.and.returnValue(of([draftAmira, draftArjun]));

    const issueResult: IssueResultSummary = {
      billingMonth: '2026-05', issuedCount: 2, totalIssuedMinor: 69600,
      issued: [
        { invoiceId: 'inv-c1-2026-05', childId: 'c1', childName: 'Amira Hassan', invoiceNumber: 'INV-202605-0001', issuedAt: '2026-06-09T12:00:00Z', totalMinor: 33600 },
        { invoiceId: 'inv-c2-2026-05', childId: 'c2', childName: 'Arjun Patel', invoiceNumber: 'INV-202605-0002', issuedAt: '2026-06-09T12:00:00Z', totalMinor: 36000 },
      ],
      skipped: [],
    };
    mockService.bulkIssue.and.returnValue(of(issueResult));

    fixture.detectChanges();
    const genBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Generate'));
    genBtn?.click();
    fixture.detectChanges();

    const issueBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Issue selected'));
    issueBtn?.click();
    fixture.detectChanges();

    const confirmBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Issue invoices'));
    confirmBtn?.click();
    fixture.detectChanges();

    expect(native.textContent).toContain('INV-202605-0001');
    expect(native.textContent).toContain('INV-202605-0002');
  });

  it('one-by-one issue opens confirmation and issues single draft', () => {
    mockService.loadPreflight.and.returnValue(of(preflight2026_05));
    mockService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05', generatedCount: 2, updatedCount: 0, blockedCount: 0, blockedChildren: [],
    }));
    mockService.listDrafts.and.returnValue(of([draftAmira, draftArjun]));

    const singleResult: IssueResultSummary = {
      billingMonth: '2026-05', issuedCount: 1, totalIssuedMinor: 33600,
      issued: [
        { invoiceId: 'inv-c1-2026-05', childId: 'c1', childName: 'Amira Hassan', invoiceNumber: 'INV-202605-0001', issuedAt: '2026-06-09T12:00:00Z', totalMinor: 33600 },
      ],
      skipped: [],
    };
    mockService.issueOne.and.returnValue(of(singleResult));

    fixture.detectChanges();
    const genBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Generate'));
    genBtn?.click();
    fixture.detectChanges();

    const issueLinks = Array.from(native.querySelectorAll('button')).filter(b => b.textContent?.trim() === 'Issue');
    expect(issueLinks.length).toBeGreaterThanOrEqual(1);
    issueLinks[0].click();
    fixture.detectChanges();

    expect(native.textContent).toContain('Amira Hassan');
    expect(native.textContent).toContain('immutable');
  });

  it('month change reloads preflight and resets draft state', () => {
    mockService.loadPreflight.and.returnValue(of(preflight2026_05));
    fixture.detectChanges();

    const input = native.querySelector('input[type="month"]') as HTMLInputElement;
    input.value = '2026-03';
    input.dispatchEvent(new Event('input'));
    input.dispatchEvent(new Event('change'));
    fixture.detectChanges();

    expect(mockService.resetMonth).toHaveBeenCalledWith('2026-03');
    expect(mockService.loadPreflight).toHaveBeenCalledWith('2026-03');
  });
});
