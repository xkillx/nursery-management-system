import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideLocationMocks } from '@angular/common/testing';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { FundingOverviewRecord } from '../../models/funding.models';
import { ManagerFundingOverviewComponent } from './manager-funding-overview.component';

describe('ManagerFundingOverviewComponent', () => {
  let component: ManagerFundingOverviewComponent;
  let httpMock: HttpTestingController;

  const emptyOverview: FundingOverviewRecord = {
    billingMonth: '2026-06',
    summary: {
      includedChildCount: 0,
      flaggedChildCount: 0,
      missingProfileCount: 0,
      explicitZeroCount: 0,
      underOneHourCount: 0,
      above160HoursCount: 0,
    },
    items: [],
  };

  function createComponent(): void {
    const fixture = TestBed.createComponent(ManagerFundingOverviewComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
    fixture.detectChanges();
  }

  function flushInitialOverview(data: FundingOverviewRecord = emptyOverview): void {
    const req = httpMock.expectOne((r) => r.url === '/api/v1/funding/overview');
    req.flush(data);
  }

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        provideLocationMocks(),
      ],
    });
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('loads overview on init with current month', () => {
    createComponent();
    const req = httpMock.expectOne((r) => r.url === '/api/v1/funding/overview');
    expect(req.request.params.get('billing_month')).toMatch(/^\d{4}-\d{2}$/);
    req.flush(emptyOverview);
  });

  it('reloads overview when month changes', () => {
    createComponent();
    flushInitialOverview();

    component.onMonthChange('2025-12');
    const req = httpMock.expectOne((r) => r.url === '/api/v1/funding/overview');
    expect(req.request.params.get('billing_month')).toBe('2025-12');
    req.flush(emptyOverview);
  });

  it('stores overview after successful load', () => {
    createComponent();
    const apiResponse = {
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
          photo_url: null,
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
          photo_url: null,
          flags: ['explicit_zero_allowance'],
        },
      ],
    };
    const req = httpMock.expectOne((r) => r.url === '/api/v1/funding/overview');
    req.flush(apiResponse);
    expect(component.overview!.billingMonth).toBe('2026-06');
    expect(component.overview!.summary.includedChildCount).toBe(3);
    expect(component.overview!.items.length).toBe(2);
    expect(component.overview!.items[0].childName).toBe('Alice');
    expect(component.overview!.items[1].fundedAllowanceMinutes).toBe(0);
  });

  it('sets errorMessage on API error', () => {
    createComponent();
    const req = httpMock.expectOne((r) => r.url === '/api/v1/funding/overview');
    req.flush({ code: 'validation_error', message: 'Invalid billing month.' }, { status: 400, statusText: 'Bad Request' });
    expect(component.errorMessage).toContain('Invalid billing month');
    expect(component.overview).toBeNull();
  });

  it('generates correct review link and query params', () => {
    createComponent();
    flushInitialOverview();
    component.selectedBillingMonth = '2026-06';
    expect(component.reviewLink('child-1')).toEqual(['/manager/children', 'child-1']);
    expect(component.reviewQueryParams()).toEqual({ billing_month: '2026-06' });
  });

  it('formatAllowance handles null, zero, and normal values', () => {
    createComponent();
    flushInitialOverview();
    const fmt = component.formatAllowance;
    expect(fmt(null)).toBe('Not set');
    expect(fmt(0)).toBe('0h 0m');
    expect(fmt(30)).toBe('30m');
    expect(fmt(60)).toBe('1h');
    expect(fmt(90)).toBe('1h 30m');
    expect(fmt(9601)).toBe('160h 1m');
  });

  it('flagLabel returns human-readable labels', () => {
    createComponent();
    flushInitialOverview();
    expect(component.flagLabel('missing_profile')).toBe('Missing allowance');
    expect(component.flagLabel('explicit_zero_allowance')).toBe('Zero allowance');
    expect(component.flagLabel('under_one_hour_allowance')).toBe('Under one hour');
    expect(component.flagLabel('above_160_hours_allowance')).toBe('Above 160 hours');
  });

  it('flagColor returns correct color for each flag', () => {
    createComponent();
    flushInitialOverview();
    expect(component.flagColor('missing_profile')).toBe('danger');
    expect(component.flagColor('explicit_zero_allowance')).toBe('warning');
    expect(component.flagColor('under_one_hour_allowance')).toBe('warning');
    expect(component.flagColor('above_160_hours_allowance')).toBe('info');
  });
});
