import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, provideRouter, Router } from '@angular/router';

import { AuthService } from '../../../../core/services/auth.service';
import { OwnerRoomFormComponent } from './owner-room-form.component';

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

const roomResponse = {
  id: 'room-1',
  name: 'Baby Room',
  description: 'Calm baby space',
  age_group: 'baby',
  capacity: 12,
  is_active: true,
  created_at: '2026-06-01T00:00:00Z',
  updated_at: '2026-06-01T00:00:00Z',
};

describe('OwnerRoomFormComponent', () => {
  let component: OwnerRoomFormComponent;
  let fixture: ComponentFixture<OwnerRoomFormComponent>;
  let httpMock: HttpTestingController;
  let authStub: jasmine.SpyObj<AuthService>;

  beforeEach(async () => {
    authStub = jasmine.createSpyObj<AuthService>('AuthService', ['currentRole', 'activeMembership']);
    authStub.currentRole.and.returnValue('owner');
    authStub.activeMembership.and.returnValue(null);

    await TestBed.configureTestingModule({
      imports: [OwnerRoomFormComponent],
      providers: [
        provideRouter([]),
        provideHttpClient(),
        provideHttpClientTesting(),
        { provide: AuthService, useValue: authStub },
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: {
              paramMap: { get: () => null },
              queryParamMap: { get: (key: string) => key === 'site_id' ? 'site-1' : null },
            },
          },
        },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(OwnerRoomFormComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('loads owner sites for create mode and posts room payload', () => {
    fixture.detectChanges();
    httpMock.expectOne((req) => req.url === '/api/v1/owner/site-summaries').flush(summariesResponse);
    fixture.detectChanges();

    const router = TestBed.inject(Router);
    spyOn(router, 'navigate');

    component.model = {
      name: 'Baby Room',
      ageGroup: 'baby',
      capacity: 12,
      description: 'Calm baby space',
    };

    const form = fixture.nativeElement.querySelector('form');
    form.dispatchEvent(new Event('submit'));

    const createReq = httpMock.expectOne('/api/v1/sites/site-1/rooms');
    expect(createReq.request.method).toBe('POST');
    expect(createReq.request.body).toEqual({
      name: 'Baby Room',
      age_group: 'baby',
      capacity: 12,
      description: 'Calm baby space',
    });
    createReq.flush(roomResponse);

    expect(router.navigate).toHaveBeenCalledWith(['/owner/rooms'], { queryParams: { site_id: 'site-1' } });
  });

  it('loads an active room in edit mode and patches changes', async () => {
    TestBed.resetTestingModule();
    await TestBed.configureTestingModule({
      imports: [OwnerRoomFormComponent],
      providers: [
        provideRouter([]),
        provideHttpClient(),
        provideHttpClientTesting(),
        { provide: AuthService, useValue: authStub },
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: {
              paramMap: { get: (key: string) => key === 'roomId' ? 'room-1' : null },
              queryParamMap: { get: (key: string) => key === 'site_id' ? 'site-1' : null },
            },
          },
        },
      ],
    }).compileComponents();

    const editFixture = TestBed.createComponent(OwnerRoomFormComponent);
    const editHttp = TestBed.inject(HttpTestingController);
    editFixture.detectChanges();
    editHttp.expectOne((req) => req.url === '/api/v1/owner/site-summaries').flush(summariesResponse);
    editHttp.expectOne('/api/v1/sites/site-1/rooms/room-1').flush(roomResponse);

    const editComponent = editFixture.componentInstance;
    editComponent.model.capacity = 14;
    const router = TestBed.inject(Router);
    spyOn(router, 'navigate');

    editFixture.nativeElement.querySelector('form').dispatchEvent(new Event('submit'));

    const patchReq = editHttp.expectOne('/api/v1/sites/site-1/rooms/room-1');
    expect(patchReq.request.method).toBe('PATCH');
    expect(patchReq.request.body.capacity).toBe(14);
    patchReq.flush({ ...roomResponse, capacity: 14 });

    expect(router.navigate).toHaveBeenCalledWith(['/owner/rooms'], { queryParams: { site_id: 'site-1' } });
    editHttp.verify();
  });

  it('uses manager active membership for create mode', () => {
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

    expect(component.selectedSiteId).toBe('site-1');
    expect(component.selectedSiteName).toBe('Oak Lane');
    expect(httpMock.match((req) => req.url === '/api/v1/owner/site-summaries').length).toBe(0);
  });
});
