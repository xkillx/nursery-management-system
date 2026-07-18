import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { OwnerApiService } from './owner-api.service';

const siteSummariesApiResponse = {
  billing_month: '2026-06',
  attendance_local_date: '2026-06-11',
  currency_code: 'GBP',
  totals: {
    active_manager_count: 4,
    pending_manager_invite_count: 1,
    active_children_count: 60,
    checked_in_today_count: 45,
    incomplete_attendance_count: 3,
    draft_count: 2,
    issued_count: 10,
    overdue_count: 1,
    payment_failed_count: 0,
    paid_count: 8,
    total_issued_minor: 150000,
    total_paid_minor: 120000,
    outstanding_minor: 30000,
    overdue_outstanding_minor: 5000,
  },
  sites: [
    {
      site_id: 'site-1',
      site_name: 'Oak Lane',
      setup_status: 'complete',
      active_manager_count: 1,
      pending_manager_invite_count: 0,
      active_children_count: 15,
      attendance: { checked_in_today_count: 12, incomplete_attendance_count: 1 },
      funding_readiness: { included_child_count: 14, flagged_child_count: 1, missing_profile_count: 0, explicit_zero_count: 0, under_one_hour_count: 0, above_160_hours_count: 0 },
      invoice_payment_health: { draft_count: 0, issued_count: 3, overdue_count: 0, payment_failed_count: 0, paid_count: 3, total_issued_minor: 45000, total_paid_minor: 45000, outstanding_minor: 0, overdue_outstanding_minor: 0, failed_payment_count: 0 },
    },
    {
      site_id: 'site-2',
      site_name: 'Elm Street',
      setup_status: 'incomplete_attendance',
      active_manager_count: 0,
      pending_manager_invite_count: 1,
      active_children_count: 20,
      attendance: { checked_in_today_count: 15, incomplete_attendance_count: 2 },
      funding_readiness: { included_child_count: 18, flagged_child_count: 2, missing_profile_count: 1, explicit_zero_count: 0, under_one_hour_count: 1, above_160_hours_count: 0 },
      invoice_payment_health: { draft_count: 2, issued_count: 4, overdue_count: 1, payment_failed_count: 0, paid_count: 3, total_issued_minor: 60000, total_paid_minor: 45000, outstanding_minor: 15000, overdue_outstanding_minor: 5000, failed_payment_count: 0 },
    },
  ],
};

const managerAccessApiResponse = [
  { membership_id: 'mem-1', user_id: 'user-1', email: 'alice@example.com', is_active: true },
  { membership_id: 'mem-2', user_id: 'user-2', email: 'bob@example.com', is_active: false },
];

describe('OwnerApiService', () => {
  let service: OwnerApiService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });

    service = TestBed.inject(OwnerApiService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  describe('getSiteSummaries', () => {
    it('calls GET /owner/site-summaries and maps response', () => {
      service.getSiteSummaries().subscribe((response) => {
        expect(response.billingMonth).toBe('2026-06');
        expect(response.attendanceLocalDate).toBe('2026-06-11');
        expect(response.currencyCode).toBe('GBP');
        expect(response.totals.activeManagerCount).toBe(4);
        expect(response.totals.outstandingMinor).toBe(30000);
        expect(response.sites.length).toBe(2);
        expect(response.sites[0].siteId).toBe('site-1');
        expect(response.sites[0].siteName).toBe('Oak Lane');
        expect(response.sites[0].attendance.checkedInTodayCount).toBe(12);
        expect(response.sites[0].fundingReadiness.flaggedChildCount).toBe(1);
        expect(response.sites[0].invoicePaymentHealth.outstandingMinor).toBe(0);
      });

      const req = httpMock.expectOne('/api/v1/owner/site-summaries');
      expect(req.request.method).toBe('GET');
      req.flush(siteSummariesApiResponse);
    });

    it('sends billing_month and site_id query params', () => {
      service.getSiteSummaries({ billingMonth: '2026-05', siteId: 'site-1' }).subscribe();

      const req = httpMock.expectOne((r) => r.url === '/api/v1/owner/site-summaries');
      expect(req.request.params.get('billing_month')).toBe('2026-05');
      expect(req.request.params.get('site_id')).toBe('site-1');
      req.flush(siteSummariesApiResponse);
    });

    it('sends no query params when none provided', () => {
      service.getSiteSummaries().subscribe();

      const req = httpMock.expectOne('/api/v1/owner/site-summaries');
      expect(req.request.params.keys().length).toBe(0);
      req.flush(siteSummariesApiResponse);
    });
  });

  describe('listManagerAccess', () => {
    it('calls GET /owner/manager-access with site_id and status params', () => {
      service.listManagerAccess('site-1', 'active').subscribe((records) => {
        expect(records.length).toBe(2);
        expect(records[0]).toEqual({
          membershipId: 'mem-1',
          userId: 'user-1',
          email: 'alice@example.com',
          isActive: true,
        });
        expect(records[1].isActive).toBeFalse();
      });

      const req = httpMock.expectOne((r) => r.url === '/api/v1/owner/manager-access');
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('site_id')).toBe('site-1');
      expect(req.request.params.get('status')).toBe('active');
      req.flush({ items: managerAccessApiResponse, total: 2, page: 1, page_size: 25 });
    });

    it('defaults status to active', () => {
      service.listManagerAccess('site-1').subscribe();

      const req = httpMock.expectOne((r) => r.url === '/api/v1/owner/manager-access');
      expect(req.request.params.get('status')).toBe('active');
      req.flush({ items: [], total: 0, page: 1, page_size: 25 });
    });
  });

  describe('grantManagerAccess', () => {
    it('calls POST /owner/sites/:site_id/manager-access with email', () => {
      const apiResponse = {
        outcome: 'manager_membership_granted',
        membership_id: 'mem-3',
        invite: null,
      };

      service.grantManagerAccess('site-1', 'new@example.com').subscribe((result) => {
        expect(result.outcome).toBe('manager_membership_granted');
        expect(result.membershipId).toBe('mem-3');
        expect(result.invite).toBeNull();
      });

      const req = httpMock.expectOne('/api/v1/owner/sites/site-1/manager-access');
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({ email: 'new@example.com' });
      req.flush(apiResponse);
    });

    it('maps invite_pending outcome with invite details', () => {
      const apiResponse = {
        outcome: 'invite_pending',
        membership_id: null,
        invite: { email: 'invited@example.com', expires_at: '2026-06-18T12:00:00Z' },
      };

      service.grantManagerAccess('site-1', 'invited@example.com').subscribe((result) => {
        expect(result.outcome).toBe('invite_pending');
        expect(result.membershipId).toBeNull();
        expect(result.invite).toEqual({ email: 'invited@example.com', expiresAt: '2026-06-18T12:00:00Z' });
      });

      const req = httpMock.expectOne('/api/v1/owner/sites/site-1/manager-access');
      req.flush(apiResponse);
    });

    it('accepts Go-style manager_membership_granted outcome', () => {
      const apiResponse = {
        outcome: 'manager_membership_granted',
        membership_id: 'mem-4',
        invite: null,
      };

      service.grantManagerAccess('site-1', 'x@example.com').subscribe((result) => {
        expect(result.outcome).toBe('manager_membership_granted');
      });

      httpMock.expectOne('/api/v1/owner/sites/site-1/manager-access').flush(apiResponse);
    });

    it('accepts Go-style manager_membership_reactivated outcome', () => {
      const apiResponse = {
        outcome: 'manager_membership_reactivated',
        membership_id: 'mem-5',
        invite: null,
      };

      service.grantManagerAccess('site-1', 'x@example.com').subscribe((result) => {
        expect(result.outcome).toBe('manager_membership_reactivated');
      });

      httpMock.expectOne('/api/v1/owner/sites/site-1/manager-access').flush(apiResponse);
    });

    it('accepts Go-style manager_membership_already_active outcome', () => {
      const apiResponse = {
        outcome: 'manager_membership_already_active',
        membership_id: 'mem-6',
        invite: null,
      };

      service.grantManagerAccess('site-1', 'x@example.com').subscribe((result) => {
        expect(result.outcome).toBe('manager_membership_already_active');
      });

      httpMock.expectOne('/api/v1/owner/sites/site-1/manager-access').flush(apiResponse);
    });

    it('accepts Go-style manager_invite_pending outcome', () => {
      const apiResponse = {
        outcome: 'manager_invite_pending',
        membership_id: null,
        invite: { email: 'x@example.com', expires_at: '2026-06-18T12:00:00Z' },
      };

      service.grantManagerAccess('site-1', 'x@example.com').subscribe((result) => {
        expect(result.outcome).toBe('manager_invite_pending');
      });

      httpMock.expectOne('/api/v1/owner/sites/site-1/manager-access').flush(apiResponse);
    });
  });

  describe('deactivateManagerAccess', () => {
    it('calls POST deactivate action endpoint', () => {
      service.deactivateManagerAccess('site-1', 'mem-1').subscribe();

      const req = httpMock.expectOne('/api/v1/owner/sites/site-1/manager-access/mem-1/actions/deactivate');
      expect(req.request.method).toBe('POST');
      req.flush(null, { status: 204, statusText: 'No Content' });
    });
  });

  describe('reactivateManagerAccess', () => {
    it('calls POST activate action endpoint', () => {
      service.reactivateManagerAccess('site-1', 'mem-2').subscribe();

      const req = httpMock.expectOne('/api/v1/owner/sites/site-1/manager-access/mem-2/actions/activate');
      expect(req.request.method).toBe('POST');
      req.flush(null, { status: 204, statusText: 'No Content' });
    });
  });

  describe('rooms', () => {
    const roomApiResponse = {
      id: 'room-1',
      name: 'Baby Room',
      description: null,
      age_group: 'baby',
      capacity: 12,
      is_active: true,
      created_at: '2026-06-15T10:00:00Z',
      updated_at: '2026-06-15T10:00:00Z',
    };

    it('listRooms calls GET with site ID', () => {
      service.listRooms('site-1').subscribe((rooms) => {
        expect(rooms.length).toBe(1);
        expect(rooms[0].name).toBe('Baby Room');
        expect(rooms[0].ageGroup).toBe('baby');
        expect(rooms[0].capacity).toBe(12);
        expect(rooms[0].isActive).toBeTrue();
      });

      const req = httpMock.expectOne((r) => r.url === '/api/v1/sites/site-1/rooms');
      expect(req.request.method).toBe('GET');
      req.flush({ items: [roomApiResponse], total: 1, page: 1, page_size: 25 });
    });

    it('listRooms with include_archived sends query param', () => {
      service.listRooms('site-1', true).subscribe();

      const req = httpMock.expectOne((r) => r.url === '/api/v1/sites/site-1/rooms');
      expect(req.request.params.get('include_archived')).toBe('true');
      req.flush({ items: [], total: 0, page: 1, page_size: 25 });
    });

    it('getRoom calls GET with site and room IDs', () => {
      service.getRoom('site-1', 'room-1').subscribe((room) => {
        expect(room.name).toBe('Baby Room');
      });

      const req = httpMock.expectOne('/api/v1/sites/site-1/rooms/room-1');
      expect(req.request.method).toBe('GET');
      req.flush(roomApiResponse);
    });

    it('createRoom calls POST with body', () => {
      service.createRoom('site-1', { name: 'Baby Room', age_group: 'baby', capacity: 12 }).subscribe((room) => {
        expect(room.name).toBe('Baby Room');
      });

      const req = httpMock.expectOne('/api/v1/sites/site-1/rooms');
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({ name: 'Baby Room', age_group: 'baby', capacity: 12 });
      req.flush(roomApiResponse);
    });

    it('updateRoom calls PATCH with partial body', () => {
      service.updateRoom('site-1', 'room-1', { capacity: 15 }).subscribe((room) => {
        expect(room.capacity).toBe(12);
      });

      const req = httpMock.expectOne('/api/v1/sites/site-1/rooms/room-1');
      expect(req.request.method).toBe('PATCH');
      expect(req.request.body).toEqual({ capacity: 15 });
      req.flush(roomApiResponse);
    });

    it('archiveRoom calls POST action endpoint', () => {
      service.archiveRoom('site-1', 'room-1').subscribe();

      const req = httpMock.expectOne('/api/v1/sites/site-1/rooms/room-1/actions/archive');
      expect(req.request.method).toBe('POST');
      req.flush({});
    });

    it('reactivateRoom calls POST action endpoint', () => {
      service.reactivateRoom('site-1', 'room-1').subscribe((room) => {
        expect(room.isActive).toBeTrue();
      });

      const req = httpMock.expectOne('/api/v1/sites/site-1/rooms/room-1/actions/activate');
      expect(req.request.method).toBe('POST');
      req.flush(roomApiResponse);
    });
  });
});
