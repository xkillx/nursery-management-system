import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { OwnerOverviewComponent } from './owner-overview.component';
import { OwnerApiService } from '../../data/owner-api.service';

const summariesResponse = {
  billing_month: '2026-06',
  attendance_local_date: '2026-06-11',
  currency_code: 'GBP',
  totals: {
    active_manager_count: 3,
    pending_manager_invite_count: 0,
    active_children_count: 45,
    checked_in_today_count: 30,
    incomplete_attendance_count: 2,
    draft_count: 1,
    issued_count: 8,
    overdue_count: 1,
    payment_failed_count: 0,
    paid_count: 7,
    total_issued_minor: 120000,
    total_paid_minor: 100000,
    outstanding_minor: 20000,
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
      site_core_hourly_rate_minor: 750,
      setup_issues: [],
      attendance: { checked_in_today_count: 12, incomplete_attendance_count: 0 },
      funding_readiness: { included_child_count: 14, flagged_child_count: 0, missing_profile_count: 0, explicit_zero_count: 0, under_one_hour_count: 0, above_160_hours_count: 0 },
      invoice_payment_health: { draft_count: 0, issued_count: 3, overdue_count: 0, payment_failed_count: 0, paid_count: 3, total_issued_minor: 45000, total_paid_minor: 45000, outstanding_minor: 0, overdue_outstanding_minor: 0, failed_payment_count: 0 },
    },
    {
      site_id: 'site-2',
      site_name: 'Elm Street',
      setup_status: 'incomplete_setup',
      active_manager_count: 0,
      pending_manager_invite_count: 1,
      active_children_count: 20,
      site_core_hourly_rate_minor: null,
      setup_issues: ['missing_site_core_hourly_rate'],
      attendance: { checked_in_today_count: 15, incomplete_attendance_count: 2 },
      funding_readiness: { included_child_count: 18, flagged_child_count: 1, missing_profile_count: 0, explicit_zero_count: 0, under_one_hour_count: 0, above_160_hours_count: 0 },
      invoice_payment_health: { draft_count: 1, issued_count: 4, overdue_count: 1, payment_failed_count: 0, paid_count: 3, total_issued_minor: 60000, total_paid_minor: 45000, outstanding_minor: 15000, overdue_outstanding_minor: 5000, failed_payment_count: 0 },
    },
  ],
};

describe('OwnerOverviewComponent', () => {
  let fixture: ComponentFixture<OwnerOverviewComponent>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [OwnerOverviewComponent],
      providers: [provideRouter([]), provideHttpClient(), provideHttpClientTesting()],
    }).compileComponents();

    fixture = TestBed.createComponent(OwnerOverviewComponent);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('loads all-site summaries on init and renders totals', () => {
    fixture.detectChanges();

    const req = httpMock.expectOne((r) => r.url === '/api/v1/owner/site-summaries');
    expect(req.request.method).toBe('GET');
    req.flush(summariesResponse);

    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('45');
    expect(text).toContain('Oak Lane');
    expect(text).toContain('Elm Street');
  });

  it('emphasizes exception sites before healthy sites', () => {
    fixture.detectChanges();
    httpMock.expectOne((r) => r.url === '/api/v1/owner/site-summaries').flush(summariesResponse);
    fixture.detectChanges();

    const cards = fixture.nativeElement.querySelectorAll('[data-testid^="owner-site-card-"]');
    expect(cards.length).toBe(2);
    expect(cards[0].getAttribute('data-testid')).toBe('owner-site-card-site-2');
    expect(cards[0].textContent).toContain('No active manager');
  });

  it('displays site core hourly rate for sites with a rate', () => {
    fixture.detectChanges();
    httpMock.expectOne((r) => r.url === '/api/v1/owner/site-summaries').flush(summariesResponse);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('£7.50/hr');
  });

  it('shows missing-rate badge for sites without a rate', () => {
    fixture.detectChanges();
    httpMock.expectOne((r) => r.url === '/api/v1/owner/site-summaries').flush(summariesResponse);
    fixture.detectChanges();

    const card = fixture.nativeElement.querySelector('[data-testid="owner-site-card-site-2"]');
    expect(card.textContent).toContain('No rate set');
  });

  it('shows rate edit UI when editing rate', () => {
    fixture.detectChanges();
    httpMock.expectOne((r) => r.url === '/api/v1/owner/site-summaries').flush(summariesResponse);
    fixture.detectChanges();

    const component = fixture.componentInstance;
    component.startRateEdit(component.sites[0]);
    fixture.detectChanges();

    expect(component.editingRateSiteId).toBe('site-1');
    expect(component.rateEditValue).toBe('7.50');
  });

  it('cancels rate edit', () => {
    const component = fixture.componentInstance;
    component.startRateEdit({ siteId: 'site-1', siteCoreHourlyRateMinor: 750 } as any);
    component.cancelRateEdit();

    expect(component.editingRateSiteId).toBeNull();
    expect(component.rateEditValue).toBe('');
  });

  it('validates rate must be positive', () => {
    const component = fixture.componentInstance;
    component.rateEditValue = '0';
    component.saveSiteRate('site-1');

    expect(component.rateEditError).toContain('positive');
    component.rateEditValue = '-5';
    component.saveSiteRate('site-1');

    expect(component.rateEditError).toContain('positive');
  });

  it('saves rate and reloads summaries', () => {
    fixture.detectChanges();
    httpMock.expectOne((r) => r.url === '/api/v1/owner/site-summaries').flush(summariesResponse);
    fixture.detectChanges();

    const component = fixture.componentInstance;
    component.rateEditValue = '8.50';
    component.saveSiteRate('site-1');

    const putReq = httpMock.expectOne((r) => r.url === '/api/v1/owner/sites/site-1/billing-setup');
    expect(putReq.request.method).toBe('PUT');
    expect(putReq.request.body).toEqual({ core_hourly_rate_minor: 850 });
    putReq.flush({});

    const reloadReq = httpMock.expectOne((r) => r.url === '/api/v1/owner/site-summaries');
    expect(reloadReq.request.method).toBe('GET');
    reloadReq.flush(summariesResponse);
  });

  it('calls API with site_id when site focus applied', () => {
    fixture.detectChanges();
    httpMock.expectOne((r) => r.url === '/api/v1/owner/site-summaries').flush(summariesResponse);
    fixture.detectChanges();

    const component = fixture.componentInstance;
    component.onSiteFocus('site-1');
    fixture.detectChanges();

    const req = httpMock.expectOne((r) => r.url === '/api/v1/owner/site-summaries');
    expect(req.request.params.get('site_id')).toBe('site-1');
    req.flush(summariesResponse);
  });

  it('shows validation error for invalid billing month', () => {
    const component = fixture.componentInstance;
    component.billingMonthControl = 'invalid';
    component.onBillingMonthChange();

    expect(component.billingMonthError).toBe('Enter a valid month in YYYY-MM format.');
  });

  it('does not call API for invalid billing month', () => {
    fixture.detectChanges();
    httpMock.expectOne((r) => r.url === '/api/v1/owner/site-summaries').flush(summariesResponse);

    const component = fixture.componentInstance;
    component.billingMonthControl = 'bad';
    component.onBillingMonthChange();
    fixture.detectChanges();

    httpMock.expectNone((r) => r.url === '/api/v1/owner/site-summaries' && r.params.get('billing_month') === 'bad');
    expect(component.billingMonthError).toBe('Enter a valid month in YYYY-MM format.');
  });

  it('shows error state on API failure', () => {
    fixture.detectChanges();
    httpMock.expectOne((r) => r.url === '/api/v1/owner/site-summaries').flush('Error', { status: 500, statusText: 'Server Error' });
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Failed to load site summaries');
  });

  it('shows empty state when no sites returned', () => {
    const emptyResponse = { ...summariesResponse, sites: [] };
    fixture.detectChanges();
    httpMock.expectOne((r) => r.url === '/api/v1/owner/site-summaries').flush(emptyResponse);
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('No active sites');
  });

  it('does not contain staff or parent links', () => {
    fixture.detectChanges();
    httpMock.expectOne((r) => r.url === '/api/v1/owner/site-summaries').flush(summariesResponse);
    fixture.detectChanges();

    const html = fixture.nativeElement.innerHTML;
    expect(html).not.toContain('/staff/');
    expect(html).not.toContain('/parent/');
    expect(html).not.toContain('/app/');
  });
});
