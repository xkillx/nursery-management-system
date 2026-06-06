import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ChildFormComponent } from './child-form.component';
import { ChildRecord } from '../../models/children.models';

describe('ChildFormComponent', () => {
  let fixture: ComponentFixture<ChildFormComponent>;
  let component: ChildFormComponent;

  const childRecord: ChildRecord = {
    id: 'child-1',
    fullName: 'Ada Lovelace',
    dateOfBirth: '2022-01-15',
    startDate: '2024-09-01',
    endDate: null,
    coreHourlyRateMinor: 750,
    notes: null,
    isActive: true,
    leftAt: null,
    leftReasonCode: null,
    leftReasonNote: null,
    enrollmentComplete: true,
    missingRequirements: [],
    createdAt: '2024-08-01T00:00:00Z',
    updatedAt: '2024-08-01T00:00:00Z',
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ChildFormComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(ChildFormComponent);
    component = fixture.componentInstance;
  });

  it('creates with empty form defaulting rate to 0', () => {
    expect(component.form.core_hourly_rate_gbp).toBe(0);
  });

  it('populates form with GBP rate from minor units in edit mode', () => {
    component.selectedChild = childRecord;
    component.ngOnChanges({ selectedChild: { currentValue: childRecord, previousValue: null, firstChange: true, isFirstChange: () => true } });

    expect(component.form.core_hourly_rate_gbp).toBe(7.5);
  });

  it('emits payload with minor units when submitting GBP rate', () => {
    const savedSpy = spyOn(component.saved, 'emit');

    component.form.full_name = 'Ada Lovelace';
    component.form.date_of_birth = '2022-01-15';
    component.form.start_date = '2024-09-01';
    component.form.core_hourly_rate_gbp = 7.5;
    component.form.end_date = '';
    component.form.notes = '';

    component.submit();

    expect(savedSpy).toHaveBeenCalledWith({
      full_name: 'Ada Lovelace',
      date_of_birth: '2022-01-15',
      start_date: '2024-09-01',
      core_hourly_rate_minor: 750,
      end_date: '',
      notes: '',
    });
  });

  it('rounds fractional pounds to nearest penny on submit', () => {
    const savedSpy = spyOn(component.saved, 'emit');

    component.form.full_name = 'Test';
    component.form.date_of_birth = '2022-01-01';
    component.form.start_date = '2024-01-01';
    component.form.core_hourly_rate_gbp = 7.555;
    component.form.end_date = '';
    component.form.notes = '';

    component.submit();

    expect(savedSpy).toHaveBeenCalledWith(
      jasmine.objectContaining({ core_hourly_rate_minor: 756 }),
    );
  });

  it('handles string rate input from form control', () => {
    const savedSpy = spyOn(component.saved, 'emit');

    component.form.full_name = 'Test';
    component.form.date_of_birth = '2022-01-01';
    component.form.start_date = '2024-01-01';
    (component.form as any).core_hourly_rate_gbp = '7.50';
    component.form.end_date = '';
    component.form.notes = '';

    component.submit();

    expect(savedSpy).toHaveBeenCalledWith(
      jasmine.objectContaining({ core_hourly_rate_minor: 750 }),
    );
  });

  it('resets form when selectedChild changes to null', () => {
    component.selectedChild = childRecord;
    component.ngOnChanges({ selectedChild: { currentValue: childRecord, previousValue: null, firstChange: true, isFirstChange: () => true } });
    expect(component.form.core_hourly_rate_gbp).toBe(7.5);

    component.selectedChild = null;
    component.ngOnChanges({ selectedChild: { currentValue: null, previousValue: childRecord, firstChange: false, isFirstChange: () => false } });
    expect(component.form.core_hourly_rate_gbp).toBe(0);
    expect(component.form.full_name).toBe('');
  });
});
