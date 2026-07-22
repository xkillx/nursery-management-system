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
  middleName: null,
  lastName: 'Smith',
  dateOfBirth: '2022-03-15',
  startDate: '2024-01-10',
  endDate: null,
  siteCoreHourlyRateMinor: null,
  notes: null,
  isActive: true,
  primaryRoomId: null,
  hasCurrentRoom: false,
  hasBookingPattern: false,
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

  describe('initial state', () => {
    it('should have no selected child', () => {
      expect(component.selectedChild).toBeNull();
    });

    it('should have empty lines', () => {
      expect(component.lines()).toEqual([]);
    });

    it('should have default payment terms', () => {
      expect(component.paymentTerms).toContain('7 days');
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

    it('canIssue returns false when no data', () => {
      expect(component.canIssue()).toBeFalse();
    });

    it('canIssue returns true when child, month, and lines exist', () => {
      component.selectedChild = mockChild;
      component.billingMonth.set('2026-07');
      component.lines.set([toFormLine(mockPrefill.lines[0], 'auto-1')]);
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

  describe('clearChild', () => {
    it('resets selectedChild and related state', () => {
      component.selectedChild = mockChild;
      component.billingMonth.set('2026-07');
      component.lines.set([toFormLine(mockPrefill.lines[0], 'auto-1')]);
      component.entitlementLabel = 'test';
      component.hasFundingProfile = true;

      component.clearChild();

      expect(component.selectedChild).toBeNull();
      expect(component.childSearchTerm).toBe('');
      expect(component.lines()).toEqual([]);
      expect(component.entitlementLabel).toBe('');
      expect(component.hasFundingProfile).toBeFalse();
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
      expect(component.fundedDeductionMinor()).toBe(9000);
    });

    it('totalDueMinor is subtotal minus deduction', () => {
      component.lines.set([
        toFormLine(mockPrefill.lines[0], 'auto-1'),
        toFormLine(mockPrefill.lines[1], 'auto-2'),
      ]);
      expect(component.totalDueMinor()).toBe(34000);
    });
  });

  describe('calculateAgeGroup', () => {
    it('returns correct age group for different dates', () => {
      const now = new Date();
      const under1 = new Date(now.getFullYear(), now.getMonth() - 3, 1).toISOString().split('T')[0];
      expect(component.calculateAgeGroup(under1)).toBe('Under 1 Year');

      const age1 = new Date(now.getFullYear() - 1, now.getMonth(), 15).toISOString().split('T')[0];
      expect(component.calculateAgeGroup(age1)).toBe('1-2 Years');

      const age2 = new Date(now.getFullYear() - 2, now.getMonth(), 15).toISOString().split('T')[0];
      expect(component.calculateAgeGroup(age2)).toBe('2-3 Years');

      const age4 = new Date(now.getFullYear() - 4, now.getMonth(), 15).toISOString().split('T')[0];
      expect(component.calculateAgeGroup(age4)).toBe('3-5 Years');
    });

    it('returns Unknown for empty string', () => {
      expect(component.calculateAgeGroup('')).toBe('Unknown');
    });
  });

  describe('getRoomNameByAgeGroup', () => {
    it('returns correct room names', () => {
      expect(component.getRoomNameByAgeGroup('Under 1 Year')).toBe('Babies Room');
      expect(component.getRoomNameByAgeGroup('1-2 Years')).toBe('Minnows Room');
      expect(component.getRoomNameByAgeGroup('2-3 Years')).toBe('Squirrels Room');
      expect(component.getRoomNameByAgeGroup('3-5 Years')).toBe('Badgers Room');
      expect(component.getRoomNameByAgeGroup('Unknown')).toBe('Main Hall');
    });
  });
});
