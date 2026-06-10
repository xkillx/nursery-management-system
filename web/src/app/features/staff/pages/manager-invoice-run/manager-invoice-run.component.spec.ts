import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { provideRouter } from '@angular/router';
import { of, throwError } from 'rxjs';

import { InvoiceRunApiService } from '../../data/invoice-run-api.service';
import { InvoiceDraftReviewItem, IssueResultSummary } from '../../models/invoice-run.models';
import { ManagerInvoiceRunComponent } from './manager-invoice-run.component';

describe('ManagerInvoiceRunComponent', () => {
  let fixture: ComponentFixture<ManagerInvoiceRunComponent>;
  let component: ManagerInvoiceRunComponent;
  let apiService: jasmine.SpyObj<InvoiceRunApiService>;
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
    const spy = jasmine.createSpyObj<InvoiceRunApiService>('InvoiceRunApiService', [
      'loadPreflight', 'generateDrafts', 'listDrafts', 'bulkIssue', 'issueOne',
    ]);

    await TestBed.configureTestingModule({
      imports: [ManagerInvoiceRunComponent],
      providers: [
        provideRouter([]),
        { provide: InvoiceRunApiService, useValue: spy },
      ],
    }).compileComponents();

    apiService = TestBed.inject(InvoiceRunApiService) as jasmine.SpyObj<InvoiceRunApiService>;
    fixture = TestBed.createComponent(ManagerInvoiceRunComponent);
    component = fixture.componentInstance;
    native = fixture.nativeElement;
  });

  it('renders page header', () => {
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
    fixture.detectChanges();
    expect(native.textContent).toContain('Invoice run');
  });

  it('does not contain ecommerce terminology', () => {
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
    fixture.detectChanges();
    const banned = ['Ecommerce', 'Sales', 'Orders', 'Customers', 'Products', 'Revenue'];
    for (const term of banned) {
      expect(native.textContent).not.toContain(term);
    }
  });

  it('shows preflight summary metrics', () => {
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
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
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
    fixture.detectChanges();

    expect(native.textContent).toContain('Yusuf Ali');
    expect(native.textContent).toContain('Incomplete attendance');
    const link = native.querySelector('a[href*="attendance-corrections"]');
    expect(link).toBeTruthy();
  });

  it('shows generate button when eligible children exist', () => {
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
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
    apiService.loadPreflight.and.returnValue(of(noEligible));
    fixture.detectChanges();

    expect(native.textContent).toContain('No eligible children');
    const genBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Generate'));
    expect(genBtn).toBeFalsy();
  });

  it('generates drafts and shows review', () => {
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
    apiService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05', generatedCount: 4, updatedCount: 0, blockedCount: 1, blockedChildren: [],
    }));
    apiService.listDrafts.and.returnValue(of([draftAmira, draftArjun]));

    fixture.detectChanges();

    const genBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Generate'));
    genBtn?.click();
    fixture.detectChanges();

    expect(apiService.generateDrafts).toHaveBeenCalledWith('2026-05');
    expect(native.textContent).toContain('Amira Hassan');
    expect(native.textContent).toContain('Arjun Patel');
  });

  it('shows draft review without edit controls', () => {
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
    apiService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05', generatedCount: 2, updatedCount: 0, blockedCount: 0, blockedChildren: [],
    }));
    apiService.listDrafts.and.returnValue(of([draftAmira, draftArjun]));

    fixture.detectChanges();
    const genBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Generate'));
    genBtn?.click();
    fixture.detectChanges();

    const inputs = native.querySelectorAll('input[type="text"], input[type="number"], textarea');
    expect(inputs.length).toBe(0);
  });

  it('selects all ready drafts by default after generation', () => {
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
    apiService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05', generatedCount: 2, updatedCount: 0, blockedCount: 0, blockedChildren: [],
    }));
    apiService.listDrafts.and.returnValue(of([draftAmira, draftArjun]));

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
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
    apiService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05', generatedCount: 2, updatedCount: 0, blockedCount: 0, blockedChildren: [],
    }));
    apiService.listDrafts.and.returnValue(of([draftAmira, draftArjun]));

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
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
    apiService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05', generatedCount: 2, updatedCount: 0, blockedCount: 0, blockedChildren: [],
    }));
    apiService.listDrafts.and.returnValue(of([draftAmira, draftArjun]));

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
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
    apiService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05', generatedCount: 2, updatedCount: 0, blockedCount: 0, blockedChildren: [],
    }));
    apiService.listDrafts.and.returnValue(of([draftAmira, draftArjun]));

    const issueResult: IssueResultSummary = {
      billingMonth: '2026-05', issuedCount: 2, totalIssuedMinor: 69600,
      issued: [
        { invoiceId: 'inv-c1-2026-05', childId: 'c1', childName: 'Amira Hassan', invoiceNumber: 'INV-202605-0001', issuedAt: '2026-06-09T12:00:00Z', totalMinor: 33600 },
        { invoiceId: 'inv-c2-2026-05', childId: 'c2', childName: 'Arjun Patel', invoiceNumber: 'INV-202605-0002', issuedAt: '2026-06-09T12:00:00Z', totalMinor: 36000 },
      ],
      skipped: [],
    };
    apiService.bulkIssue.and.returnValue(of(issueResult));

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
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
    apiService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05', generatedCount: 2, updatedCount: 0, blockedCount: 0, blockedChildren: [],
    }));
    apiService.listDrafts.and.returnValue(of([draftAmira, draftArjun]));

    const singleResult: IssueResultSummary = {
      billingMonth: '2026-05', issuedCount: 1, totalIssuedMinor: 33600,
      issued: [
        { invoiceId: 'inv-c1-2026-05', childId: 'c1', childName: 'Amira Hassan', invoiceNumber: 'INV-202605-0001', issuedAt: '2026-06-09T12:00:00Z', totalMinor: 33600 },
      ],
      skipped: [],
    };
    apiService.issueOne.and.returnValue(of(singleResult));

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
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
    fixture.detectChanges();

    const input = native.querySelector('input[type="month"]') as HTMLInputElement;
    input.value = '2026-03';
    input.dispatchEvent(new Event('input'));
    input.dispatchEvent(new Event('change'));
    fixture.detectChanges();

    expect(apiService.loadPreflight).toHaveBeenCalledWith('2026-03');
  });

  it('handles preflight load failure', () => {
    apiService.loadPreflight.and.returnValue(throwError(() => new Error('Network error')));
    fixture.detectChanges();

    expect(fixture.componentInstance.errorMessage).toBe('Something went wrong. Try again.');
    expect(fixture.componentInstance.isLoadingPreflight).toBe(false);
  });

  it('generation with blocked children still enables draft review', () => {
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
    apiService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05',
      generatedCount: 2,
      updatedCount: 0,
      blockedCount: 1,
      blockedChildren: [
        { childId: 'cb1', childName: 'Yusuf Ali', blockers: [{ code: 'incomplete_attendance', detail: 'Missing' }] },
      ],
    }));
    apiService.listDrafts.and.returnValue(of([draftAmira, draftArjun]));

    fixture.detectChanges();
    const genBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Generate'));
    genBtn?.click();
    fixture.detectChanges();

    expect(native.textContent).toContain('Amira Hassan');
    expect(native.textContent).toContain('Generated 2, updated 0. 1 blocked.');
  });

  it('month change resets all review state', () => {
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
    apiService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05', generatedCount: 2, updatedCount: 0, blockedCount: 0, blockedChildren: [],
    }));
    apiService.listDrafts.and.returnValue(of([draftAmira, draftArjun]));

    fixture.detectChanges();
    const genBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Generate'));
    genBtn?.click();
    fixture.detectChanges();

    // Verify we are in review step
    expect(component.currentStep).toBe('review');
    expect(component.generationResult).toBeTruthy();
    expect(component.drafts.length).toBe(2);

    // Set some selection/expand/confirm state
    component.expandedInvoiceIds.add(draftAmira.invoiceId);
    component.showBulkConfirm = true;

    // Trigger month change
    apiService.loadPreflight.calls.reset();
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
    const input = native.querySelector('input[type="month"]') as HTMLInputElement;
    input.value = '2026-04';
    input.dispatchEvent(new Event('input'));
    input.dispatchEvent(new Event('change'));
    fixture.detectChanges();

    expect(component.currentStep).toBe('preflight');
    expect(component.generationResult).toBeNull();
    expect(component.drafts.length).toBe(0);
    expect(component.issueResult).toBeNull();
    expect(component.selectedInvoiceIds.size).toBe(0);
    expect(component.expandedInvoiceIds.size).toBe(0);
    expect(component.showBulkConfirm).toBe(false);
    expect(component.singleIssueDraft).toBeNull();
    expect(apiService.loadPreflight).toHaveBeenCalledWith('2026-04');
  });

  it('generation failure leaves preflight step', () => {
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
    apiService.generateDrafts.and.returnValue(throwError(() => new Error('Network error')));

    fixture.detectChanges();
    const genBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Generate'));
    genBtn?.click();
    fixture.detectChanges();

    expect(component.currentStep).toBe('preflight');
    expect(component.isGenerating).toBe(false);
    expect(component.errorMessage).toBeTruthy();
  });

  it('draft-list failure after successful generation still moves to review', () => {
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
    apiService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05', generatedCount: 2, updatedCount: 0, blockedCount: 0, blockedChildren: [],
    }));
    apiService.listDrafts.and.returnValue(throwError(() => new Error('List failed')));

    fixture.detectChanges();
    const genBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Generate'));
    genBtn?.click();
    fixture.detectChanges();

    expect(component.generationResult).toBeTruthy();
    expect(component.currentStep).toBe('review');
    expect(component.isGenerating).toBe(false);
    expect(component.errorMessage).toBeTruthy();
  });

  it('toggleAllReady only selects drafts with draft status', () => {
    const issuedDraft: InvoiceDraftReviewItem = {
      ...draftAmira, status: 'issued' as const, invoiceId: 'inv-issued',
    };
    const draftsMixed = [issuedDraft, draftArjun];

    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
    apiService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05', generatedCount: 2, updatedCount: 0, blockedCount: 0, blockedChildren: [],
    }));
    apiService.listDrafts.and.returnValue(of(draftsMixed));

    fixture.detectChanges();
    const genBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Generate'));
    genBtn?.click();
    fixture.detectChanges();

    // Auto-select on generation already selects only ready drafts
    expect(component.selectedInvoiceIds.has('inv-issued')).toBe(false);
    expect(component.selectedInvoiceIds.has(draftArjun.invoiceId)).toBe(true);

    // Clear and toggle all to verify toggleAllReady
    component.selectedInvoiceIds.clear();
    component.toggleAllReady();

    expect(component.selectedInvoiceIds.has('inv-issued')).toBe(false);
    expect(component.selectedInvoiceIds.has(draftArjun.invoiceId)).toBe(true);

    // selectedTotalMinor should only include the ready draft
    expect(component.selectedTotalMinor).toBe(draftArjun.netDueMinor);
  });

  it('after bulk issue, refreshed draft list rebuilds selection for remaining drafts', () => {
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
    apiService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05', generatedCount: 2, updatedCount: 0, blockedCount: 0, blockedChildren: [],
    }));
    apiService.listDrafts.and.returnValue(of([draftAmira, draftArjun]));

    const issueResult: IssueResultSummary = {
      billingMonth: '2026-05', issuedCount: 2, totalIssuedMinor: 69600,
      issued: [
        { invoiceId: 'inv-c1-2026-05', childId: 'c1', childName: 'Amira Hassan', invoiceNumber: 'INV-202605-0001', issuedAt: '2026-06-09T12:00:00Z', totalMinor: 33600 },
        { invoiceId: 'inv-c2-2026-05', childId: 'c2', childName: 'Arjun Patel', invoiceNumber: 'INV-202605-0002', issuedAt: '2026-06-09T12:00:00Z', totalMinor: 36000 },
      ],
      skipped: [],
    };
    apiService.bulkIssue.and.returnValue(of(issueResult));

    fixture.detectChanges();
    const genBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Generate'));
    genBtn?.click();
    fixture.detectChanges();

    const issueBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Issue selected'));
    issueBtn?.click();
    fixture.detectChanges();

    // After bulk issue, listDrafts returns only 1 remaining draft
    const remainingDraft: InvoiceDraftReviewItem = {
      ...draftArjun, invoiceId: 'inv-c3-2026-05', childId: 'c3', childName: 'Emma Chen', netDueMinor: 42000,
    };
    apiService.listDrafts.and.returnValue(of([remainingDraft]));

    const confirmBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Issue invoices'));
    confirmBtn?.click();
    fixture.detectChanges();

    expect(component.selectedInvoiceIds.size).toBe(1);
    expect(component.selectedInvoiceIds.has('inv-c3-2026-05')).toBe(true);
  });

  it('bulk issue failure closes confirmation and keeps review state', () => {
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
    apiService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05', generatedCount: 2, updatedCount: 0, blockedCount: 0, blockedChildren: [],
    }));
    apiService.listDrafts.and.returnValue(of([draftAmira, draftArjun]));
    apiService.bulkIssue.and.returnValue(throwError(() => new Error('Issue failed')));

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

    expect(component.isIssuing).toBe(false);
    expect(component.showBulkConfirm).toBe(false);
    expect(component.drafts.length).toBe(2);
    expect(component.errorMessage).toBeTruthy();
  });

  it('single issue accumulates into existing issueResult', () => {
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
    apiService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05', generatedCount: 2, updatedCount: 0, blockedCount: 0, blockedChildren: [],
    }));
    apiService.listDrafts.and.returnValue(of([draftAmira, draftArjun]));

    const firstResult: IssueResultSummary = {
      billingMonth: '2026-05', issuedCount: 1, totalIssuedMinor: 33600,
      issued: [
        { invoiceId: 'inv-c1-2026-05', childId: 'c1', childName: 'Amira Hassan', invoiceNumber: 'INV-202605-0001', issuedAt: '2026-06-09T12:00:00Z', totalMinor: 33600 },
      ],
      skipped: [],
    };
    const secondResult: IssueResultSummary = {
      billingMonth: '2026-05', issuedCount: 1, totalIssuedMinor: 36000,
      issued: [
        { invoiceId: 'inv-c2-2026-05', childId: 'c2', childName: 'Arjun Patel', invoiceNumber: 'INV-202605-0002', issuedAt: '2026-06-09T12:01:00Z', totalMinor: 36000 },
      ],
      skipped: [],
    };
    apiService.issueOne.and.returnValues(of(firstResult), of(secondResult));

    fixture.detectChanges();
    const genBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Generate'));
    genBtn?.click();
    fixture.detectChanges();

    // First single issue
    const issueLinks = Array.from(native.querySelectorAll('button')).filter(b => b.textContent?.trim() === 'Issue');
    issueLinks[0].click();
    fixture.detectChanges();

    const confirmBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Issue invoice'));
    confirmBtn?.click();
    fixture.detectChanges();

    expect(component.issueResult).toBeTruthy();
    expect(component.issueResult!.issuedCount).toBe(1);
    expect(component.issueResult!.issued.length).toBe(1);

    // Second single issue
    const issueLinksAfter = Array.from(native.querySelectorAll('button')).filter(b => b.textContent?.trim() === 'Issue');
    issueLinksAfter[0].click();
    fixture.detectChanges();

    const confirmBtn2 = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Issue invoice'));
    confirmBtn2?.click();
    fixture.detectChanges();

    expect(component.issueResult!.issuedCount).toBe(2);
    expect(component.issueResult!.issued.length).toBe(2);
    expect(component.issueResult!.totalIssuedMinor).toBe(33600 + 36000);
  });

  it('single issue failure clears singleIssueDraft and keeps review state', () => {
    apiService.loadPreflight.and.returnValue(of(preflight2026_05));
    apiService.generateDrafts.and.returnValue(of({
      billingMonth: '2026-05', generatedCount: 2, updatedCount: 0, blockedCount: 0, blockedChildren: [],
    }));
    apiService.listDrafts.and.returnValue(of([draftAmira, draftArjun]));
    apiService.issueOne.and.returnValue(throwError(() => new Error('Issue failed')));

    fixture.detectChanges();
    const genBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Generate'));
    genBtn?.click();
    fixture.detectChanges();

    const issueLinks = Array.from(native.querySelectorAll('button')).filter(b => b.textContent?.trim() === 'Issue');
    issueLinks[0].click();
    fixture.detectChanges();

    const confirmBtn = Array.from(native.querySelectorAll('button')).find(b => b.textContent?.includes('Issue invoice'));
    confirmBtn?.click();
    fixture.detectChanges();

    expect(component.singleIssueDraft).toBeNull();
    expect(component.isIssuing).toBe(false);
    expect(component.drafts.length).toBe(2);
    expect(component.errorMessage).toBeTruthy();
  });
});
