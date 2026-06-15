import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { OwnerRoomsComponent } from './owner-rooms.component';
import { AuthService } from '../../../../core/services/auth.service';

const mockRoomsApiResponse = {
  rooms: [
    { id: 'room-1', name: 'Baby Room', description: null, age_group: 'baby', capacity: 12, is_active: true, created_at: '2026-06-15T10:00:00Z', updated_at: '2026-06-15T10:00:00Z' },
    { id: 'room-2', name: 'Archived Room', description: null, age_group: 'toddler', capacity: 8, is_active: false, created_at: '2026-06-15T10:00:00Z', updated_at: '2026-06-15T10:00:00Z' },
  ],
};

const mockSitesResponse = {
  billing_month: '2026-06',
  attendance_local_date: '2026-06-15',
  currency_code: 'GBP',
  totals: { active_manager_count: 0, pending_manager_invite_count: 0, active_children_count: 0, checked_in_today_count: 0, incomplete_attendance_count: 0, draft_count: 0, issued_count: 0, overdue_count: 0, payment_failed_count: 0, paid_count: 0, total_issued_minor: 0, total_paid_minor: 0, outstanding_minor: 0, overdue_outstanding_minor: 0 },
  sites: [{ site_id: 'site-1', site_name: 'Oak Lane', setup_status: 'complete', active_manager_count: 0, pending_manager_invite_count: 0, active_children_count: 0, site_core_hourly_rate_minor: null, setup_issues: [], attendance: { checked_in_today_count: 0, incomplete_attendance_count: 0 }, funding_readiness: { included_child_count: 0, flagged_child_count: 0, missing_profile_count: 0, explicit_zero_count: 0, under_one_hour_count: 0, above_160_hours_count: 0 }, invoice_payment_health: { draft_count: 0, issued_count: 0, overdue_count: 0, payment_failed_count: 0, paid_count: 0, total_issued_minor: 0, total_paid_minor: 0, outstanding_minor: 0, overdue_outstanding_minor: 0, failed_payment_count: 0 } }],
};

describe('OwnerRoomsComponent', () => {
  let component: OwnerRoomsComponent;
  let fixture: ComponentFixture<OwnerRoomsComponent>;
  let httpMock: HttpTestingController;
  let authService: AuthService;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [OwnerRoomsComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(OwnerRoomsComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
    authService = TestBed.inject(AuthService);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('creates the component', () => {
    expect(component).toBeTruthy();
  });

  it('loads sites for owner role', () => {
    Object.defineProperty(authService, 'currentRole', { get: () => () => 'owner' });

    fixture.detectChanges();

    const req = httpMock.expectOne('/api/v1/owner/site-summaries');
    expect(req.request.method).toBe('GET');
    req.flush(mockSitesResponse);

    expect(component.sites.length).toBe(1);
    expect(component.isOwner).toBeTrue();
  });

  it('loads rooms when site is selected', () => {
    Object.defineProperty(authService, 'currentRole', { get: () => () => 'owner' });

    fixture.detectChanges();
    httpMock.expectOne('/api/v1/owner/site-summaries').flush(mockSitesResponse);
    fixture.detectChanges();

    component.selectedSiteId = 'site-1';
    component['loadRooms']();

    const roomsReq = httpMock.expectOne((r) => r.url === '/api/v1/sites/site-1/rooms');
    expect(roomsReq.request.method).toBe('GET');
    roomsReq.flush(mockRoomsApiResponse);

    expect(component.rooms.length).toBe(2);
  });

  it('toggles archived rooms visibility', () => {
    component.rooms = mockRoomsApiResponse.rooms.map((r: any) => ({
      id: r.id,
      name: r.name,
      description: r.description,
      ageGroup: r.age_group,
      capacity: r.capacity,
      isActive: r.is_active,
      createdAt: r.created_at,
      updatedAt: r.updated_at,
    }));

    expect(component.filteredRooms.length).toBe(1);

    component.showArchived = true;
    expect(component.filteredRooms.length).toBe(2);
  });

  it('shows empty state when no rooms', () => {
    component.rooms = [];
    component.loadingRooms = false;
    fixture.detectChanges();

    const emptyState = fixture.nativeElement.querySelector('app-empty-state');
    expect(emptyState).toBeTruthy();
  });

  it('validates form fields', () => {
    component.formName = '';
    component.formAgeGroup = '';
    component.formCapacity = null;
    expect(component.hasFormErrors).toBeTrue();

    component.formName = 'Baby Room';
    component.formAgeGroup = 'baby';
    component.formCapacity = 12;
    expect(component.hasFormErrors).toBeFalse();
  });
});
