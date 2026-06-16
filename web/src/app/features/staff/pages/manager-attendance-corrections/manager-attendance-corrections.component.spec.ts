import { HttpErrorResponse } from '@angular/common/http';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter, Router } from '@angular/router';
import { of, throwError } from 'rxjs';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { AuthService } from '../../../../core/services/auth.service';
import { StaffApiService } from '../../../staff/data/staff-api.service';
import { ChildRecord } from '../../../staff/models/children.models';
import {
  AttendanceSessionRecord,
  CorrectionSessionContext,
} from '../../models/attendance-child.models';

import { ManagerAttendanceCorrectionsComponent } from './manager-attendance-corrections.component';

describe('ManagerAttendanceCorrectionsComponent', () => {
  let fixture: ComponentFixture<ManagerAttendanceCorrectionsComponent>;
  let component: ManagerAttendanceCorrectionsComponent;
  let native: HTMLElement;
  let apiService: jasmine.SpyObj<StaffApiService>;
  let authService: jasmine.SpyObj<AuthService>;

  const mockChild: ChildRecord = {
    id: 'child-1',
    fullName: 'Ada Lovelace',
    dateOfBirth: '2022-01-15',
    startDate: '2022-09-01',
    endDate: '2025-07-31',
    coreHourlyRateMinor: 750,
    siteCoreHourlyRateMinor: null,
    notes: null,
    isActive: true,
    leftAt: null,
    leftReasonCode: null,
    leftReasonNote: null,
    primaryRoomId: null,
    enrollmentComplete: true,
    missingRequirements: [],
    createdAt: '2022-08-01T00:00:00Z',
    updatedAt: '2022-08-01T00:00:00Z',
  };

  const mockChildNoEndDate: ChildRecord = {
    ...mockChild,
    id: 'child-2',
    endDate: null,
  };

  const mockSession: AttendanceSessionRecord = {
    id: 'session-1',
    childId: 'child-1',
    status: 'complete',
    checkInAt: '2024-06-01T08:00:00Z',
    checkOutAt: '2024-06-01T16:00:00Z',
    checkInLocalDate: '2024-06-01',
    checkOutLocalDate: '2024-06-01',
    durationMinutes: 480,
    createdAt: '2024-06-01T08:00:00Z',
    updatedAt: '2024-06-01T16:00:00Z',
  };

  const mockOpenSession: AttendanceSessionRecord = {
    ...mockSession,
    id: 'session-open',
    status: 'open',
    checkOutAt: null,
    checkOutLocalDate: null,
    durationMinutes: null,
  };

  const mockSessionContext: CorrectionSessionContext = {
    childId: 'child-1',
    selectedLocalDate: '2024-06-01',
    invoiceWarning: null,
    items: [mockSession, mockOpenSession],
  };

  function apiError(code: string, message: string, requestId?: string): HttpErrorResponse {
    return new HttpErrorResponse({
      error: { code, message, request_id: requestId },
      status: 400,
    });
  }

  beforeEach(async () => {
    const apiSpy = jasmine.createSpyObj('StaffApiService', [
      'listChildren',
      'listCorrectionSessions',
      'getCorrectionHistory',
      'correctAttendance',
    ]);
    const authSpy = jasmine.createSpyObj('AuthService', ['clearSession']);

    apiSpy.listChildren.and.returnValue(of([mockChild, mockChildNoEndDate]));
    apiSpy.listCorrectionSessions.and.returnValue(of(mockSessionContext));
    apiSpy.getCorrectionHistory.and.returnValue(
      of({ session: mockSession, items: [] }),
    );
    apiSpy.correctAttendance.and.returnValue(of(undefined));

    await TestBed.configureTestingModule({
      imports: [ManagerAttendanceCorrectionsComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: StaffApiService, useValue: apiSpy },
        { provide: AuthService, useValue: authSpy },
        ApiErrorMapper,
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ManagerAttendanceCorrectionsComponent);
    component = fixture.componentInstance;
    native = fixture.nativeElement;
    apiService = apiSpy;
    authService = authSpy;

    fixture.detectChanges();
  });

  it('creates the component', () => {
    expect(component).toBeTruthy();
  });

  describe('error mapping', () => {
    function setupSubmitReady(): void {
      component.selectedChildId = 'child-1';
      component.selectedLocalDate = '2024-06-01';
      component.selectedSessionId = 'session-1';
      component.checkInTime = '08:00';
      component.checkOutTime = '16:00';
      component.reasonCode = 'incorrect_time';
    }

    it('maps attendance_session_overlap to overlap error with actionable copy', () => {
      setupSubmitReady();
      apiService.correctAttendance.and.returnValue(throwError(() => apiError('attendance_session_overlap', 'overlap')));
      component.onSubmit();
      fixture.detectChanges();

      expect(component.correctionError?.kind).toBe('overlap');
      expect(component.correctionError?.title).toBe('Session overlap');
      expect(component.correctionError?.field).toBe('checkInTime');
      expect(native.textContent).toContain('Change the check-in or check-out time');
    });

    it('maps attendance_outside_enrollment_window with child enrollment range', () => {
      setupSubmitReady();
      apiService.correctAttendance.and.returnValue(throwError(() => apiError('attendance_outside_enrollment_window', 'outside')));
      component.onSubmit();
      fixture.detectChanges();

      expect(component.correctionError?.kind).toBe('enrollment_window');
      expect(component.correctionError?.title).toBe('Outside enrollment window');
      expect(native.textContent).toContain('Choose a date from');
      expect(native.textContent).toContain('2022');
      expect(native.textContent).toContain('2025');
    });

    it('maps attendance_outside_enrollment_window without end date', () => {
      setupSubmitReady();
      component.selectedChildId = 'child-2';
      apiService.correctAttendance.and.returnValue(throwError(() => apiError('attendance_outside_enrollment_window', 'outside')));
      component.onSubmit();
      fixture.detectChanges();

      expect(component.correctionError?.message).toContain('on or after');
      expect(component.correctionError?.message).toContain('2022');
    });

    it('maps attendance_correction_reason_required to reason field error', () => {
      setupSubmitReady();
      apiService.correctAttendance.and.returnValue(throwError(() => apiError('attendance_correction_reason_required', 'reason')));
      component.onSubmit();
      fixture.detectChanges();

      expect(component.correctionError?.kind).toBe('reason');
      expect(component.correctionError?.field).toBe('reasonCode');
      expect(component.correctionError?.message).toBe('Select a correction reason.');
    });

    it('maps attendance_correction_reason_invalid to reason field error', () => {
      setupSubmitReady();
      apiService.correctAttendance.and.returnValue(throwError(() => apiError('attendance_correction_reason_invalid', 'invalid')));
      component.onSubmit();

      expect(component.correctionError?.kind).toBe('reason');
      expect(component.correctionError?.field).toBe('reasonCode');
    });

    it('maps reason_note_required_for_other to note field error', () => {
      setupSubmitReady();
      apiService.correctAttendance.and.returnValue(throwError(() => apiError('reason_note_required_for_other', 'note')));
      component.onSubmit();
      fixture.detectChanges();

      expect(component.correctionError?.kind).toBe('note');
      expect(component.correctionError?.field).toBe('reasonNote');
      expect(component.correctionError?.message).toBe('Add a note when the reason is Other.');
    });

    it('maps attendance_invalid_time_order to time field error', () => {
      setupSubmitReady();
      apiService.correctAttendance.and.returnValue(throwError(() => apiError('attendance_invalid_time_order', 'order')));
      component.onSubmit();
      fixture.detectChanges();

      expect(component.correctionError?.kind).toBe('time_order');
      expect(component.correctionError?.field).toBe('checkInTime');
      expect(native.textContent).toContain('Set check-out after check-in.');
    });

    it('maps attendance_correction_future_time to time field error', () => {
      setupSubmitReady();
      apiService.correctAttendance.and.returnValue(throwError(() => apiError('attendance_correction_future_time', 'future')));
      component.onSubmit();
      fixture.detectChanges();

      expect(component.correctionError?.kind).toBe('future_time');
      expect(component.correctionError?.field).toBe('checkInTime');
      expect(native.textContent).toContain('already happened');
    });

    it('maps forbidden_role to authorization error', () => {
      setupSubmitReady();
      apiService.correctAttendance.and.returnValue(throwError(() => apiError('forbidden_role', 'forbidden')));
      component.onSubmit();
      fixture.detectChanges();

      expect(component.correctionError?.kind).toBe('authorization');
      expect(component.correctionError?.title).toBe('No access');
      expect(native.textContent).toContain('Sign in as a manager');
    });

    it('maps forbidden_scope_selection to authorization error', () => {
      setupSubmitReady();
      apiService.correctAttendance.and.returnValue(throwError(() => apiError('forbidden_scope_selection', 'forbidden')));
      component.onSubmit();

      expect(component.correctionError?.kind).toBe('authorization');
    });

    it('maps attendance_session_not_found to not_found error', () => {
      setupSubmitReady();
      apiService.correctAttendance.and.returnValue(throwError(() => apiError('attendance_session_not_found', 'not found')));
      component.onSubmit();
      fixture.detectChanges();

      expect(component.correctionError?.kind).toBe('not_found');
      expect(native.textContent).toContain('Select the session again');
    });

    it('maps child_not_found to not_found error', () => {
      setupSubmitReady();
      apiService.correctAttendance.and.returnValue(throwError(() => apiError('child_not_found', 'not found')));
      component.onSubmit();
      fixture.detectChanges();

      expect(component.correctionError?.kind).toBe('not_found');
      expect(native.textContent).toContain('Select the child again');
    });

    it('falls back to mapped message and request ID for unknown errors', () => {
      setupSubmitReady();
      apiService.correctAttendance.and.returnValue(
        throwError(() => apiError('unknown_code', 'Something broke', 'req-123')),
      );
      component.onSubmit();
      fixture.detectChanges();

      expect(component.correctionError?.kind).toBe('generic');
      expect(component.correctionError?.message).toContain('Something went wrong');
      expect(component.correctionError?.message).toContain('Request: req-123');
    });

    it('clears session and redirects on unauthorized', () => {
      setupSubmitReady();
      apiService.correctAttendance.and.returnValue(
        throwError(() => new HttpErrorResponse({ error: { code: 'unauthorized', message: 'no' }, status: 401 })),
      );

      spyOn(TestBed.inject(Router), 'navigate');
      component.onSubmit();

      expect(authService.clearSession).toHaveBeenCalled();
    });
  });

  describe('local validation guidance', () => {
    it('shows reason guidance when reason is missing in correction form', () => {
      component.selectedChildId = 'child-1';
      component.selectedLocalDate = '2024-06-01';
      component.selectedSessionId = 'session-1';
      component.reasonCode = '';
      fixture.detectChanges();

      expect(component.localFieldErrors['reasonCode']).toBe('Select a correction reason.');
    });

    it('shows note guidance when other is selected without a note', () => {
      component.selectedChildId = 'child-1';
      component.selectedLocalDate = '2024-06-01';
      component.selectedSessionId = 'session-1';
      component.reasonCode = 'other';
      component.reasonNote = '';
      fixture.detectChanges();

      expect(component.localFieldErrors['reasonNote']).toBe('Add a note when the reason is Other.');
    });

    it('shows time order hint when check-out is not after check-in', () => {
      component.selectedChildId = 'child-1';
      component.selectedLocalDate = '2024-06-01';
      component.selectedSessionId = 'session-1';
      component.checkInTime = '16:00';
      component.checkOutTime = '08:00';
      fixture.detectChanges();

      expect(component.localFieldErrors['timeOrder']).toBe('Set check-out after check-in.');
    });

    it('does not show validation hints when form is not visible', () => {
      component.selectedChildId = null;
      fixture.detectChanges();

      expect(Object.keys(component.localFieldErrors).length).toBe(0);
    });
  });

  describe('incomplete session', () => {
    it('detects incomplete session when selected session has no check-out', () => {
      component.selectedChildId = 'child-1';
      component.selectedLocalDate = '2024-06-01';
      component.selectedSessionId = 'session-open';
      component.sessions = [mockSession, mockOpenSession];
      fixture.detectChanges();

      expect(component.isSessionIncomplete).toBeTrue();
      expect(native.textContent).toContain('Incomplete session');
      expect(native.textContent).toContain('complete corrected check-in and check-out');
    });

    it('does not show incomplete warning for complete session', () => {
      component.selectedChildId = 'child-1';
      component.selectedLocalDate = '2024-06-01';
      component.selectedSessionId = 'session-1';
      component.sessions = [mockSession, mockOpenSession];
      fixture.detectChanges();

      expect(component.isSessionIncomplete).toBeFalse();
    });
  });

  describe('error clearing', () => {
    function setOverlapError(): void {
      component.correctionError = {
        kind: 'overlap',
        title: 'Session overlap',
        message: 'Change times.',
        field: 'checkInTime',
      };
    }

    it('clears correction error on child change', () => {
      setOverlapError();
      component.onChildChange();
      expect(component.correctionError).toBeNull();
    });

    it('clears correction error on date change', () => {
      setOverlapError();
      component.onDateChange();
      expect(component.correctionError).toBeNull();
    });

    it('clears correction error on session select', () => {
      setOverlapError();
      component.onSessionSelect('session-1');
      expect(component.correctionError).toBeNull();
    });

    it('clears time field error on time change', () => {
      setOverlapError();
      component.onTimeChange();
      expect(component.correctionError).toBeNull();
    });

    it('clears reason field error on reason change', () => {
      component.correctionError = {
        kind: 'reason',
        title: 'Missing reason',
        message: 'Select reason.',
        field: 'reasonCode',
      };
      component.onReasonChange();
      expect(component.correctionError).toBeNull();
    });

    it('clears note field error on note change', () => {
      component.correctionError = {
        kind: 'note',
        title: 'Note required',
        message: 'Add note.',
        field: 'reasonNote',
      };
      component.onNoteChange();
      expect(component.correctionError).toBeNull();
    });

    it('does not clear non-matching field error', () => {
      component.correctionError = {
        kind: 'overlap',
        title: 'Session overlap',
        message: 'Change times.',
        field: 'checkInTime',
      };
      component.onReasonChange();
      expect(component.correctionError).not.toBeNull();
    });
  });
});
