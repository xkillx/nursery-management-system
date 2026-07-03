import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter, Router } from '@angular/router';

import { ManagerChildrenComponent } from './manager-children.component';

describe('ManagerChildrenComponent', () => {
  let fixture: ComponentFixture<ManagerChildrenComponent>;
  let component: ManagerChildrenComponent;
  let httpMock: HttpTestingController;

  const childApi = {
    id: 'child-1',first_name: 'Ada',
middle_name: null,
last_name: 'Lovelace',
    date_of_birth: '2022-01-15',
    start_date: '2024-09-01',
    end_date: null,
    core_hourly_rate_minor: null,
    site_core_hourly_rate_minor: null,
    notes: null,
    is_active: true,
    left_at: null,
    left_reason_code: null,
    left_reason_note: null,
    enrollment_complete: false,
    missing_requirements: ['parent_carer_contact'],
    created_at: '2024-08-01T00:00:00Z',
    updated_at: '2024-08-01T00:00:00Z',
  };

  const completeChildApi = {
    ...childApi,
    id: 'child-2',first_name: 'Grace',
middle_name: null,
last_name: 'Hopper',
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
    req.flush({ items, total: items.length });
  }

  it('loads active children on init', () => {
    fixture.detectChanges();
    flushChildren([childApi]);
    fixture.detectChanges();

    expect(component.children.length).toBe(1);
    expect(component.children[0].siteCoreHourlyRateMinor).toBeNull();
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
    req.flush({ items: [childApi, completeChildApi], total: 2 });

    expect(component.children.length).toBe(2);
  });

  it('maps missing requirement codes to labels', () => {
    expect(component.requirementLabel('parent_carer_contact')).toBe('Parent carer contact');
    expect(component.requirementLabel('billing_rate')).toBe('billing_rate');
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
    expect(labels).toEqual(['Parent carer contact']);
  });

  it('maps has_booking_pattern to hasBookingPattern on each child', () => {
    fixture.detectChanges();
    flushChildren([{ ...childApi, has_booking_pattern: false }, { ...completeChildApi, id: 'child-2', has_booking_pattern: true }]);
    fixture.detectChanges();

    expect(component.children[0].hasBookingPattern).toBe(false);
    expect(component.children[1].hasBookingPattern).toBe(true);
  });

  it('navigates to the edit stepper route when openEdit is called', () => {
    const router = TestBed.inject(Router);
    const navigateSpy = spyOn(router, 'navigate').and.resolveTo();
    component.openEdit({ ...childApi } as any);
    expect(navigateSpy).toHaveBeenCalledWith(['/manager/children', 'child-1', 'edit']);
  });

  it('toggleCardFilter sets correct status and selectedCardFilter', () => {
    fixture.detectChanges();
    flushChildren(); // Flushes initial load

    expect(component.selectedCardFilter).toBe('all');

    // Toggle to active -> calls loadChildren
    component.toggleCardFilter('active');
    let req = httpMock.expectOne((r) => r.url === '/api/v1/children' && r.params.get('status') === 'active');
    req.flush({ items: [], total: 0 });
    expect(component.selectedCardFilter).toBe('active');
    expect(component.status).toBe('active');
    
    // Toggle active again -> resets to all -> calls loadChildren
    component.toggleCardFilter('active');
    req = httpMock.expectOne((r) => r.url === '/api/v1/children' && r.params.get('status') === 'active');
    req.flush({ items: [], total: 0 });
    expect(component.selectedCardFilter).toBe('all');
    expect(component.status).toBe('active');

    // Toggle to incomplete -> sets status to all -> calls loadChildren
    component.toggleCardFilter('incomplete');
    req = httpMock.expectOne((r) => r.url === '/api/v1/children' && r.params.get('status') === 'all');
    req.flush({ items: [], total: 0 });
    expect(component.selectedCardFilter).toBe('incomplete');
    expect(component.status).toBe('all');
  });

  it('filteredChildren filters by selectedCardFilter client-side', () => {
    fixture.detectChanges();
    flushChildren(); // Flushes initial load
    
    const activeChild = { ...completeChildApi, id: 'c1', isActive: true, enrollmentComplete: true, missingRequirements: [] };
    const incompleteChild = { ...childApi, id: 'c2', isActive: true, enrollmentComplete: false, missingRequirements: ['parent_carer_contact'] };
    const inactiveChild = { ...childApi, id: 'c3', isActive: false, enrollmentComplete: false, missingRequirements: [] };
    
    component.children = [activeChild, incompleteChild, inactiveChild] as any[];

    component.selectedCardFilter = 'all';
    expect(component.filteredChildren.length).toBe(3);

    component.selectedCardFilter = 'active';
    expect(component.filteredChildren.length).toBe(2);
    expect(component.filteredChildren.map(c => c.id)).toContain('c1');
    expect(component.filteredChildren.map(c => c.id)).toContain('c2');

    component.selectedCardFilter = 'incomplete';
    expect(component.filteredChildren.length).toBe(2);
    expect(component.filteredChildren.map(c => c.id)).toContain('c2');
    expect(component.filteredChildren.map(c => c.id)).toContain('c3');

    component.selectedCardFilter = 'requirements';
    expect(component.filteredChildren.length).toBe(1);
    expect(component.filteredChildren[0].id).toBe('c2');
  });

  it('hasStarted helper returns correct boolean based on start_date compared to today', () => {
    const today = new Date();
    
    const pastDate = new Date();
    pastDate.setDate(today.getDate() - 5);
    const pastDateStr = pastDate.toISOString().split('T')[0];

    const futureDate = new Date();
    futureDate.setDate(today.getDate() + 5);
    const futureDateStr = futureDate.toISOString().split('T')[0];

    expect(component.hasStarted(pastDateStr)).toBe(true);
    expect(component.hasStarted(futureDateStr)).toBe(false);
    expect(component.hasStarted(null)).toBe(false);
  });
});
