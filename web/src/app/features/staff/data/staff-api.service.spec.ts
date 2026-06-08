import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { AttendanceChildRecord, AbsenceMarkerRecord } from '../models/attendance-child.models';
import { ChildRecord } from '../models/children.models';
import { FundingProfileRecord } from '../models/funding.models';
import { InviteRecord } from '../models/invites.models';
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
        has_incomplete_session: false,
        absence_marker_id: null,
        absence_marked_at: null,
      },
      {
        id: 'child-2',
        full_name: 'Grace Hopper',
        enrollment_complete: false,
        attendance_state: 'checked_in',
        open_session_id: 'session-1',
        checked_in_at: '2026-05-24T07:42:00Z',
        has_incomplete_session: true,
        absence_marker_id: null,
        absence_marked_at: null,
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
        hasIncompleteSession: false,
        absenceMarkerId: null,
        absenceMarkedAt: null,
      });

      expect(children[1]).toEqual({
        id: 'child-2',
        fullName: 'Grace Hopper',
        enrollmentComplete: false,
        attendanceState: 'checked_in',
        openSessionId: 'session-1',
        checkedInAt: '2026-05-24T07:42:00Z',
        hasIncompleteSession: true,
        absenceMarkerId: null,
        absenceMarkedAt: null,
      });
    });

    const req = httpMock.expectOne('/api/v1/children/attendance');
    expect(req.request.method).toBe('GET');
    req.flush(attendanceApiResponse);
  });

  it('maps absent attendance_state and absence marker fields', () => {
    const absentResponse = {
      items: [
        {
          id: 'child-3',
          full_name: 'Margaret Hamilton',
          enrollment_complete: true,
          attendance_state: 'absent',
          open_session_id: null,
          checked_in_at: null,
          has_incomplete_session: false,
          absence_marker_id: 'marker-1',
          absence_marked_at: '2026-06-08T08:00:00Z',
        },
      ],
    };

    service.listAttendanceChildren().subscribe((children) => {
      expect(children[0]).toEqual({
        id: 'child-3',
        fullName: 'Margaret Hamilton',
        enrollmentComplete: true,
        attendanceState: 'absent',
        openSessionId: null,
        checkedInAt: null,
        hasIncompleteSession: false,
        absenceMarkerId: 'marker-1',
        absenceMarkedAt: '2026-06-08T08:00:00Z',
      });
    });

    const req = httpMock.expectOne('/api/v1/children/attendance');
    req.flush(absentResponse);
  });

  it('maps missing absence_marker_id and absence_marked_at to null', () => {
    const noAbsenceResponse = {
      items: [
        {
          id: 'child-4',
          full_name: 'No Absence',
          enrollment_complete: true,
          attendance_state: 'not_checked_in',
          open_session_id: null,
          checked_in_at: null,
          has_incomplete_session: false,
        },
      ],
    };

    service.listAttendanceChildren().subscribe((children) => {
      expect(children[0].absenceMarkerId).toBeNull();
      expect(children[0].absenceMarkedAt).toBeNull();
    });

    const req = httpMock.expectOne('/api/v1/children/attendance');
    req.flush(noAbsenceResponse);
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

  const absenceMarkerApiResponse = {
    id: 'marker-1',
    child_id: 'child-1',
    local_date: '2026-06-08',
    marked_at: '2026-06-08T08:00:00Z',
    cleared_at: null,
    created_at: '2026-06-08T08:00:00Z',
    updated_at: '2026-06-08T08:00:00Z',
  };

  it('markChildAbsent POSTs to /attendance/absence-markers with child_id', () => {
    service.markChildAbsent('child-1').subscribe((marker: AbsenceMarkerRecord) => {
      expect(marker).toEqual({
        id: 'marker-1',
        childId: 'child-1',
        localDate: '2026-06-08',
        markedAt: '2026-06-08T08:00:00Z',
        clearedAt: null,
        createdAt: '2026-06-08T08:00:00Z',
        updatedAt: '2026-06-08T08:00:00Z',
      });
    });

    const req = httpMock.expectOne('/api/v1/attendance/absence-markers');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ child_id: 'child-1' });
    req.flush(absenceMarkerApiResponse);
  });

  it('clearAbsenceMarker POSTs to /attendance/absence-markers/{id}/clear with null body', () => {
    const clearedResponse = {
      ...absenceMarkerApiResponse,
      cleared_at: '2026-06-08T09:00:00Z',
      updated_at: '2026-06-08T09:00:00Z',
    };

    service.clearAbsenceMarker('marker-1').subscribe((marker: AbsenceMarkerRecord) => {
      expect(marker.clearedAt).toBe('2026-06-08T09:00:00Z');
    });

    const req = httpMock.expectOne('/api/v1/attendance/absence-markers/marker-1/clear');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toBeNull();
    req.flush(clearedResponse);
  });

  it('maps nullable cleared_at to null in absence marker response', () => {
    service.markChildAbsent('child-1').subscribe((marker: AbsenceMarkerRecord) => {
      expect(marker.clearedAt).toBeNull();
    });

    const req = httpMock.expectOne('/api/v1/attendance/absence-markers');
    req.flush(absenceMarkerApiResponse);
  });
});

describe('StaffApiService — Invites', () => {
  let service: StaffApiService;
  let httpMock: HttpTestingController;

  const inviteApiModel = {
    id: 'invite-1',
    email: 'practitioner@example.com',
    role: 'practitioner',
    status: 'pending',
    expires_at: '2026-06-13T00:00:00Z',
    accepted_at: null,
    revoked_at: null,
    created_at: '2026-06-06T10:00:00Z',
    updated_at: '2026-06-06T10:00:00Z',
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

  it('listInvites defaults to status=pending', () => {
    service.listInvites().subscribe((invites: InviteRecord[]) => {
      expect(invites.length).toBe(1);
      expect(invites[0]).toEqual({
        id: 'invite-1',
        email: 'practitioner@example.com',
        role: 'practitioner',
        status: 'pending',
        expiresAt: '2026-06-13T00:00:00Z',
        acceptedAt: null,
        revokedAt: null,
        createdAt: '2026-06-06T10:00:00Z',
        updatedAt: '2026-06-06T10:00:00Z',
      });
    });

    const req = httpMock.expectOne((r) => r.url === '/api/v1/invites');
    expect(req.request.method).toBe('GET');
    expect(req.request.params.get('status')).toBe('pending');
    req.flush({ items: [inviteApiModel] });
  });

  it('listInvites passes status filter to query params', () => {
    service.listInvites('all').subscribe();

    const req = httpMock.expectOne((r) => r.url === '/api/v1/invites');
    expect(req.request.params.get('status')).toBe('all');
    req.flush({ items: [] });
  });

  it('createInvite posts email and role', () => {
    service.createInvite({ email: 'parent@example.com', role: 'parent' }).subscribe((invite: InviteRecord) => {
      expect(invite.role).toBe('parent');
      expect(invite.email).toBe('parent@example.com');
    });

    const req = httpMock.expectOne('/api/v1/invites');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ email: 'parent@example.com', role: 'parent' });
    req.flush({ ...inviteApiModel, email: 'parent@example.com', role: 'parent' });
  });

  it('resendInvite posts to /invites/{id}/resend', () => {
    service.resendInvite('invite-1').subscribe((invite: InviteRecord) => {
      expect(invite.id).toBe('invite-1');
    });

    const req = httpMock.expectOne('/api/v1/invites/invite-1/resend');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toBeNull();
    req.flush(inviteApiModel);
  });

  it('revokeInvite posts to /invites/{id}/revoke', () => {
    service.revokeInvite('invite-1').subscribe((invite: InviteRecord) => {
      expect(invite.status).toBe('pending');
    });

    const req = httpMock.expectOne('/api/v1/invites/invite-1/revoke');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toBeNull();
    req.flush(inviteApiModel);
  });

  it('maps nullable accepted_at and revoked_at to null', () => {
    const acceptedModel = {
      ...inviteApiModel,
      status: 'accepted',
      accepted_at: '2026-06-07T09:00:00Z',
    };

    service.listInvites('accepted').subscribe((invites: InviteRecord[]) => {
      expect(invites[0].acceptedAt).toBe('2026-06-07T09:00:00Z');
      expect(invites[0].revokedAt).toBeNull();
    });

    const req = httpMock.expectOne((r) => r.url === '/api/v1/invites');
    req.flush({ items: [acceptedModel] });
  });
});

describe('StaffApiService — Children', () => {
  let service: StaffApiService;
  let httpMock: HttpTestingController;

  const childApiModel = {
    id: 'child-1',
    full_name: 'Ada Lovelace',
    date_of_birth: '2022-01-15',
    start_date: '2024-09-01',
    end_date: null,
    core_hourly_rate_minor: 750,
    notes: null,
    is_active: true,
    left_at: null,
    left_reason_code: null,
    left_reason_note: null,
    enrollment_complete: false,
    missing_requirements: ['guardian_link', 'billing_rate'],
    created_at: '2024-08-01T00:00:00Z',
    updated_at: '2024-08-01T00:00:00Z',
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

  it('listChildren sends status, limit, offset query params', () => {
    service.listChildren({ status: 'active', limit: 10, offset: 0 }).subscribe();

    const req = httpMock.expectOne((r) => r.url === '/api/v1/children');
    expect(req.request.method).toBe('GET');
    expect(req.request.params.get('status')).toBe('active');
    expect(req.request.params.get('limit')).toBe('10');
    expect(req.request.params.get('offset')).toBe('0');
    req.flush({ items: [] });
  });

  it('listChildren maps snake_case to camelCase with enrollment fields', () => {
    service.listChildren({ status: 'active', limit: 10, offset: 0 }).subscribe((children: ChildRecord[]) => {
      expect(children.length).toBe(1);
      expect(children[0]).toEqual({
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
        enrollmentComplete: false,
        missingRequirements: ['guardian_link', 'billing_rate'],
        createdAt: '2024-08-01T00:00:00Z',
        updatedAt: '2024-08-01T00:00:00Z',
      });
    });

    const req = httpMock.expectOne((r) => r.url === '/api/v1/children');
    req.flush({ items: [childApiModel] });
  });

  it('listChildren defaults missing_requirements to empty array', () => {
    const noRequirements = { ...childApiModel, missing_requirements: undefined };

    service.listChildren({ status: 'active', limit: 10, offset: 0 }).subscribe((children: ChildRecord[]) => {
      expect(children[0].missingRequirements).toEqual([]);
    });

    const req = httpMock.expectOne((r) => r.url === '/api/v1/children');
    req.flush({ items: [noRequirements] });
  });
});

describe('StaffApiService — Guardians', () => {
  let service: StaffApiService;
  let httpMock: HttpTestingController;

  const guardianApiModel = {
    id: 'guardian-1',
    full_name: 'Sarah Thompson',
    email: 'sarah@example.com',
    phone: '+44 7700 900001',
    notes: null,
    is_active: true,
    deactivated_at: null,
    deactivation_reason_code: null,
    deactivation_reason_note: null,
    created_at: '2024-08-01T00:00:00Z',
    updated_at: '2024-08-01T00:00:00Z',
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

  it('listGuardians sends status, limit, offset query params', () => {
    service.listGuardians({ status: 'inactive', limit: 5, offset: 10 }).subscribe();

    const req = httpMock.expectOne((r) => r.url === '/api/v1/guardians');
    expect(req.request.method).toBe('GET');
    expect(req.request.params.get('status')).toBe('inactive');
    expect(req.request.params.get('limit')).toBe('5');
    expect(req.request.params.get('offset')).toBe('10');
    req.flush({ items: [] });
  });

  it('listGuardians maps nullable contact and lifecycle fields', () => {
    const minimalGuardian = {
      ...guardianApiModel,
      email: undefined,
      phone: undefined,
      notes: undefined,
    };

    service.listGuardians({ status: 'active', limit: 10, offset: 0 }).subscribe((guardians) => {
      expect(guardians.length).toBe(1);
      expect(guardians[0]).toEqual({
        id: 'guardian-1',
        fullName: 'Sarah Thompson',
        email: null,
        phone: null,
        notes: null,
        isActive: true,
        deactivatedAt: null,
        deactivationReasonCode: null,
        deactivationReasonNote: null,
        createdAt: '2024-08-01T00:00:00Z',
        updatedAt: '2024-08-01T00:00:00Z',
      });
    });

    const req = httpMock.expectOne((r) => r.url === '/api/v1/guardians');
    req.flush({ items: [minimalGuardian] });
  });
});

describe('StaffApiService — Funding', () => {
  let service: StaffApiService;
  let httpMock: HttpTestingController;

  const fundingProfileApiModel = {
    id: 'fp-1',
    child_id: 'child-1',
    billing_month: '2026-06',
    funded_allowance_minutes: 570,
    created_at: '2026-06-01T10:00:00Z',
    updated_at: '2026-06-08T12:00:00Z',
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

  it('getFundingProfile sends billing_month query param and maps response', () => {
    service.getFundingProfile('child-1', '2026-06').subscribe((profile: FundingProfileRecord) => {
      expect(profile).toEqual({
        id: 'fp-1',
        childId: 'child-1',
        billingMonth: '2026-06',
        fundedAllowanceMinutes: 570,
        createdAt: '2026-06-01T10:00:00Z',
        updatedAt: '2026-06-08T12:00:00Z',
      });
    });

    const req = httpMock.expectOne((r) => r.url === '/api/v1/funding/children/child-1');
    expect(req.request.method).toBe('GET');
    expect(req.request.params.get('billing_month')).toBe('2026-06');
    req.flush(fundingProfileApiModel);
  });

  it('upsertFundingProfile PUTs billing_month and funded_allowance_minutes', () => {
    service
      .upsertFundingProfile('child-1', { billing_month: '2026-06', funded_allowance_minutes: 570 })
      .subscribe((profile: FundingProfileRecord) => {
        expect(profile.fundedAllowanceMinutes).toBe(570);
        expect(profile.billingMonth).toBe('2026-06');
      });

    const req = httpMock.expectOne('/api/v1/funding/children/child-1');
    expect(req.request.method).toBe('PUT');
    expect(req.request.body).toEqual({ billing_month: '2026-06', funded_allowance_minutes: 570 });
    req.flush(fundingProfileApiModel);
  });

  it('upsertFundingProfile accepts 201 created response', () => {
    const createdModel = {
      ...fundingProfileApiModel,
      id: 'fp-new',
      funded_allowance_minutes: 0,
      created_at: '2026-06-08T14:00:00Z',
      updated_at: '2026-06-08T14:00:00Z',
    };

    service
      .upsertFundingProfile('child-1', { billing_month: '2026-06', funded_allowance_minutes: 0 })
      .subscribe((profile: FundingProfileRecord) => {
        expect(profile.id).toBe('fp-new');
        expect(profile.fundedAllowanceMinutes).toBe(0);
      });

    const req = httpMock.expectOne('/api/v1/funding/children/child-1');
    expect(req.request.method).toBe('PUT');
    req.flush(createdModel, { status: 201, statusText: 'Created' });
  });
});
