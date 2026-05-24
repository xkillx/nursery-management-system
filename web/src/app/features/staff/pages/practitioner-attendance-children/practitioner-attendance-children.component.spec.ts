import { provideHttpClient } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';
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
    },
    {
      id: 'child-2',
      fullName: 'Grace Hopper',
      enrollmentComplete: false,
      attendanceState: 'not_checked_in',
      openSessionId: null,
      checkedInAt: null,
      hasIncompleteSession: false,
    },
    {
      id: 'child-3',
      fullName: 'Katherine Johnson',
      enrollmentComplete: true,
      attendanceState: 'checked_in',
      openSessionId: 'session-1',
      checkedInAt: '2026-05-24T07:42:00Z',
      hasIncompleteSession: true,
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

  it('shows row error on conflict and reloads list', () => {
    setChildrenAndDetectChanges(mockChildren);

    const conflictError = new Error('conflict') as any;
    conflictError.status = 409;

    staffApiSpy.checkInChild.and.returnValue(throwError(() => conflictError));
    staffApiSpy.listAttendanceChildren.and.returnValue(of(mockChildren));

    component.checkIn(mockChildren[0]);
    fixture.detectChanges();

    expect(component.rowErrors['child-1']).toBeTruthy();
    expect(staffApiSpy.listAttendanceChildren).toHaveBeenCalledTimes(2);
  });

  it('displays global error when list load fails', () => {
    const loadError = new Error('server error') as any;
    loadError.status = 500;

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
});
