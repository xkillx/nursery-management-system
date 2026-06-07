import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute } from '@angular/router';
import { of, throwError } from 'rxjs';

import { StaffApiService } from '../../data/staff-api.service';
import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { ManagerChildDetailComponent } from './manager-child-detail.component';

describe('ManagerChildDetailComponent', () => {
  let fixture: ComponentFixture<ManagerChildDetailComponent>;
  let component: ManagerChildDetailComponent;
  let staffApiMock: jasmine.SpyObj<StaffApiService>;

  const mockChild = {
    id: 'child-1',
    fullName: 'Emma Thompson',
    dateOfBirth: '2022-03-15',
    startDate: '2023-01-10',
    endDate: null,
    coreHourlyRateMinor: 850,
    notes: null,
    isActive: true,
    leftAt: null,
    leftReasonCode: null,
    leftReasonNote: null,
    enrollmentComplete: false,
    missingRequirements: ['guardian_link'],
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
  };

  const mockLinks = [
    {
      id: 'link-1',
      guardianId: 'guardian-1',
      childId: 'child-1',
      guardian: {
        id: 'guardian-1',
        fullName: 'Sarah Thompson',
        email: 'sarah@example.com',
        phone: '+44 7700 900001',
        isActive: true,
      },
      createdAt: '2026-06-07T10:00:00Z',
      updatedAt: '2026-06-07T10:00:00Z',
    },
  ];

  const mockGuardians = [
    {
      id: 'guardian-1',
      fullName: 'Sarah Thompson',
      email: 'sarah@example.com',
      phone: '+44 7700 900001',
      notes: null,
      isActive: true,
      deactivatedAt: null,
      deactivationReasonCode: null,
      deactivationReasonNote: null,
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-01T00:00:00Z',
    },
    {
      id: 'guardian-2',
      fullName: 'John Smith',
      email: null,
      phone: null,
      notes: null,
      isActive: true,
      deactivatedAt: null,
      deactivationReasonCode: null,
      deactivationReasonNote: null,
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-01T00:00:00Z',
    },
  ];

  beforeEach(async () => {
    staffApiMock = jasmine.createSpyObj('StaffApiService', [
      'getChild', 'listChildGuardianLinks', 'listGuardians', 'updateChild', 'createGuardianChildLink',
    ]);

    staffApiMock.getChild.and.returnValue(of(mockChild));
    staffApiMock.listChildGuardianLinks.and.returnValue(of(mockLinks));
    staffApiMock.listGuardians.and.returnValue(of(mockGuardians));
    staffApiMock.updateChild.and.returnValue(of(mockChild));
    staffApiMock.createGuardianChildLink.and.returnValue(of({} as any));

    await TestBed.configureTestingModule({
      imports: [ManagerChildDetailComponent],
      providers: [
        { provide: StaffApiService, useValue: staffApiMock },
        { provide: ActivatedRoute, useValue: { snapshot: { paramMap: { get: (key: string) => key === 'childId' ? 'child-1' : null } } } },
        ApiErrorMapper,
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ManagerChildDetailComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('creates', () => {
    expect(component).toBeTruthy();
  });

  it('loads child detail on init', () => {
    expect(staffApiMock.getChild).toHaveBeenCalledWith('child-1');
    expect(component.child?.fullName).toBe('Emma Thompson');
  });

  it('loads linked guardians', () => {
    expect(staffApiMock.listChildGuardianLinks).toHaveBeenCalledWith('child-1');
    expect(component.linkedGuardians.length).toBe(1);
    expect(component.linkedGuardians[0].guardian.fullName).toBe('Sarah Thompson');
  });

  it('loads all active guardians for selector', () => {
    expect(staffApiMock.listGuardians).toHaveBeenCalledWith({ status: 'active', limit: 200, offset: 0 });
  });

  it('availableGuardians excludes already-linked guardians', () => {
    expect(component.availableGuardians.length).toBe(1);
    expect(component.availableGuardians[0].id).toBe('guardian-2');
  });

  it('shows missing enrollment requirements', () => {
    expect(component.child?.missingRequirements).toContain('guardian_link');
  });

  it('shows core hourly rate formatted as GBP per hour', () => {
    const rate = component.formatRate(component.child!.coreHourlyRateMinor);
    expect(rate).toBe('£8.50/hr');
  });

  it('opens edit form', () => {
    component.onEditChild();
    expect(component.showEditForm).toBeTrue();
  });

  it('closes edit form', () => {
    component.onEditChild();
    component.closeEditForm();
    expect(component.showEditForm).toBeFalse();
  });

  it('saves child and reloads', () => {
    component.childId = 'child-1';
    component.onEditChild();
    component.saveChild({
      full_name: 'Emma Thompson',
      date_of_birth: '2022-03-15',
      start_date: '2023-01-10',
      core_hourly_rate_minor: 900,
    });

    expect(staffApiMock.updateChild).toHaveBeenCalledWith('child-1', jasmine.objectContaining({
      core_hourly_rate_minor: 900,
    }));
  });

  it('links guardian and reloads', () => {
    component.childId = 'child-1';
    component.selectedGuardianId = 'guardian-2';
    component.linkGuardian();

    expect(staffApiMock.createGuardianChildLink).toHaveBeenCalledWith({
      guardian_id: 'guardian-2',
      child_id: 'child-1',
    });
  });

  it('handles child load error', () => {
    staffApiMock.getChild.and.returnValue(throwError(() => new Error('not found')));
    component.childId = '';
    component.ngOnInit();

    expect(component.errorMessage).toBeTruthy();
  });

  it('funded-hours section is read-only placeholder', () => {
    expect(component.currentMonthLabel).toMatch(/^\d{4}-\d{2}$/);
  });
});
