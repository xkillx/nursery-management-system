import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { AttendanceChildRecord } from '../models/attendance-child.models';
import { StaffApiService } from './staff-api.service';

describe('StaffApiService — Attendance', () => {
  let service: StaffApiService;
  let httpMock: HttpTestingController;

  const attendanceApiResponse = {
    items: [
      {
        id: 'child-1',
        full_name: 'Ada Lovelace',
        enrollment_complete: true,
        attendance_state: 'not_checked_in',
        open_session_id: null,
        checked_in_at: null,
      },
      {
        id: 'child-2',
        full_name: 'Grace Hopper',
        enrollment_complete: false,
        attendance_state: 'checked_in',
        open_session_id: 'session-1',
        checked_in_at: '2026-05-24T07:42:00Z',
      },
    ],
  };

  const sessionApiResponse = {
    id: 'session-1',
    child_id: 'child-1',
    status: 'open',
    check_in_at: '2026-05-24T07:42:00Z',
    check_out_at: null,
    check_in_local_date: '2026-05-24',
    check_out_local_date: null,
    duration_minutes: null,
    created_at: '2026-05-24T07:42:00Z',
    updated_at: '2026-05-24T07:42:00Z',
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });

    service = TestBed.inject(StaffApiService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('listAttendanceChildren maps snake_case to camelCase', () => {
    service.listAttendanceChildren().subscribe((children: AttendanceChildRecord[]) => {
      expect(children.length).toBe(2);

      expect(children[0]).toEqual({
        id: 'child-1',
        fullName: 'Ada Lovelace',
        enrollmentComplete: true,
        attendanceState: 'not_checked_in',
        openSessionId: null,
        checkedInAt: null,
      });

      expect(children[1]).toEqual({
        id: 'child-2',
        fullName: 'Grace Hopper',
        enrollmentComplete: false,
        attendanceState: 'checked_in',
        openSessionId: 'session-1',
        checkedInAt: '2026-05-24T07:42:00Z',
      });
    });

    const req = httpMock.expectOne('/api/v1/children/attendance');
    expect(req.request.method).toBe('GET');
    req.flush(attendanceApiResponse);
  });

  it('checkInChild POSTs to /attendance/check-ins with child_id', () => {
    service.checkInChild('child-1').subscribe((session) => {
      expect(session.id).toBe('session-1');
      expect(session.childId).toBe('child-1');
      expect(session.checkInAt).toBe('2026-05-24T07:42:00Z');
      expect(session.checkOutAt).toBeNull();
      expect(session.checkInLocalDate).toBe('2026-05-24');
      expect(session.checkOutLocalDate).toBeNull();
      expect(session.durationMinutes).toBeNull();
    });

    const req = httpMock.expectOne('/api/v1/attendance/check-ins');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ child_id: 'child-1' });
    req.flush(sessionApiResponse);
  });

  it('checkOutChild POSTs to /attendance/check-outs with child_id', () => {
    const checkoutResponse = {
      ...sessionApiResponse,
      status: 'closed',
      check_out_at: '2026-05-24T15:30:00Z',
      check_out_local_date: '2026-05-24',
      duration_minutes: 468,
      updated_at: '2026-05-24T15:30:00Z',
    };

    service.checkOutChild('child-1').subscribe((session) => {
      expect(session.status).toBe('closed');
      expect(session.checkOutAt).toBe('2026-05-24T15:30:00Z');
      expect(session.checkOutLocalDate).toBe('2026-05-24');
      expect(session.durationMinutes).toBe(468);
    });

    const req = httpMock.expectOne('/api/v1/attendance/check-outs');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ child_id: 'child-1' });
    req.flush(checkoutResponse);
  });
});
