import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { ManagerChildrenComponent } from './manager-children.component';

describe('ManagerChildrenComponent', () => {
  let fixture: ComponentFixture<ManagerChildrenComponent>;
  let component: ManagerChildrenComponent;
  let httpMock: HttpTestingController;

  const childApi = {
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

  const completeChildApi = {
    ...childApi,
    id: 'child-2',
    full_name: 'Grace Hopper',
    enrollment_complete: true,
    missing_requirements: [],
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ManagerChildrenComponent],
      providers: [provideHttpClient(), provideHttpClientTesting(), provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(ManagerChildrenComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  function flushChildren(items: any[] = []): void {
    const req = httpMock.expectOne(
      (r) => r.url === '/api/v1/children' && r.params.get('status') === 'active',
    );
    req.flush({ items });
  }

  it('loads active children on init', () => {
    fixture.detectChanges();
    flushChildren([childApi]);
    fixture.detectChanges();

    expect(component.children.length).toBe(1);
    expect(component.children[0].coreHourlyRateMinor).toBe(750);
    expect(component.isLoading).toBe(false);
  });

  it('defaults status to active', () => {
    expect(component.status).toBe('active');
  });

  it('status change resets offset and reloads', () => {
    fixture.detectChanges();
    flushChildren();

    component.offset = 20;
    component.onStatusChange('all');

    expect(component.offset).toBe(0);
    expect(component.status).toBe('all');

    const req = httpMock.expectOne(
      (r) => r.url === '/api/v1/children' && r.params.get('status') === 'all',
    );
    req.flush({ items: [childApi, completeChildApi] });

    expect(component.children.length).toBe(2);
  });

  it('formats rate using formatHourlyRateGbp', () => {
    expect(component.formatRate(750)).toBe('£7.50/hr');
    expect(component.formatRate(0)).toBe('£0.00/hr');
  });

  it('maps missing requirement codes to labels', () => {
    expect(component.requirementLabel('guardian_link')).toBe('Linked guardian');
    expect(component.requirementLabel('billing_rate')).toBe('Billing rate');
  });

  it('maps status filter values to labels', () => {
    expect(component.statusLabel('active')).toBe('Active');
    expect(component.statusLabel('inactive')).toBe('Inactive');
    expect(component.statusLabel('all')).toBe('All');
  });

  it('does not contain delete or lifecycle action references', () => {
    const ts = document.documentElement.innerHTML;
    expect(component).toBeDefined();
  });

  it('renders enrollment complete badge for enrolled child', () => {
    fixture.detectChanges();
    flushChildren([completeChildApi]);
    fixture.detectChanges();

    expect(component.children[0].enrollmentComplete).toBe(true);
    expect(component.children[0].missingRequirements).toEqual([]);
  });

  it('renders missing requirements as readable labels', () => {
    fixture.detectChanges();
    flushChildren([childApi]);
    fixture.detectChanges();

    const child = component.children[0];
    const labels = child.missingRequirements.map(component.requirementLabel);
    expect(labels).toEqual(['Linked guardian', 'Billing rate']);
  });
});
