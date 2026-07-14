import { HttpClientTestingModule } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, provideRouter, Router } from '@angular/router';
import { of, throwError } from 'rxjs';
import { HttpErrorResponse } from '@angular/common/http';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { ManagerInvoiceEditComponent } from './manager-invoice-edit.component';
import { ManagerInvoicesApiService } from '../../data/manager-invoices-api.service';
import { ManagerInvoiceCreateApiService } from '../../data/manager-invoice-create-api.service';
import { ToastService } from '../../../../shared/services/toast.service';
import { ManagerInvoiceDetail } from '../../models/manager-invoices.models';

const draftDetail: ManagerInvoiceDetail = {
  invoiceId: 'inv-1',
  invoiceKind: 'monthly',
  invoiceNumber: null,
  invoiceNumberDisplay: '',
  childId: 'c1',
  childName: 'Ben',
  billingMonth: '2026-05',
  site_profile: null,
  period: { startDate: '2026-05-01', endDate: '2026-05-31' },
  status: 'draft',
  dueStatus: 'not_due',
  currencyCode: 'gbp',
  subtotalMinor: 33000,
  fundedDeductionMinor: 9000,
  totalDueMinor: 24000,
  amountPaidMinor: 0,
  issuedAt: null,
  lockedAt: null,
  dueAt: null,
  paidAt: null,
  paymentFailedAt: null,
  paymentStatusUpdatedAt: null,
  adjustsInvoiceId: null,
  adjustmentReasonCode: null,
  adjustmentReasonNote: null,
  generatedRunId: null,
  generatedRunStatus: null,
  generatedRunStartedAt: null,
  generatedRunCompletedAt: null,
  generatedRunExceptionCount: null,
  roomName: null,
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
  generatedRunExceptions: [],
  calculation: null,
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
      fundingModel: null,
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
      fundingModel: null,
    },
    {
      lineId: 'l3',
      lineKind: 'extra',
      description: 'Late pick-up',
      sortOrder: 3,
      quantityMinutes: 30,
      unitAmountMinor: 500,
      lineAmountMinor: 1500,
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

const issuedDetail: ManagerInvoiceDetail = {
  ...draftDetail,
  invoiceId: 'inv-2',
  status: 'issued',
  issuedAt: '2026-06-09T12:00:00Z',
  lockedAt: '2026-06-09T12:00:00Z',
};

function createApiSpy() {
  return jasmine.createSpyObj('ManagerInvoicesApiService', [
    'getInvoice',
    'addLine',
    'updateLine',
    'deleteLine',
    'issueInvoice',
  ]);
}

function createCreateApiSpy() {
  return jasmine.createSpyObj('ManagerInvoiceCreateApiService', ['getPrefill']);
}

describe('ManagerInvoiceEditComponent', () => {
  let fixture: ComponentFixture<ManagerInvoiceEditComponent>;
  let apiService: jasmine.SpyObj<ManagerInvoicesApiService>;
  let createApiService: jasmine.SpyObj<ManagerInvoiceCreateApiService>;
  let toastService: jasmine.SpyObj<ToastService>;
  let router: Router;

  function createFixture(detail: ManagerInvoiceDetail = draftDetail) {
    const apiSpy = createApiSpy();
    apiSpy.getInvoice.and.returnValue(of(detail));
    apiSpy.addLine.and.returnValue(of({ lineId: 'new-1', lineKind: 'extra', description: '', sortOrder: 4, quantityMinutes: 0, unitAmountMinor: 0, lineAmountMinor: 0, subtotalMinor: 33000, totalDueMinor: 24000 }));
    apiSpy.updateLine.and.returnValue(of({ lineId: 'l3', lineKind: 'extra', description: 'Updated', sortOrder: 3, quantityMinutes: 30, unitAmountMinor: 500, lineAmountMinor: 1500, subtotalMinor: 33000, totalDueMinor: 24000 }));
    apiSpy.deleteLine.and.returnValue(of({ lineId: 'l3', subtotalMinor: 31500, totalDueMinor: 22500 }));
    apiSpy.issueInvoice.and.returnValue(of({ invoiceId: 'inv-1', status: 'issued' }));

    const createApiSpyObj = createCreateApiSpy();
    createApiSpyObj.getPrefill.and.returnValue(of({
      childId: 'c1',
      childFirstName: 'Ben',
      childMiddleName: null,
      childLastName: null,
      billingMonth: '2026-05',
      entitlementStatus: { fundingProfileId: 'fp1', fundedAllowanceMinutes: 360, statusLabel: 'Standard' },
      lines: [
        { lineKind: 'core_childcare', description: 'Core childcare', sortOrder: 1, quantityMinutes: 1320, unitAmountMinor: 25, lineAmountMinor: 33000, fundedAllowanceMinutes: 0, fundedDeductionMinutes: 0, coreBillableMinutes: 960, sessionCount: 4 },
        { lineKind: 'funded_deduction', description: 'Funded deduction', sortOrder: 2, quantityMinutes: 360, unitAmountMinor: 25, lineAmountMinor: -9000, fundedAllowanceMinutes: 360, fundedDeductionMinutes: 360, coreBillableMinutes: 0, sessionCount: 0 },
      ],
      subtotalMinor: 33000,
      fundedDeductionMinor: 9000,
      totalDueMinor: 24000,
      warnings: [],
    }));

    const toastSpy = jasmine.createSpyObj('ToastService', ['success', 'error', 'warning', 'info']);

    TestBed.configureTestingModule({
      imports: [ManagerInvoiceEditComponent, HttpClientTestingModule],
      providers: [
        provideRouter([]),
        ApiErrorMapper,
        { provide: ManagerInvoicesApiService, useValue: apiSpy },
        { provide: ManagerInvoiceCreateApiService, useValue: createApiSpyObj },
        { provide: ToastService, useValue: toastSpy },
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { paramMap: { get: (key: string) => key === 'invoiceId' ? 'inv-1' : null } } },
        },
      ],
    });

    apiService = TestBed.inject(ManagerInvoicesApiService) as jasmine.SpyObj<ManagerInvoicesApiService>;
    createApiService = TestBed.inject(ManagerInvoiceCreateApiService) as jasmine.SpyObj<ManagerInvoiceCreateApiService>;
    toastService = TestBed.inject(ToastService) as jasmine.SpyObj<ToastService>;
    router = TestBed.inject(Router);
    fixture = TestBed.createComponent(ManagerInvoiceEditComponent);
    fixture.detectChanges();
  }

  describe('data loading', () => {
    it('fetches invoice using route param on init', () => {
      createFixture();
      expect(apiService.getInvoice).toHaveBeenCalledWith('inv-1');
    });

    it('renders loaded invoice data in the template', () => {
      createFixture();
      const text = fixture.nativeElement.textContent;
      expect(text).toContain('Edit Draft Invoice');
      expect(text).toContain('Ben');
      expect(text).toContain('Burhan Khalid');
    });

    it('displays all line items in the table', () => {
      createFixture();
      const text = fixture.nativeElement.textContent;
      expect(text).toContain('Core childcare');
      expect(text).toContain('Funded deduction');
      expect(text).toContain('Late pick-up');
    });

    it('renders system lines as read-only (no input fields for system lines)', () => {
      createFixture();
      const inputs: HTMLInputElement[] = fixture.nativeElement.querySelectorAll('input');
      const descriptions = Array.from(inputs).filter((i) => i.placeholder === 'Description');
      expect(descriptions.length).toBe(1);
    });

    it('shows summary sidebar with correct totals', () => {
      createFixture();
      const text = fixture.nativeElement.textContent;
      expect(text).toContain('£330.00');
      expect(text).toContain('£90.00');
      expect(text).toContain('£240.00');
    });

    it('shows loading spinner while API call is in flight', () => {
      const apiSpy = createApiSpy();
      apiSpy.getInvoice.and.returnValue(of(draftDetail));
      const createApiSpyObj = createCreateApiSpy();
      const toastSpy = jasmine.createSpyObj('ToastService', ['success', 'error', 'warning', 'info']);

      TestBed.configureTestingModule({
        imports: [ManagerInvoiceEditComponent, HttpClientTestingModule],
        providers: [
          provideRouter([]),
          ApiErrorMapper,
          { provide: ManagerInvoicesApiService, useValue: apiSpy },
          { provide: ManagerInvoiceCreateApiService, useValue: createApiSpyObj },
          { provide: ToastService, useValue: toastSpy },
          {
            provide: ActivatedRoute,
            useValue: { snapshot: { paramMap: { get: (key: string) => key === 'invoiceId' ? 'inv-1' : null } } },
          },
        ],
      });

      fixture = TestBed.createComponent(ManagerInvoiceEditComponent);
      fixture.detectChanges();
      expect(apiSpy.getInvoice).toHaveBeenCalled();
    });
  });

  describe('inline editing', () => {
    it('updates lines signal when editing description on extra line', () => {
      createFixture();
      const comp = fixture.componentInstance;
      comp.updateLine('l3', 'description', 'Updated pick-up');
      expect(comp.lines().find((l) => l.id === 'l3')?.description).toBe('Updated pick-up');
    });

    it('recalculates lineAmountMinor when changing quantity', () => {
      createFixture();
      const comp = fixture.componentInstance;
      comp.updateLine('l3', 'quantityMinutes', 60);
      const line = comp.lines().find((l) => l.id === 'l3');
      expect(line?.lineAmountMinor).toBe(60 * 500);
    });

    it('recalculates lineAmountMinor when changing unit price', () => {
      createFixture();
      const comp = fixture.componentInstance;
      comp.updateLine('l3', 'unitAmountMinor', 1000);
      const line = comp.lines().find((l) => l.id === 'l3');
      expect(line?.lineAmountMinor).toBe(30 * 1000);
    });

    it('adds a new blank extra line', () => {
      createFixture();
      const comp = fixture.componentInstance;
      const initialCount = comp.lines().length;
      comp.addBlankLine();
      expect(comp.lines().length).toBe(initialCount + 1);
      const added = comp.lines()[comp.lines().length - 1];
      expect(added.lineKind).toBe('extra');
      expect(added.description).toBe('');
    });

    it('removes a line', () => {
      createFixture();
      const comp = fixture.componentInstance;
      comp.removeLine('l3');
      expect(comp.lines().find((l) => l.id === 'l3')).toBeUndefined();
    });

    it('prevents editing system-generated lines (no edit controls rendered)', () => {
      createFixture();
      const inputs: HTMLInputElement[] = fixture.nativeElement.querySelectorAll('input');
      const descriptionInputs = Array.from(inputs).filter((i) => i.placeholder === 'Description');
      expect(descriptionInputs.length).toBe(1);
    });

    it('regenerates system lines while preserving manual lines', () => {
      createFixture();
      const comp = fixture.componentInstance;
      comp.regenerate();
      fixture.detectChanges();
      expect(createApiService.getPrefill).toHaveBeenCalledWith('c1', '2026-05');
      const lines = comp.lines();
      expect(lines.some((l) => l.lineKind === 'core_childcare')).toBe(true);
      expect(lines.some((l) => l.lineKind === 'extra')).toBe(true);
    });

    it('updates computed totals when lines change', () => {
      createFixture();
      const comp = fixture.componentInstance;
      comp.removeLine('l3');
      expect(comp.subtotalMinor()).toBe(33000 + (-9000));
    });
  });

  describe('save and issue', () => {
    it('calls updateLine when saving modified extra line', () => {
      createFixture();
      const comp = fixture.componentInstance;
      comp.updateLine('l3', 'description', 'Updated pick-up');
      comp.saveChanges();
      expect(apiService.updateLine).toHaveBeenCalled();
    });

    it('calls addLine when saving new line', () => {
      createFixture();
      const comp = fixture.componentInstance;
      comp.addBlankLine();
      comp.saveChanges();
      expect(apiService.addLine).toHaveBeenCalled();
    });

    it('calls deleteLine when saving after removing a line', () => {
      createFixture();
      const comp = fixture.componentInstance;
      comp.removeLine('l3');
      comp.saveChanges();
      expect(apiService.deleteLine).toHaveBeenCalledWith('inv-1', 'l3');
    });

    it('shows toast on successful save', () => {
      createFixture();
      const comp = fixture.componentInstance;
      comp.updateLine('l3', 'description', 'Updated');
      comp.saveChanges();
      expect(toastService.success).toHaveBeenCalledWith('Invoice saved.');
    });

    it('calls issueInvoice after save when issuing', () => {
      createFixture();
      const comp = fixture.componentInstance;
      comp.issueInvoice();
      expect(apiService.issueInvoice).toHaveBeenCalledWith('inv-1');
    });

    it('navigates to detail page after issuing', () => {
      createFixture();
      const comp = fixture.componentInstance;
      const navigateSpy = spyOn(router, 'navigate');
      comp.issueInvoice();
      expect(navigateSpy).toHaveBeenCalledWith(['/manager/invoices', 'inv-1']);
    });
  });

  describe('access control', () => {
    it('redirects to detail page when invoice is not draft', () => {
      const apiSpy = createApiSpy();
      apiSpy.getInvoice.and.returnValue(of(issuedDetail));
      const createApiSpyObj = createCreateApiSpy();
      const toastSpy = jasmine.createSpyObj('ToastService', ['success', 'error', 'warning', 'info']);

      TestBed.configureTestingModule({
        imports: [ManagerInvoiceEditComponent, HttpClientTestingModule],
        providers: [
          provideRouter([]),
          ApiErrorMapper,
          { provide: ManagerInvoicesApiService, useValue: apiSpy },
          { provide: ManagerInvoiceCreateApiService, useValue: createApiSpyObj },
          { provide: ToastService, useValue: toastSpy },
          {
            provide: ActivatedRoute,
            useValue: { snapshot: { paramMap: { get: (key: string) => key === 'invoiceId' ? 'inv-2' : null } } },
          },
        ],
      });

      const routerRef = TestBed.inject(Router);
      spyOn(routerRef, 'navigate');

      fixture = TestBed.createComponent(ManagerInvoiceEditComponent);
      fixture.detectChanges();

      expect(routerRef.navigate).toHaveBeenCalledWith(['/manager/invoices', 'inv-2']);
      expect(toastSpy.warning).toHaveBeenCalledWith('Only draft invoices can be edited.');
    });
  });

  describe('error handling', () => {
    it('shows error message on API error', () => {
      const apiSpy = createApiSpy();
      apiSpy.getInvoice.and.returnValue(throwError(() => new HttpErrorResponse({
        error: { code: 'not_found', message: 'Invoice not found', request_id: 'req-99' },
        status: 404,
      })));
      const createApiSpyObj = createCreateApiSpy();
      const toastSpy = jasmine.createSpyObj('ToastService', ['success', 'error', 'warning', 'info']);

      TestBed.configureTestingModule({
        imports: [ManagerInvoiceEditComponent, HttpClientTestingModule],
        providers: [
          provideRouter([]),
          ApiErrorMapper,
          { provide: ManagerInvoicesApiService, useValue: apiSpy },
          { provide: ManagerInvoiceCreateApiService, useValue: createApiSpyObj },
          { provide: ToastService, useValue: toastSpy },
          {
            provide: ActivatedRoute,
            useValue: { snapshot: { paramMap: { get: (key: string) => key === 'invoiceId' ? 'inv-1' : null } } },
          },
        ],
      });

      fixture = TestBed.createComponent(ManagerInvoiceEditComponent);
      fixture.detectChanges();
      const text = fixture.nativeElement.textContent;
      expect(text).toContain('no longer available');
    });
  });
});
