import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { ManagerInvoiceCreateComponent } from './manager-invoice-create.component';
import { ManagerInvoiceCreateApiService } from '../../data/manager-invoice-create-api.service';
import { StaffApiService } from '../../data/staff-api.service';
import { AuthService } from '../../../../core/services/auth.service';
import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { ToastService } from '../../../../shared/services/toast.service';
import { ChildRecord } from '../../models/children.models';
import { FormInvoiceLine } from '../../models/manager-invoice-create.models';

const mockChild: ChildRecord = {
  id: 'child-1',
  fullName: 'Ben Smith',
  firstName: 'Ben',
  lastName: 'Smith',
  dateOfBirth: '2022-03-15',
  startDate: '2024-01-10',
  endDate: null,
  siteCoreHourlyRateMinor: null,
  notes: null,
  isActive: true,
  enrollmentComplete: true,
  missingRequirements: [],
  photoUrl: null,
  createdAt: '2024-01-10T00:00:00Z',
  updatedAt: '2024-01-10T00:00:00Z',
};

const mockPrefill = {
  lines: [
    {
      lineKind: 'core_childcare',
      description: 'Core Childcare',
      sortOrder: 1,
      quantityMinutes: 4800,
      unitAmountMinor: 6500,
      lineAmountMinor: 52000,
      fundedAllowanceMinutes: 0,
      fundedDeductionMinutes: 0,
      coreBillableMinutes: 4800,
      sessionCount: 20,
    },
    {
      lineKind: 'funded_deduction',
      description: 'Funded Hours Deduction',
      sortOrder: 2,
      quantityMinutes: 0,
      unitAmountMinor: 0,
      lineAmountMinor: -9000,
      fundedAllowanceMinutes: 0,
      fundedDeductionMinutes: 0,
      coreBillableMinutes: 0,
      sessionCount: 0,
    },
  ],
  entitlementStatus: {
    statusLabel: '15h Funded',
    fundingProfileId: 'fp-1',
    fundedAllowanceMinutes: 5760,
  },
  childId: 'child-1',
  childFirstName: 'Ben',
  childMiddleName: null,
  childLastName: 'Smith',
  billingMonth: '2026-07',
  subtotalMinor: 43000,
  fundedDeductionMinor: -9000,
  totalDueMinor: 52000,
  warnings: [],
};

function toFormLine(line: typeof mockPrefill.lines[0], id: string): FormInvoiceLine {
  return {
    ...line,
    id,
    isFundingOffset: line.lineKind === 'funded_deduction',
  };
}

describe('ManagerInvoiceCreateComponent', () => {
  let fixture: ComponentFixture<ManagerInvoiceCreateComponent>;
  let component: ManagerInvoiceCreateComponent;
  let apiService: jasmine.SpyObj<ManagerInvoiceCreateApiService>;
  let staffApiService: jasmine.SpyObj<StaffApiService>;
  let toastService: jasmine.SpyObj<ToastService>;

  beforeEach(async () => {
    apiService = jasmine.createSpyObj('ManagerInvoiceCreateApiService', ['getPrefill', 'createDraft', 'createAndIssue']);
    staffApiService = jasmine.createSpyObj('StaffApiService', ['listChildren', 'getChildContacts']);
    toastService = jasmine.createSpyObj('ToastService', ['success', 'error']);

    const authService = jasmine.createSpyObj('AuthService', [], {
      activeMembership: () => ({ branch_id: 'branch-1' }),
    });

    await TestBed.configureTestingModule({
      imports: [ManagerInvoiceCreateComponent],
      providers: [
        provideRouter([]),
        { provide: ManagerInvoiceCreateApiService, useValue: apiService },
        { provide: StaffApiService, useValue: staffApiService },
        { provide: AuthService, useValue: authService },
        { provide: ToastService, useValue: toastService },
        ApiErrorMapper,
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ManagerInvoiceCreateComponent);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  describe('step initialization', () => {
    it('should start at step child-month', () => {
      expect(component.currentStep()).toBe('child-month');
    });

    it('should have empty completedSteps', () => {
      expect(component.completedSteps().size).toBe(0);
    });

    it('should have activeStepIndex 0', () => {
      expect(component.activeStepIndex()).toBe(0);
    });
  });

  describe('step queries', () => {
    it('stepIsActive returns true for current step', () => {
      expect(component.stepIsActive('child-month')).toBeTrue();
      expect(component.stepIsActive('review-lines')).toBeFalse();
    });

    it('stepIsComplete returns false when no steps completed', () => {
      expect(component.stepIsComplete('child-month')).toBeFalse();
    });

    it('canOpenStep returns true for current step', () => {
      expect(component.canOpenStep('child-month')).toBeTrue();
    });

    it('canOpenStep returns false for forward steps when prior steps incomplete', () => {
      expect(component.canOpenStep('review-lines')).toBeFalse();
      expect(component.canOpenStep('add-extras')).toBeFalse();
      expect(component.canOpenStep('summary-confirm')).toBeFalse();
    });
  });

  describe('validation', () => {
    it('validateStep child-month returns error when no child selected', () => {
      expect(component.validateStep('child-month')).toBe('Select a child.');
    });

    it('validateStep child-month returns error when no billing month', () => {
      component.selectedChild = mockChild;
      expect(component.validateStep('child-month')).toBe('Select a billing month.');
    });

    it('validateStep child-month returns null when valid', () => {
      component.selectedChild = mockChild;
      component.billingMonth.set('2026-07');
      expect(component.validateStep('child-month')).toBeNull();
    });

    it('validateStep review-lines returns error when no lines', () => {
      expect(component.validateStep('review-lines')).toBe('At least one line item is required.');
    });

    it('validateStep review-lines returns null when lines exist', () => {
      component.lines.set([toFormLine(mockPrefill.lines[0], 'auto-1')]);
      expect(component.validateStep('review-lines')).toBeNull();
    });

    it('validateStep add-extras always returns null', () => {
      expect(component.validateStep('add-extras')).toBeNull();
    });

    it('validateStep summary-confirm always returns null', () => {
      expect(component.validateStep('summary-confirm')).toBeNull();
    });
  });

  describe('navigation', () => {
    it('nextStep advances when validation passes', () => {
      component.selectedChild = mockChild;
      component.billingMonth.set('2026-07');
      component.nextStep();
      expect(component.currentStep()).toBe('review-lines');
    });

    it('nextStep does not advance when validation fails', () => {
      component.nextStep();
      expect(component.currentStep()).toBe('child-month');
    });

    it('nextStep marks current step as complete on advance', () => {
      component.selectedChild = mockChild;
      component.billingMonth.set('2026-07');
      component.nextStep();
      expect(component.stepIsComplete('child-month')).toBeTrue();
    });

    it('prevStep goes back', () => {
      component.selectedChild = mockChild;
      component.billingMonth.set('2026-07');
      component.nextStep();
      expect(component.currentStep()).toBe('review-lines');
      component.prevStep();
      expect(component.currentStep()).toBe('child-month');
    });

    it('prevStep does nothing on first step', () => {
      component.prevStep();
      expect(component.currentStep()).toBe('child-month');
    });

    it('goToStep navigates to completed step', () => {
      component.selectedChild = mockChild;
      component.billingMonth.set('2026-07');
      component.nextStep(); // -> review-lines, child-month complete
      component.lines.set([toFormLine(mockPrefill.lines[0], 'auto-1')]);
      component.nextStep(); // -> add-extras, review-lines complete
      component.goToStep('child-month');
      expect(component.currentStep()).toBe('child-month');
    });

    it('goToStep blocks forward navigation to locked steps', () => {
      component.goToStep('summary-confirm');
      expect(component.currentStep()).toBe('child-month');
    });
  });

  describe('submit guards', () => {
    it('canSaveDraft returns false when no child', () => {
      expect(component.canSaveDraft()).toBeFalse();
    });

    it('canSaveDraft returns false when no lines', () => {
      component.selectedChild = mockChild;
      component.billingMonth.set('2026-07');
      expect(component.canSaveDraft()).toBeFalse();
    });

    it('canSaveDraft returns true when child, month, and lines exist', () => {
      component.selectedChild = mockChild;
      component.billingMonth.set('2026-07');
      component.lines.set([toFormLine(mockPrefill.lines[0], 'auto-1')]);
      expect(component.canSaveDraft()).toBeTrue();
    });

    it('canIssue returns false when not on summary-confirm step', () => {
      component.selectedChild = mockChild;
      component.billingMonth.set('2026-07');
      component.lines.set([toFormLine(mockPrefill.lines[0], 'auto-1')]);
      expect(component.canIssue()).toBeFalse();
    });

    it('canIssue returns true on summary-confirm step with all data', () => {
      component.selectedChild = mockChild;
      component.billingMonth.set('2026-07');
      component.lines.set([toFormLine(mockPrefill.lines[0], 'auto-1')]);
      // Navigate to summary-confirm
      component.nextStep(); // -> review-lines
      component.nextStep(); // -> add-extras
      component.nextStep(); // -> summary-confirm
      expect(component.currentStep()).toBe('summary-confirm');
      expect(component.canIssue()).toBeTrue();
    });
  });

  describe('line management', () => {
    it('addBlankLine adds an extra line', () => {
      component.addBlankLine();
      expect(component.lines().length).toBe(1);
      expect(component.lines()[0].lineKind).toBe('extra');
    });

    it('removeLine removes a line by id', () => {
      component.addBlankLine();
      const lineId = component.lines()[0].id;
      component.removeLine(lineId);
      expect(component.lines().length).toBe(0);
    });

    it('updateLine recalculates lineAmountMinor for extras', () => {
      component.addBlankLine();
      const line = component.lines()[0];
      component.updateLine(line.id, 'quantityMinutes', 3);
      component.updateLine(line.id, 'unitAmountMinor', 1000);
      expect(component.lines()[0].lineAmountMinor).toBe(3000);
    });

    it('addPresetLine adds a preset extra line', () => {
      component.addPresetLine('Full Day', 6500, 1);
      expect(component.lines().length).toBe(1);
      expect(component.lines()[0].description).toBe('Full Day');
      expect(component.lines()[0].lineAmountMinor).toBe(6500);
    });

    it('autoGeneratedLines filters out extras', () => {
      component.lines.set([
        toFormLine(mockPrefill.lines[0], 'auto-1'),
        { id: 'extra-1', lineKind: 'extra', description: 'Test', sortOrder: 2, quantityMinutes: 1, unitAmountMinor: 1000, lineAmountMinor: 1000, fundedAllowanceMinutes: 0, fundedDeductionMinutes: 0, coreBillableMinutes: 0, sessionCount: 0, isFundingOffset: false },
      ]);
      expect(component.autoGeneratedLines().length).toBe(1);
      expect(component.autoGeneratedLines()[0].id).toBe('auto-1');
    });

    it('extraLines returns only extras', () => {
      component.lines.set([
        toFormLine(mockPrefill.lines[0], 'auto-1'),
        { id: 'extra-1', lineKind: 'extra', description: 'Test', sortOrder: 2, quantityMinutes: 1, unitAmountMinor: 1000, lineAmountMinor: 1000, fundedAllowanceMinutes: 0, fundedDeductionMinutes: 0, coreBillableMinutes: 0, sessionCount: 0, isFundingOffset: false },
      ]);
      expect(component.extraLines().length).toBe(1);
      expect(component.extraLines()[0].id).toBe('extra-1');
    });
  });

  describe('state persistence across steps', () => {
    it('lines signal retains data when navigating forward and back', () => {
      component.selectedChild = mockChild;
      component.billingMonth.set('2026-07');
      component.nextStep(); // -> review-lines

      component.lines.set([toFormLine(mockPrefill.lines[0], 'auto-1')]);
      const linesBefore = component.lines();

      component.nextStep(); // -> add-extras
      expect(component.lines()).toEqual(linesBefore);

      component.prevStep(); // -> review-lines
      expect(component.lines()).toEqual(linesBefore);
    });

    it('internalNotes and paymentTerms persist across steps', () => {
      component.selectedChild = mockChild;
      component.billingMonth.set('2026-07');
      component.nextStep();
      component.lines.set([toFormLine(mockPrefill.lines[0], 'auto-1')]);
      component.nextStep();
      component.nextStep(); // -> summary-confirm

      component.internalNotes = 'Test note';
      component.paymentTerms = 'Custom terms';
      component.prevStep();
      component.nextStep(); // -> summary-confirm again

      expect(component.internalNotes).toBe('Test note');
      expect(component.paymentTerms).toBe('Custom terms');
    });
  });

  describe('computed totals', () => {
    it('subtotalMinor sums all line amounts', () => {
      component.lines.set([
        toFormLine(mockPrefill.lines[0], 'auto-1'),
        toFormLine(mockPrefill.lines[1], 'auto-2'),
      ]);
      expect(component.subtotalMinor()).toBe(43000);
    });

    it('fundedDeductionMinor sums only funding offset lines', () => {
      component.lines.set([
        toFormLine(mockPrefill.lines[0], 'auto-1'),
        toFormLine(mockPrefill.lines[1], 'auto-2'),
      ]);
      expect(component.fundedDeductionMinor()).toBe(-9000);
    });

    it('totalDueMinor is subtotal minus deduction', () => {
      component.lines.set([
        toFormLine(mockPrefill.lines[0], 'auto-1'),
        toFormLine(mockPrefill.lines[1], 'auto-2'),
      ]);
      expect(component.totalDueMinor()).toBe(61000);
    });
  });
});
