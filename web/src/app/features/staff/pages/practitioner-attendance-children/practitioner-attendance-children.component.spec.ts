import { HttpErrorResponse, provideHttpClient } from '@angular/common/http';
import { ComponentFixture, fakeAsync, flush, TestBed, tick } from '@angular/core/testing';
import { Observable, of, throwError } from 'rxjs';

import { MappedApiError } from '../../../../core/models/api-error.models';
import { StaffApiService } from '../../data/staff-api.service';
import { AttendanceChildRecord } from '../../models/attendance-child.models';
import { PractitionerAttendanceChildrenComponent } from './practitioner-attendance-children.component';

describe('PractitionerAttendanceChildrenComponent', () => {
  let fixture: ComponentFixture<PractitionerAttendanceChildrenComponent>;
  let component: PractitionerAttendanceChildrenComponent;
  let staffApiSpy: jasmine.SpyObj<StaffApiService>;

  const mockChildren: AttendanceChildRecord[] = [
    {
      id: 'child-1',
      fullName: 'Ada Lovelace',
      enrollmentComplete: true,
      attendanceState: 'not_checked_in',
      openSessionId: null,
      checkedInAt: null,
      hasIncompleteSession: false,
      absenceMarkerId: null,
      absenceMarkedAt: null,
    },
    {
      id: 'child-2',
      fullName: 'Grace Hopper',
      enrollmentComplete: false,
      attendanceState: 'not_checked_in',
      openSessionId: null,
      checkedInAt: null,
      hasIncompleteSession: false,
      absenceMarkerId: null,
      absenceMarkedAt: null,
    },
    {
      id: 'child-3',
      fullName: 'Katherine Johnson',
      enrollmentComplete: true,
      attendanceState: 'checked_in',
      openSessionId: 'session-1',
      checkedInAt: '2026-05-24T07:42:00Z',
      hasIncompleteSession: true,
      absenceMarkerId: null,
      absenceMarkedAt: null,
    },
  ];

  const mockCheckedInIncomplete: AttendanceChildRecord = {
    id: 'child-4',
    fullName: 'Incomplete But In',
    enrollmentComplete: false,
    attendanceState: 'checked_in',
    openSessionId: 'session-2',
    checkedInAt: '2026-05-24T08:00:00Z',
    hasIncompleteSession: true,
    absenceMarkerId: null,
    absenceMarkedAt: null,
  };

  beforeEach(async () => {
    const spy = jasmine.createSpyObj<StaffApiService>('StaffApiService', [
      'listAttendanceChildren',
      'checkInChild',
      'checkOutChild',
    ]);

    await TestBed.configureTestingModule({
      imports: [PractitionerAttendanceChildrenComponent],
      providers: [
        provideHttpClient(),
        { provide: StaffApiService, useValue: spy },
      ],
    }).compileComponents();

    staffApiSpy = TestBed.inject(StaffApiService) as jasmine.SpyObj<StaffApiService>;
    fixture = TestBed.createComponent(PractitionerAttendanceChildrenComponent);
    component = fixture.componentInstance;
  });

  function setChildrenAndDetectChanges(children: AttendanceChildRecord[]): void {
    staffApiSpy.listAttendanceChildren.and.returnValue(of(children));
    fixture.detectChanges();
  }

  it('creates the component', () => {
    staffApiSpy.listAttendanceChildren.and.returnValue(of([]));
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });

  it('loads children on init without showing child IDs', () => {
    setChildrenAndDetectChanges(mockChildren);

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Ada Lovelace');
    expect(compiled.textContent).toContain('Grace Hopper');
    expect(compiled.textContent).toContain('Katherine Johnson');
    expect(compiled.textContent).not.toContain('child-1');
  });

  it('filters by child name search', () => {
    setChildrenAndDetectChanges(mockChildren);

    component.onSearchChange('Ada');
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Ada Lovelace');
    expect(compiled.textContent).not.toContain('Grace Hopper');
  });

  it('shows status filter counts', () => {
    setChildrenAndDetectChanges(mockChildren);
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('All (3)');
    expect(compiled.textContent).toContain('Not in (2)');
    expect(compiled.textContent).toContain('Checked in (1)');
  });

  it('filters to checked-in children', () => {
    setChildrenAndDetectChanges(mockChildren);

    component.setStatusFilter('checked_in');
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Katherine Johnson');
    expect(compiled.textContent).not.toContain('Ada Lovelace');
  });

  it('filters to not-checked-in children', () => {
    setChildrenAndDetectChanges(mockChildren);

    component.setStatusFilter('not_checked_in');
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Ada Lovelace');
    expect(compiled.textContent).not.toContain('Katherine Johnson');
  });

  it('disables check-in and shows enrollment incomplete for incomplete child', () => {
    setChildrenAndDetectChanges(mockChildren);
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Enrollment incomplete');

    const checkInBtn = compiled.querySelector('[data-testid="attendance-action-child-2"]') as HTMLButtonElement;
    expect(checkInBtn).toBeTruthy();
    expect(checkInBtn.disabled).toBeTrue();
  });

  it('allows check-out for checked-in incomplete-enrollment child', () => {
    setChildrenAndDetectChanges([...mockChildren, mockCheckedInIncomplete]);
    fixture.detectChanges();

    expect(component.canCheckOut(mockCheckedInIncomplete)).toBeTrue();
  });

  it('calls checkInChild then reloads list on successful check-in', () => {
    setChildrenAndDetectChanges(mockChildren);

    staffApiSpy.checkInChild.and.returnValue(of({ id: 'session-new' } as any));
    staffApiSpy.listAttendanceChildren.and.returnValue(of(mockChildren));

    component.checkIn(mockChildren[0]);
    fixture.detectChanges();

    expect(staffApiSpy.checkInChild).toHaveBeenCalledWith('child-1');
    expect(staffApiSpy.listAttendanceChildren).toHaveBeenCalledTimes(2);
  });

  it('calls checkOutChild then reloads list on successful check-out', () => {
    setChildrenAndDetectChanges(mockChildren);

    staffApiSpy.checkOutChild.and.returnValue(of({ id: 'session-1' } as any));
    staffApiSpy.listAttendanceChildren.and.returnValue(of(mockChildren));

    component.checkOut(mockChildren[2]);
    fixture.detectChanges();

    expect(staffApiSpy.checkOutChild).toHaveBeenCalledWith('child-3');
    expect(staffApiSpy.listAttendanceChildren).toHaveBeenCalledTimes(2);
  });

  it('shows row error on check-in 409 attendance_session_already_open and reloads list', () => {
    setChildrenAndDetectChanges(mockChildren);

    const conflictError = new HttpErrorResponse({
      status: 409,
      error: {
        code: 'attendance_session_already_open',
        message: 'An open attendance session already exists for this child.',
        request_id: 'req-1',
      },
    });

    staffApiSpy.checkInChild.and.returnValue(throwError(() => conflictError));
    staffApiSpy.listAttendanceChildren.and.returnValue(of(mockChildren));

    component.checkIn(mockChildren[0]);
    fixture.detectChanges();

    expect(component.rowErrors['child-1']).toContain('An open attendance session already exists for this child.');
    expect(component.rowErrors['child-1']).toContain('Request: req-1');
    expect(staffApiSpy.listAttendanceChildren).toHaveBeenCalledTimes(2);
  });

  it('shows row error on check-out 409 attendance_session_not_open and reloads list', () => {
    setChildrenAndDetectChanges(mockChildren);

    const conflictError = new HttpErrorResponse({
      status: 409,
      error: {
        code: 'attendance_session_not_open',
        message: 'No open attendance session found for this child.',
        request_id: 'req-2',
      },
    });

    staffApiSpy.checkOutChild.and.returnValue(throwError(() => conflictError));
    staffApiSpy.listAttendanceChildren.and.returnValue(of(mockChildren));

    component.checkOut(mockChildren[2]);
    fixture.detectChanges();

    expect(component.rowErrors['child-3']).toContain('No open attendance session found for this child.');
    expect(component.rowErrors['child-3']).toContain('Request: req-2');
    expect(staffApiSpy.listAttendanceChildren).toHaveBeenCalledTimes(2);
  });

  it('shows row error on check-in 409 child_enrollment_incomplete and reloads list', () => {
    setChildrenAndDetectChanges(mockChildren);

    const conflictError = new HttpErrorResponse({
      status: 409,
      error: {
        code: 'child_enrollment_incomplete',
        message: 'Child enrollment is not complete.',
        request_id: 'req-3',
      },
    });

    staffApiSpy.checkInChild.and.returnValue(throwError(() => conflictError));
    staffApiSpy.listAttendanceChildren.and.returnValue(of(mockChildren));

    component.checkIn(mockChildren[0]);
    fixture.detectChanges();

    expect(component.rowErrors['child-1']).toContain('Child enrollment is not complete.');
    expect(component.rowErrors['child-1']).toContain('Request: req-3');
    expect(staffApiSpy.listAttendanceChildren).toHaveBeenCalledTimes(2);
  });

  it('displays global error when list load fails', () => {
    const loadError = new HttpErrorResponse({
      status: 500,
      statusText: 'Internal Server Error',
    });

    staffApiSpy.listAttendanceChildren.and.returnValue(throwError(() => loadError));
    fixture.detectChanges();

    expect(component.errorMessage).toBeTruthy();
  });

  it('shows empty state when no children from API', () => {
    setChildrenAndDetectChanges([]);

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('No children available for today\'s attendance');
  });

  it('shows filtered empty state when search hides all', () => {
    setChildrenAndDetectChanges(mockChildren);

    component.onSearchChange('nonexistent');
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('No children match the current filters');
  });

  it('refresh button triggers loadChildren', () => {
    setChildrenAndDetectChanges(mockChildren);
    const initialCallCount = staffApiSpy.listAttendanceChildren.calls.count();

    const buttons: HTMLButtonElement[] = Array.from(fixture.nativeElement.querySelectorAll('button'));
    const refreshBtn = buttons.find((b) => b.textContent?.includes('Refresh'));
    expect(refreshBtn).toBeTruthy();
    staffApiSpy.listAttendanceChildren.and.returnValue(of(mockChildren));
    refreshBtn!.click();
    fixture.detectChanges();

    expect(staffApiSpy.listAttendanceChildren.calls.count()).toBe(initialCallCount + 1);
  });

  it('disables action and shows pending text during check-in', () => {
    setChildrenAndDetectChanges(mockChildren);

    staffApiSpy.checkInChild.and.returnValue(new Observable());

    component.checkIn(mockChildren[0]);
    fixture.detectChanges();

    const actionBtn = fixture.nativeElement.querySelector('[data-testid="attendance-action-child-1"]') as HTMLButtonElement;
    expect(actionBtn.disabled).toBeTrue();
    expect(actionBtn.textContent).toContain('Checking in');
  });

  it('shows checked-in time for checked-in children', () => {
    setChildrenAndDetectChanges(mockChildren);

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Checked in at');
  });

  it('shows Not in badge and enabled Check in action for eligible not-in child', () => {
    setChildrenAndDetectChanges(mockChildren);

    const row = fixture.nativeElement.querySelector('[data-testid="attendance-row-child-1"]');
    expect(row.textContent).toContain('Not in');

    const actionBtn = fixture.nativeElement.querySelector('[data-testid="attendance-action-child-1"]') as HTMLButtonElement;
    expect(actionBtn).toBeTruthy();
    expect(actionBtn.textContent).toContain('Check in');
    expect(actionBtn.disabled).toBeFalse();
  });

  it('shows Check out action for checked-in incomplete-enrollment child in DOM', () => {
    setChildrenAndDetectChanges([...mockChildren, mockCheckedInIncomplete]);

    const actionBtn = fixture.nativeElement.querySelector('[data-testid="attendance-action-child-4"]') as HTMLButtonElement;
    expect(actionBtn).toBeTruthy();
    expect(actionBtn.textContent).toContain('Check out');
    expect(actionBtn.disabled).toBeFalse();
  });

  it('does not render guardian, billing, or child ID fields', () => {
    const childrenWithExtraFields: AttendanceChildRecord[] = [
      {
        id: 'child-1',
        fullName: 'Ada Lovelace',
        enrollmentComplete: true,
        attendanceState: 'not_checked_in',
        openSessionId: null,
        checkedInAt: null,
        hasIncompleteSession: false,
        absenceMarkerId: null,
        absenceMarkedAt: null,
      } as any,
    ];
    (childrenWithExtraFields[0] as any).guardianEmail = 'secret@example.com';
    (childrenWithExtraFields[0] as any).guardianPhone = '07123456789';
    (childrenWithExtraFields[0] as any).guardianName = 'Secret Guardian';
    (childrenWithExtraFields[0] as any).coreHourlyRateMinor = 1500;
    (childrenWithExtraFields[0] as any).fundingValue = 10000;

    setChildrenAndDetectChanges(childrenWithExtraFields);

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).not.toContain('secret@example.com');
    expect(compiled.textContent).not.toContain('07123456789');
    expect(compiled.textContent).not.toContain('Secret Guardian');
    expect(compiled.textContent).not.toContain('child-1');
  });

  it('shows row error near the affected child row', () => {
    setChildrenAndDetectChanges(mockChildren);

    const conflictError = new HttpErrorResponse({
      status: 409,
      error: {
        code: 'attendance_session_already_open',
        message: 'An open attendance session already exists for this child.',
        request_id: 'req-row',
      },
    });

    staffApiSpy.checkInChild.and.returnValue(throwError(() => conflictError));
    staffApiSpy.listAttendanceChildren.and.returnValue(of(mockChildren));

    component.checkIn(mockChildren[0]);
    fixture.detectChanges();

    const row = fixture.nativeElement.querySelector('[data-testid="attendance-row-child-1"]');
    const rowError = fixture.nativeElement.querySelector('[data-testid="attendance-row-error-child-1"]');

    expect(rowError).toBeTruthy();
    expect(row.textContent).toContain('Ada Lovelace');
    expect(row.parentElement).toBe(rowError.parentElement);
  });

  it('shows incomplete session warning for not-checked-in child with incomplete session', () => {
    const childWithIncompleteSession: AttendanceChildRecord = {
      id: 'child-5',
      fullName: 'Session Issue',
      enrollmentComplete: true,
      attendanceState: 'not_checked_in',
      openSessionId: null,
      checkedInAt: null,
      hasIncompleteSession: true,
      absenceMarkerId: null,
      absenceMarkedAt: null,
    };
    setChildrenAndDetectChanges([childWithIncompleteSession]);

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Incomplete session needs manager correction');
  });

  describe('absent child handling', () => {
    const absentChild: AttendanceChildRecord = {
      id: 'child-absent',
      fullName: 'Margaret Hamilton',
      enrollmentComplete: true,
      attendanceState: 'absent',
      openSessionId: null,
      checkedInAt: null,
      hasIncompleteSession: false,
      absenceMarkerId: 'marker-1',
      absenceMarkedAt: '2026-06-08T08:00:00Z',
    };

    it('shows Absent badge and Marked absent today for absent child', () => {
      setChildrenAndDetectChanges([absentChild]);

      const compiled = fixture.nativeElement as HTMLElement;
      expect(compiled.textContent).toContain('Margaret Hamilton');
      expect(compiled.textContent).toContain('Absent');
      expect(compiled.textContent).toContain('Marked absent today');
    });

    it('does not render check-in action for absent child', () => {
      setChildrenAndDetectChanges([absentChild]);

      const actionBtn = fixture.nativeElement.querySelector('[data-testid="attendance-action-child-absent"]');
      expect(actionBtn).toBeNull();
    });

    it('checkIn does not call API for absent child', () => {
      setChildrenAndDetectChanges([absentChild]);

      component.checkIn(absentChild);
      fixture.detectChanges();

      expect(staffApiSpy.checkInChild).not.toHaveBeenCalled();
    });

    it('canCheckIn returns false for absent child', () => {
      expect(component.canCheckIn(absentChild)).toBeFalse();
    });

    it('does not show enrollment incomplete for absent child', () => {
      const absentIncomplete = { ...absentChild, enrollmentComplete: false };
      setChildrenAndDetectChanges([absentIncomplete]);

      const compiled = fixture.nativeElement as HTMLElement;
      expect(compiled.textContent).not.toContain('Enrollment incomplete');
    });

    it('does not render absence marker ID or timestamp in DOM', () => {
      setChildrenAndDetectChanges([absentChild]);

      const compiled = fixture.nativeElement as HTMLElement;
      expect(compiled.textContent).not.toContain('marker-1');
      expect(compiled.textContent).not.toContain('2026-06-08T08:00:00Z');
    });
  });

  describe('FE-14 auto-refresh polling', () => {
    it('defaults auto-refresh to enabled and renders toggle', fakeAsync(() => {
      staffApiSpy.listAttendanceChildren.and.returnValue(of(mockChildren));
      fixture.detectChanges();

      expect(component.autoRefreshEnabled).toBeTrue();

      const toggle = fixture.nativeElement.querySelector('[data-testid="auto-refresh-toggle"]') as HTMLInputElement;
      expect(toggle).toBeTruthy();
      expect(toggle.checked).toBeTrue();
      fixture.destroy();
    }));

    it('polls every 30 seconds when auto-refresh is on', fakeAsync(() => {
      staffApiSpy.listAttendanceChildren.and.returnValue(of(mockChildren));
      fixture.detectChanges();

      const initialCalls = staffApiSpy.listAttendanceChildren.calls.count();

      tick(30000);
      fixture.detectChanges();

      expect(staffApiSpy.listAttendanceChildren.calls.count()).toBe(initialCalls + 1);

      tick(30000);
      fixture.detectChanges();

      expect(staffApiSpy.listAttendanceChildren.calls.count()).toBe(initialCalls + 2);
      fixture.destroy();
    }));

    it('stops polling when auto-refresh is toggled off', fakeAsync(() => {
      staffApiSpy.listAttendanceChildren.and.returnValue(of(mockChildren));
      fixture.detectChanges();

      component.toggleAutoRefresh();
      fixture.detectChanges();

      const callsAfterToggle = staffApiSpy.listAttendanceChildren.calls.count();

      tick(60000);
      fixture.detectChanges();

      expect(staffApiSpy.listAttendanceChildren.calls.count()).toBe(callsAfterToggle);
      fixture.destroy();
    }));

    it('resumes polling when auto-refresh is toggled back on', fakeAsync(() => {
      staffApiSpy.listAttendanceChildren.and.returnValue(of(mockChildren));
      fixture.detectChanges();

      component.toggleAutoRefresh();
      fixture.detectChanges();

      component.toggleAutoRefresh();
      fixture.detectChanges();

      const callsAfterResume = staffApiSpy.listAttendanceChildren.calls.count();

      tick(30000);
      fixture.detectChanges();

      expect(staffApiSpy.listAttendanceChildren.calls.count()).toBe(callsAfterResume + 1);
      fixture.destroy();
    }));

    it('manual refresh works when auto-refresh is on', fakeAsync(() => {
      staffApiSpy.listAttendanceChildren.and.returnValue(of(mockChildren));
      fixture.detectChanges();

      const initialCalls = staffApiSpy.listAttendanceChildren.calls.count();

      component.loadChildren('manual');
      fixture.detectChanges();

      expect(staffApiSpy.listAttendanceChildren.calls.count()).toBe(initialCalls + 1);
      fixture.destroy();
    }));

    it('shows last-updated timestamp after successful load', fakeAsync(() => {
      staffApiSpy.listAttendanceChildren.and.returnValue(of(mockChildren));
      fixture.detectChanges();

      const compiled = fixture.nativeElement as HTMLElement;
      expect(compiled.querySelector('[data-testid="last-updated"]')?.textContent).toContain('Updated');
      expect(component.lastUpdatedAt).not.toBeNull();
      fixture.destroy();
    }));

    it('background refresh does not show full-page loading over existing children', fakeAsync(() => {
      staffApiSpy.listAttendanceChildren.and.returnValue(of(mockChildren));
      fixture.detectChanges();

      let resolvePoll!: () => void;
      staffApiSpy.listAttendanceChildren.and.returnValue(
        new Observable((subscriber) => {
          resolvePoll = () => {
            subscriber.next(mockChildren);
            subscriber.complete();
          };
        }),
      );

      tick(30000);

      expect(component.isBackgroundRefreshing).toBeTrue();
      expect(component.isLoading).toBeFalse();

      fixture.detectChanges();
      const compiled = fixture.nativeElement as HTMLElement;
      expect(compiled.textContent).toContain('Ada Lovelace');

      resolvePoll();
      fixture.destroy();
    }));

    it('cleans up polling on component destroy', fakeAsync(() => {
      staffApiSpy.listAttendanceChildren.and.returnValue(of(mockChildren));
      fixture.detectChanges();

      fixture.destroy();

      const callsAfterDestroy = staffApiSpy.listAttendanceChildren.calls.count();

      tick(60000);

      expect(staffApiSpy.listAttendanceChildren.calls.count()).toBe(callsAfterDestroy);
    }));

    it('shows Refreshing indicator during background refresh', fakeAsync(() => {
      staffApiSpy.listAttendanceChildren.and.returnValue(of(mockChildren));
      fixture.detectChanges();

      let resolvePoll!: () => void;
      staffApiSpy.listAttendanceChildren.and.returnValue(
        new Observable((subscriber) => {
          resolvePoll = () => {
            subscriber.next(mockChildren);
            subscriber.complete();
          };
        }),
      );

      tick(30000);
      fixture.detectChanges();

      const compiled = fixture.nativeElement as HTMLElement;
      expect(compiled.querySelector('[data-testid="refreshing-indicator"]')?.textContent).toContain('Refreshing');

      resolvePoll();
      fixture.destroy();
    }));

    it('check-in action remains enabled during background refresh', fakeAsync(() => {
      staffApiSpy.listAttendanceChildren.and.returnValue(of(mockChildren));
      fixture.detectChanges();

      let resolvePoll!: () => void;
      staffApiSpy.listAttendanceChildren.and.returnValue(
        new Observable((subscriber) => {
          resolvePoll = () => {
            subscriber.next(mockChildren);
            subscriber.complete();
          };
        }),
      );

      tick(30000);

      const eligibleChild = mockChildren[0];
      expect(component.canCheckIn(eligibleChild)).toBeTrue();

      resolvePoll();
      fixture.destroy();
    }));

    it('skips poll when a list request is already in flight', fakeAsync(() => {
      staffApiSpy.listAttendanceChildren.and.returnValue(of(mockChildren));
      fixture.detectChanges();

      let resolvePoll!: () => void;
      staffApiSpy.listAttendanceChildren.and.returnValue(
        new Observable((subscriber) => {
          resolvePoll = () => {
            subscriber.next(mockChildren);
            subscriber.complete();
          };
        }),
      );

      tick(30000);
      const callsDuringInFlight = staffApiSpy.listAttendanceChildren.calls.count();

      tick(30000);
      expect(staffApiSpy.listAttendanceChildren.calls.count()).toBe(callsDuringInFlight);

      resolvePoll();
      fixture.destroy();
    }));
  });
});
