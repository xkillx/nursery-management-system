import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { AttendanceChildRecord, AbsenceMarkerRecord } from '../models/attendance-child.models';
import { ChildRecord } from '../models/children.models';
import { FundingProfileRecord, FundingOverviewRecord } from '../models/funding.models';
import { InviteRecord } from '../models/invites.models';
import { StaffApiService } from './staff-api.service';

describe('StaffApiService — Attendance', () => {
  let service: StaffApiService;
  let httpMock: HttpTestingController;

  const attendanceApiResponse = {
    items: [
      {
        id: 'child-1',first_name: 'Ada',
middle_name: null,
last_name: 'Lovelace',
        enrollment_complete: true,
        attendance_state: 'not_checked_in',
        open_session_id: null,
        checked_in_at: null,
        has_incomplete_session: false,
        absence_marker_id: null,
        absence_marked_at: null,
      },
      {
        id: 'child-2',first_name: 'Grace',
middle_name: null,
last_name: 'Hopper',
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
        firstName: 'Ada',
        middleName: null,
        lastName: 'Lovelace',
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
        firstName: 'Grace',
        middleName: null,
        lastName: 'Hopper',
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
          id: 'child-3',first_name: 'Margaret',
middle_name: null,
last_name: 'Hamilton',
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
        firstName: 'Margaret',
        middleName: null,
        lastName: 'Hamilton',
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
          id: 'child-4',first_name: 'No',
middle_name: null,
last_name: 'Absence',
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
    id: 'child-1',first_name: 'Ada',
middle_name: null,
last_name: 'Lovelace',
    date_of_birth: '2022-01-15',
    start_date: '2024-09-01',
    end_date: null,
    site_core_hourly_rate_minor: 750,
    notes: null,
    is_active: true,
    has_current_room: false,
    enrollment_complete: false,
    missing_requirements: ['parent_carer_contact'],
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
        firstName: 'Ada',
        middleName: null,
        lastName: 'Lovelace',
        fullName: 'Ada Lovelace',
        dateOfBirth: '2022-01-15',
        startDate: '2024-09-01',
        endDate: null,
        siteCoreHourlyRateMinor: 750,
        notes: null,
        isActive: true,
        hasCurrentRoom: false,
        hasBookingPattern: false,
        enrollmentComplete: false,
        missingRequirements: ['parent_carer_contact'],
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

  it('getFundingOverview sends billing_month query param and maps response', () => {
    const overviewApiModel = {
      billing_month: '2026-06',
      summary: {
        included_child_count: 3,
        flagged_child_count: 2,
        missing_profile_count: 1,
        explicit_zero_count: 1,
        under_one_hour_count: 0,
        above_160_hours_count: 0,
      },
      items: [
        {
          child_id: 'child-1',child_first_name: 'Alice',
child_middle_name: null,
child_last_name: null,
          is_active: true,
          start_date: '2026-01-01',
          end_date: null,
          flags: ['missing_profile'],
        },
        {
          child_id: 'child-2',child_first_name: 'Bob',
child_middle_name: null,
child_last_name: null,
          is_active: true,
          start_date: '2026-01-01',
          end_date: null,
          funding_profile_id: 'fp-2',
          funded_allowance_minutes: 0,
          funding_updated_at: '2026-06-01T10:00:00Z',
          flags: ['explicit_zero_allowance'],
        },
      ],
    };

    service.getFundingOverview('2026-06').subscribe((overview: FundingOverviewRecord) => {
      expect(overview.billingMonth).toBe('2026-06');
      expect(overview.summary.includedChildCount).toBe(3);
      expect(overview.summary.flaggedChildCount).toBe(2);
      expect(overview.summary.missingProfileCount).toBe(1);
      expect(overview.summary.explicitZeroCount).toBe(1);
      expect(overview.items.length).toBe(2);
      expect(overview.items[0].childId).toBe('child-1');
      expect(overview.items[0].flags).toEqual(['missing_profile']);
      expect(overview.items[0].fundedAllowanceMinutes).toBeNull();
      expect(overview.items[1].fundedAllowanceMinutes).toBe(0);
      expect(overview.items[1].fundingProfileId).toBe('fp-2');
    });

    const req = httpMock.expectOne((r) => r.url === '/api/v1/funding/overview');
    expect(req.request.method).toBe('GET');
    expect(req.request.params.get('billing_month')).toBe('2026-06');
    req.flush(overviewApiModel);
  });

  it('getFundingOverview maps nullable profile fields to null', () => {
    const emptyOverview = {
      billing_month: '2026-07',
      summary: {
        included_child_count: 0,
        flagged_child_count: 0,
        missing_profile_count: 0,
        explicit_zero_count: 0,
        under_one_hour_count: 0,
        above_160_hours_count: 0,
      },
      items: [],
    };

    service.getFundingOverview('2026-07').subscribe((overview: FundingOverviewRecord) => {
      expect(overview.items).toEqual([]);
      expect(overview.summary.includedChildCount).toBe(0);
    });

    const req = httpMock.expectOne((r) => r.url === '/api/v1/funding/overview');
    req.flush(emptyOverview);
  });
});

describe('StaffApiService — Attendance Corrections', () => {
  let service: StaffApiService;
  let httpMock: HttpTestingController;

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

  describe('listCorrectionSessions', () => {
    it('sends child_id and local_date query params', () => {
      service.listCorrectionSessions('child-1', '2026-06-09').subscribe();

      const req = httpMock.expectOne((r) => r.url === '/api/v1/attendance/sessions');
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('child_id')).toBe('child-1');
      expect(req.request.params.get('local_date')).toBe('2026-06-09');
      req.flush({
        child_id: 'child-1',
        selected_local_date: '2026-06-09',
        items: [],
      });
    });

    it('maps selectedLocalDate and invoice_warning object', (done) => {
      const response = {
        child_id: 'child-1',
        selected_local_date: '2026-06-09',
        invoice_warning: {
          billing_month: '2026-06',
          invoice_id: 'inv-1',
          invoice_number: 'INV-001',
          status: 'issued',
        },
        items: [
          {
            id: 'session-1',
            child_id: 'child-1',
            status: 'closed',
            check_in_at: '2026-06-09T08:00:00Z',
            check_out_at: '2026-06-09T15:00:00Z',
            check_in_local_date: '2026-06-09',
            check_out_local_date: '2026-06-09',
            duration_minutes: 420,
            created_at: '2026-06-09T08:00:00Z',
            updated_at: '2026-06-09T15:00:00Z',
          },
        ],
      };

      service.listCorrectionSessions('child-1', '2026-06-09').subscribe((ctx) => {
        expect(ctx.childId).toBe('child-1');
        expect(ctx.selectedLocalDate).toBe('2026-06-09');
        expect(ctx.invoiceWarning).not.toBeNull();
        expect(ctx.invoiceWarning!.billingMonth).toBe('2026-06');
        expect(ctx.invoiceWarning!.invoiceId).toBe('inv-1');
        expect(ctx.invoiceWarning!.invoiceNumber).toBe('INV-001');
        expect(ctx.invoiceWarning!.status).toBe('issued');
        expect(ctx.items.length).toBe(1);
        expect(ctx.items[0].id).toBe('session-1');
        expect(ctx.items[0].checkInAt).toBe('2026-06-09T08:00:00Z');
        expect(ctx.items[0].durationMinutes).toBe(420);
        done();
      });

      const req = httpMock.expectOne((r) => r.url === '/api/v1/attendance/sessions');
      req.flush(response);
    });

    it('maps null invoice_warning to null', (done) => {
      const response = {
        child_id: 'child-2',
        selected_local_date: '2026-06-10',
        invoice_warning: null as unknown as undefined,
        items: [],
      };

      service.listCorrectionSessions('child-2', '2026-06-10').subscribe((ctx) => {
        expect(ctx.invoiceWarning).toBeNull();
        expect(ctx.items).toEqual([]);
        done();
      });

      const req = httpMock.expectOne((r) => r.url === '/api/v1/attendance/sessions');
      req.flush(response);
    });

    it('maps multiple session items', (done) => {
      const response = {
        child_id: 'child-1',
        selected_local_date: '2026-06-09',
        items: [
          {
            id: 'session-1',
            child_id: 'child-1',
            status: 'closed',
            check_in_at: '2026-06-09T08:00:00Z',
            check_out_at: '2026-06-09T12:00:00Z',
            check_in_local_date: '2026-06-09',
            check_out_local_date: '2026-06-09',
            duration_minutes: 240,
            created_at: '2026-06-09T08:00:00Z',
            updated_at: '2026-06-09T12:00:00Z',
          },
          {
            id: 'session-2',
            child_id: 'child-1',
            status: 'closed',
            check_in_at: '2026-06-09T13:00:00Z',
            check_out_at: '2026-06-09T16:00:00Z',
            check_in_local_date: '2026-06-09',
            check_out_local_date: '2026-06-09',
            duration_minutes: 180,
            created_at: '2026-06-09T13:00:00Z',
            updated_at: '2026-06-09T16:00:00Z',
          },
        ],
      };

      service.listCorrectionSessions('child-1', '2026-06-09').subscribe((ctx) => {
        expect(ctx.items.length).toBe(2);
        expect(ctx.items[0].id).toBe('session-1');
        expect(ctx.items[0].durationMinutes).toBe(240);
        expect(ctx.items[1].id).toBe('session-2');
        expect(ctx.items[1].durationMinutes).toBe(180);
        done();
      });

      const req = httpMock.expectOne((r) => r.url === '/api/v1/attendance/sessions');
      req.flush(response);
    });
  });

  describe('getCorrectionHistory', () => {
    it('maps session and event fields including nullable fields to null', (done) => {
      const response = {
        session: {
          id: 'session-1',
          child_id: 'child-1',
          status: 'closed',
          check_in_at: '2026-06-09T08:00:00Z',
          check_out_at: '2026-06-09T15:00:00Z',
          check_in_local_date: '2026-06-09',
          check_out_local_date: '2026-06-09',
          duration_minutes: 420,
          created_at: '2026-06-09T08:00:00Z',
          updated_at: '2026-06-09T15:00:00Z',
        },
        items: [
          {
            id: 'evt-1',
            event_type: 'correction',
            occurred_at: '2026-06-09T16:00:00Z',
            local_date: '2026-06-09',
            recorded_by_user_id: 'user-1',
            recorded_by_membership_id: 'member-1',
            recorded_by_label: null,
            reason_code: null,
            reason_note: null,
            previous_check_in_at: null,
            previous_check_out_at: null,
            corrected_check_in_at: null,
            corrected_check_out_at: null,
            created_by_correction: true,
          },
        ],
      };

      service.getCorrectionHistory('session-1').subscribe((history) => {
        expect(history.session.id).toBe('session-1');
        expect(history.session.childId).toBe('child-1');
        expect(history.items.length).toBe(1);
        const evt = history.items[0];
        expect(evt.id).toBe('evt-1');
        expect(evt.eventType).toBe('correction');
        expect(evt.occurredAt).toBe('2026-06-09T16:00:00Z');
        expect(evt.localDate).toBe('2026-06-09');
        expect(evt.recordedByUserId).toBe('user-1');
        expect(evt.recordedByMembershipId).toBe('member-1');
        expect(evt.recordedByLabel).toBeNull();
        expect(evt.reasonCode).toBeNull();
        expect(evt.reasonNote).toBeNull();
        expect(evt.previousCheckInAt).toBeNull();
        expect(evt.previousCheckOutAt).toBeNull();
        expect(evt.correctedCheckInAt).toBeNull();
        expect(evt.correctedCheckOutAt).toBeNull();
        expect(evt.createdByCorrection).toBe(true);
        done();
      });

      const req = httpMock.expectOne('/api/v1/attendance/sessions/session-1/history');
      expect(req.request.method).toBe('GET');
      req.flush(response);
    });

    it('maps populated nullable fields for a correction event', (done) => {
      const response = {
        session: {
          id: 'session-1',
          child_id: 'child-1',
          status: 'closed',
          check_in_at: '2026-06-09T09:00:00Z',
          check_out_at: '2026-06-09T15:30:00Z',
          check_in_local_date: '2026-06-09',
          check_out_local_date: '2026-06-09',
          duration_minutes: 390,
          created_at: '2026-06-09T08:00:00Z',
          updated_at: '2026-06-09T16:00:00Z',
        },
        items: [
          {
            id: 'evt-2',
            event_type: 'check_in',
            occurred_at: '2026-06-09T16:10:00Z',
            local_date: '2026-06-09',
            recorded_by_user_id: 'user-2',
            recorded_by_membership_id: 'member-2',
            recorded_by_label: 'Manager',
            reason_code: 'incorrect_time',
            reason_note: 'Wrong time entered',
            previous_check_in_at: '2026-06-09T08:00:00Z',
            previous_check_out_at: '2026-06-09T15:00:00Z',
            corrected_check_in_at: '2026-06-09T09:00:00Z',
            corrected_check_out_at: '2026-06-09T15:30:00Z',
            created_by_correction: false,
          },
        ],
      };

      service.getCorrectionHistory('session-1').subscribe((history) => {
        const evt = history.items[0];
        expect(evt.recordedByLabel).toBe('Manager');
        expect(evt.reasonCode).toBe('incorrect_time');
        expect(evt.reasonNote).toBe('Wrong time entered');
        expect(evt.previousCheckInAt).toBe('2026-06-09T08:00:00Z');
        expect(evt.previousCheckOutAt).toBe('2026-06-09T15:00:00Z');
        expect(evt.correctedCheckInAt).toBe('2026-06-09T09:00:00Z');
        expect(evt.correctedCheckOutAt).toBe('2026-06-09T15:30:00Z');
        expect(evt.createdByCorrection).toBe(false);
        done();
      });

      const req = httpMock.expectOne('/api/v1/attendance/sessions/session-1/history');
      req.flush(response);
    });
  });

  describe('correctAttendance', () => {
    it('sends session_id for existing-session correction', (done) => {
      const response = {
        id: 'session-1',
        child_id: 'child-1',
        status: 'closed',
        check_in_at: '2026-06-09T09:00:00Z',
        check_out_at: '2026-06-09T15:30:00Z',
        check_in_local_date: '2026-06-09',
        check_out_local_date: '2026-06-09',
        duration_minutes: 390,
        created_at: '2026-06-09T08:00:00Z',
        updated_at: '2026-06-09T16:00:00Z',
      };

      service
        .correctAttendance({
          sessionId: 'session-1',
          checkInAt: '2026-06-09T09:00:00Z',
          checkOutAt: '2026-06-09T15:30:00Z',
          reasonCode: 'incorrect_time',
          reasonNote: 'Corrected times',
        })
        .subscribe((session) => {
          expect(session.id).toBe('session-1');
          expect(session.checkInAt).toBe('2026-06-09T09:00:00Z');
          expect(session.checkOutAt).toBe('2026-06-09T15:30:00Z');
          done();
        });

      const req = httpMock.expectOne('/api/v1/attendance/corrections');
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({
        session_id: 'session-1',
        check_in_at: '2026-06-09T09:00:00Z',
        check_out_at: '2026-06-09T15:30:00Z',
        reason_code: 'incorrect_time',
        reason_note: 'Corrected times',
      });
      req.flush(response);
    });

    it('sends child_id instead of session_id for missed-session correction', (done) => {
      const response = {
        id: 'session-new',
        child_id: 'child-1',
        status: 'closed',
        check_in_at: '2026-06-09T08:00:00Z',
        check_out_at: '2026-06-09T16:00:00Z',
        check_in_local_date: '2026-06-09',
        check_out_local_date: '2026-06-09',
        duration_minutes: 480,
        created_at: '2026-06-09T16:30:00Z',
        updated_at: '2026-06-09T16:30:00Z',
      };

      service
        .correctAttendance({
          childId: 'child-1',
          checkInAt: '2026-06-09T08:00:00Z',
          checkOutAt: '2026-06-09T16:00:00Z',
          reasonCode: 'missed_check_in',
          reasonNote: 'Forgot to check in',
        })
        .subscribe((session) => {
          expect(session.id).toBe('session-new');
          expect(session.durationMinutes).toBe(480);
          done();
        });

      const req = httpMock.expectOne('/api/v1/attendance/corrections');
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({
        child_id: 'child-1',
        check_in_at: '2026-06-09T08:00:00Z',
        check_out_at: '2026-06-09T16:00:00Z',
        reason_code: 'missed_check_in',
        reason_note: 'Forgot to check in',
      });
      req.flush(response);
    });
  });
});
