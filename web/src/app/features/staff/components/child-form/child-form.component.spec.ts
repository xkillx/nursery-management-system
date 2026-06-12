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
    coreHourlyRateMinor: null,
    siteCoreHourlyRateMinor: 750,
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

  it('creates with empty form', () => {
    expect(component.form.full_name).toBe('');
  });

  it('populates form with child data in edit mode', () => {
    component.selectedChild = childRecord;
    component.ngOnChanges({ selectedChild: { currentValue: childRecord, previousValue: null, firstChange: true, isFirstChange: () => true } });

    expect(component.form.full_name).toBe('Ada Lovelace');
    expect(component.form.date_of_birth).toBe('2022-01-15');
    expect(component.form.start_date).toBe('2024-09-01');
  });

  it('emits payload without core_hourly_rate_minor', () => {
    const savedSpy = spyOn(component.saved, 'emit');

    component.form.full_name = 'Ada Lovelace';
    component.form.date_of_birth = '2022-01-15';
    component.form.start_date = '2024-09-01';
    component.form.end_date = '';
    component.form.notes = '';

    component.submit();

    expect(savedSpy).toHaveBeenCalledWith({
      full_name: 'Ada Lovelace',
      date_of_birth: '2022-01-15',
      start_date: '2024-09-01',
      end_date: '',
      notes: '',
    });
    expect(savedSpy.calls.mostRecent().args[0]).not.toEqual(jasmine.objectContaining({ core_hourly_rate_minor: jasmine.anything() as any }));
  });

  it('resets form when selectedChild changes to null', () => {
    component.selectedChild = childRecord;
    component.ngOnChanges({ selectedChild: { currentValue: childRecord, previousValue: null, firstChange: true, isFirstChange: () => true } });
    expect(component.form.full_name).toBe('Ada Lovelace');

    component.selectedChild = null;
    component.ngOnChanges({ selectedChild: { currentValue: null, previousValue: childRecord, firstChange: false, isFirstChange: () => false } });
    expect(component.form.full_name).toBe('');
  });
});
