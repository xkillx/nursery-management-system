import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, provideRouter } from '@angular/router';

import { AuthService } from '../../../../core/services/auth.service';
import { OwnerRoomsComponent } from './owner-rooms.component';

const summariesResponse = {
  billing_month: '2026-06',
  attendance_local_date: '2026-06-11',
  currency_code: 'GBP',
  totals: {
    active_manager_count: 1,
    pending_manager_invite_count: 0,
    active_children_count: 15,
    checked_in_today_count: 12,
    incomplete_attendance_count: 0,
    draft_count: 0,
    issued_count: 3,
    overdue_count: 0,
    payment_failed_count: 0,
    paid_count: 3,
    total_issued_minor: 45000,
    total_paid_minor: 45000,
    outstanding_minor: 0,
    overdue_outstanding_minor: 0,
  },
  sites: [
    {
      site_id: 'site-1',
      site_name: 'Oak Lane',
      setup_status: 'complete',
      active_manager_count: 1,
      pending_manager_invite_count: 0,
      active_children_count: 15,
      site_core_hourly_rate_minor: 750,
      setup_issues: [],
      attendance: { checked_in_today_count: 12, incomplete_attendance_count: 0 },
      funding_readiness: { included_child_count: 14, flagged_child_count: 0, missing_profile_count: 0, explicit_zero_count: 0, under_one_hour_count: 0, above_160_hours_count: 0 },
      invoice_payment_health: { draft_count: 0, issued_count: 3, overdue_count: 0, payment_failed_count: 0, paid_count: 3, total_issued_minor: 45000, total_paid_minor: 45000, outstanding_minor: 0, overdue_outstanding_minor: 0, failed_payment_count: 0 },
    },
  ],
};

const roomsResponse = {
  rooms: [
    {
      id: 'room-1',
      name: 'Baby Room',
      description: 'Calm baby space',
      age_group: 'baby',
      capacity: 12,
      is_active: true,
      created_at: '2026-06-01T00:00:00Z',
      updated_at: '2026-06-01T00:00:00Z',
    },
    {
      id: 'room-2',
      name: 'Sensory Hub',
      description: null,
      age_group: 'mixed',
      capacity: 8,
      is_active: false,
      created_at: '2026-06-01T00:00:00Z',
      updated_at: '2026-06-01T00:00:00Z',
    },
  ],
};

describe('OwnerRoomsComponent', () => {
  let component: OwnerRoomsComponent;
  let fixture: ComponentFixture<OwnerRoomsComponent>;
  let httpMock: HttpTestingController;
  let authStub: jasmine.SpyObj<AuthService>;

  beforeEach(async () => {
    authStub = jasmine.createSpyObj<AuthService>('AuthService', ['currentRole', 'activeMembership']);
    authStub.currentRole.and.returnValue('owner');
    authStub.activeMembership.and.returnValue(null);

    await TestBed.configureTestingModule({
      imports: [OwnerRoomsComponent],
      providers: [
        provideRouter([]),
        provideHttpClient(),
        provideHttpClientTesting(),
        { provide: AuthService, useValue: authStub },
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { queryParamMap: { get: () => null } } },
        },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(OwnerRoomsComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('loads owner sites and rooms with archived rooms included', () => {
    fixture.detectChanges();

    httpMock.expectOne((req) => req.url === '/api/v1/owner/site-summaries').flush(summariesResponse);
    const roomsReq = httpMock.expectOne((req) => req.url === '/api/v1/sites/site-1/rooms');
    expect(roomsReq.request.params.get('include_archived')).toBe('true');
    roomsReq.flush(roomsResponse);
    fixture.detectChanges();

    expect(component.selectedSiteId).toBe('site-1');
    expect(component.filteredRows.length).toBe(2);
    expect(fixture.nativeElement.textContent).toContain('Baby Room');
    expect(fixture.nativeElement.textContent).toContain('Archived');
  });

  it('uses the manager active membership site without loading owner sites', () => {
    authStub.currentRole.and.returnValue('manager');
    authStub.activeMembership.and.returnValue({
      membership_id: 'mem-1',
      tenant_id: 'tenant-1',
      tenant_name: 'Little Sprouts',
      branch_id: 'site-1',
      branch_name: 'Oak Lane',
      role: 'manager',
    });

    fixture.detectChanges();

    const roomsReq = httpMock.expectOne((req) => req.url === '/api/v1/sites/site-1/rooms');
    expect(roomsReq.request.params.get('include_archived')).toBe('true');
    roomsReq.flush(roomsResponse);
    fixture.detectChanges();

    expect(component.selectedSiteName).toBe('Oak Lane');
    expect(httpMock.match((req) => req.url === '/api/v1/owner/site-summaries').length).toBe(0);
  });

  it('archives an active room after confirmation', () => {
    fixture.detectChanges();
    httpMock.expectOne((req) => req.url === '/api/v1/owner/site-summaries').flush(summariesResponse);
    httpMock.expectOne((req) => req.url === '/api/v1/sites/site-1/rooms').flush(roomsResponse);

    spyOn(window, 'confirm').and.returnValue(true);
    component.archiveRoom(component.rooms[0]);

    const archiveReq = httpMock.expectOne('/api/v1/sites/site-1/rooms/room-1/actions/archive');
    expect(archiveReq.request.method).toBe('POST');
    archiveReq.flush({});
    httpMock.expectOne((req) => req.url === '/api/v1/sites/site-1/rooms').flush(roomsResponse);
  });
});
