import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, provideRouter } from '@angular/router';

import { OwnerManagerAccessComponent } from './owner-manager-access.component';

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
      attendance: { checked_in_today_count: 12, incomplete_attendance_count: 0 },
      funding_readiness: { included_child_count: 14, flagged_child_count: 0, missing_profile_count: 0, explicit_zero_count: 0, under_one_hour_count: 0, above_160_hours_count: 0 },
      invoice_payment_health: { draft_count: 0, issued_count: 3, overdue_count: 0, payment_failed_count: 0, paid_count: 3, total_issued_minor: 45000, total_paid_minor: 45000, outstanding_minor: 0, overdue_outstanding_minor: 0, failed_payment_count: 0 },
    },
  ],
};

const managerAccessResponse = [
  { membership_id: 'mem-1', user_id: 'user-1', email: 'alice@example.com', is_active: true },
  { membership_id: 'mem-2', user_id: 'user-2', email: 'bob@example.com', is_active: false },
];

describe('OwnerManagerAccessComponent', () => {
  let fixture: ComponentFixture<OwnerManagerAccessComponent>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [OwnerManagerAccessComponent],
      providers: [
        provideRouter([]),
        provideHttpClient(),
        provideHttpClientTesting(),
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { queryParamMap: { get: () => null } } },
        },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(OwnerManagerAccessComponent);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('loads sites on init', () => {
    fixture.detectChanges();
    const req = httpMock.expectOne((r) => r.url === '/api/v1/owner/site-summaries');
    expect(req.request.method).toBe('GET');
    req.flush(summariesResponse);
    fixture.detectChanges();

    expect(fixture.componentInstance.sites.length).toBe(1);
  });

  it('lists manager access for selected site', () => {
    fixture.detectChanges();
    httpMock.expectOne((r) => r.url === '/api/v1/owner/site-summaries').flush(summariesResponse);
    fixture.detectChanges();

    const component = fixture.componentInstance;
    component.selectedSiteId = 'site-1';
    component.onSiteSelect();
    fixture.detectChanges();

    const req = httpMock.expectOne((r) => r.url === '/api/v1/owner/manager-access');
    expect(req.request.params.get('site_id')).toBe('site-1');
    expect(req.request.params.get('status')).toBe('active');
    req.flush(managerAccessResponse);
    fixture.detectChanges();

    expect(component.accessRecords.length).toBe(2);
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('alice@example.com');
    expect(text).toContain('bob@example.com');
  });

  it('honors site_id query param', async () => {
    TestBed.resetTestingModule();
    await TestBed.configureTestingModule({
      imports: [OwnerManagerAccessComponent],
      providers: [
        provideRouter([]),
        provideHttpClient(),
        provideHttpClientTesting(),
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { queryParamMap: { get: (key: string) => key === 'site_id' ? 'site-1' : null } } },
        },
      ],
    }).compileComponents();

    const qFixture = TestBed.createComponent(OwnerManagerAccessComponent);
    const qHttp = TestBed.inject(HttpTestingController);

    qFixture.detectChanges();
    qHttp.expectOne((r) => r.url === '/api/v1/owner/site-summaries').flush(summariesResponse);
    qFixture.detectChanges();

    const accessReq = qHttp.expectOne((r) => r.url === '/api/v1/owner/manager-access');
    expect(accessReq.request.params.get('site_id')).toBe('site-1');
    accessReq.flush(managerAccessResponse);

    qHttp.verify();
  });

  it('grant maps each success outcome and refreshes list', () => {
    fixture.detectChanges();
    httpMock.expectOne((r) => r.url === '/api/v1/owner/site-summaries').flush(summariesResponse);
    fixture.detectChanges();

    const component = fixture.componentInstance;
    component.selectedSiteId = 'site-1';
    component.onSiteSelect();
    httpMock.expectOne((r) => r.url === '/api/v1/owner/manager-access').flush([]);
    fixture.detectChanges();

    component.grantEmail = 'new@example.com';
    component.onGrantSubmit();

    const grantReq = httpMock.expectOne('/api/v1/owner/sites/site-1/manager-access');
    grantReq.flush({ outcome: 'manager_membership_granted', membership_id: 'mem-3', invite: null });
    fixture.detectChanges();

    expect(component.successMessage).toBe('Manager access granted.');

    const refreshReq = httpMock.expectOne((r) => r.url === '/api/v1/owner/manager-access');
    refreshReq.flush(managerAccessResponse);
    fixture.detectChanges();
  });

  it('deactivate requires confirmation and calls endpoint', () => {
    fixture.detectChanges();
    httpMock.expectOne((r) => r.url === '/api/v1/owner/site-summaries').flush(summariesResponse);
    fixture.detectChanges();

    const component = fixture.componentInstance;
    component.selectedSiteId = 'site-1';
    component.onSiteSelect();
    httpMock.expectOne((r) => r.url === '/api/v1/owner/manager-access').flush(managerAccessResponse);
    fixture.detectChanges();

    spyOn(window, 'confirm').and.returnValue(true);
    component.onDeactivate({ membershipId: 'mem-1', userId: 'user-1', email: 'alice@example.com', isActive: true });

    const req = httpMock.expectOne('/api/v1/owner/sites/site-1/manager-access/mem-1/actions/deactivate');
    expect(req.request.method).toBe('POST');
    req.flush(null, { status: 204, statusText: 'No Content' });

    expect(component.successMessage).toContain('deactivated');
    httpMock.expectOne((r) => r.url === '/api/v1/owner/manager-access');
  });

  it('inactive rows expose reactivate and active rows expose deactivate', () => {
    fixture.detectChanges();
    httpMock.expectOne((r) => r.url === '/api/v1/owner/site-summaries').flush(summariesResponse);
    fixture.detectChanges();

    const component = fixture.componentInstance;
    component.selectedSiteId = 'site-1';
    component.onSiteSelect();
    httpMock.expectOne((r) => r.url === '/api/v1/owner/manager-access').flush(managerAccessResponse);
    fixture.detectChanges();

    const buttons = fixture.nativeElement.querySelectorAll('tbody button');
    expect(buttons.length).toBe(2);
    expect(buttons[0].textContent.trim()).toBe('Deactivate');
    expect(buttons[1].textContent.trim()).toBe('Reactivate');
  });

  it('contains no practitioner or parent role selector', () => {
    fixture.detectChanges();
    httpMock.expectOne((r) => r.url === '/api/v1/owner/site-summaries').flush(summariesResponse);
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).not.toContain('Practitioner');
    expect(text).not.toContain('Parent');
  });
});
