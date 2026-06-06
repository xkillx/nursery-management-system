import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { ManagerGuardiansComponent } from './manager-guardians.component';

describe('ManagerGuardiansComponent', () => {
  let fixture: ComponentFixture<ManagerGuardiansComponent>;
  let component: ManagerGuardiansComponent;
  let httpMock: HttpTestingController;

  const guardianApi = {
    id: 'guardian-1',
    full_name: 'Sarah Thompson',
    email: 'sarah@example.com',
    phone: '+44 7700 900001',
    notes: null,
    is_active: true,
    deactivated_at: null,
    deactivation_reason_code: null,
    deactivation_reason_note: null,
    created_at: '2024-08-01T00:00:00Z',
    updated_at: '2024-08-01T00:00:00Z',
  };

  const inactiveGuardianApi = {
    ...guardianApi,
    id: 'guardian-2',
    full_name: 'James Brown',
    email: null,
    phone: null,
    is_active: false,
    deactivated_at: '2025-01-15T10:00:00Z',
    deactivation_reason_code: 'moved_away',
    deactivation_reason_note: 'Family relocated',
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ManagerGuardiansComponent],
      providers: [provideHttpClient(), provideHttpClientTesting(), provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(ManagerGuardiansComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  function flushGuardians(items: any[] = []): void {
    const req = httpMock.expectOne(
      (r) => r.url === '/api/v1/guardians' && r.params.get('status') === 'active',
    );
    req.flush({ items });
  }

  it('loads active guardians on init', () => {
    fixture.detectChanges();
    flushGuardians([guardianApi]);
    fixture.detectChanges();

    expect(component.guardians.length).toBe(1);
    expect(component.guardians[0].fullName).toBe('Sarah Thompson');
    expect(component.isLoading).toBe(false);
  });

  it('defaults status to active', () => {
    expect(component.status).toBe('active');
  });

  it('status change resets offset and reloads', () => {
    fixture.detectChanges();
    flushGuardians();

    component.offset = 20;
    component.onStatusChange('all');

    expect(component.offset).toBe(0);
    expect(component.status).toBe('all');

    const req = httpMock.expectOne(
      (r) => r.url === '/api/v1/guardians' && r.params.get('status') === 'all',
    );
    req.flush({ items: [guardianApi, inactiveGuardianApi] });

    expect(component.guardians.length).toBe(2);
  });

  it('maps status filter values to labels', () => {
    expect(component.statusLabel('active')).toBe('Active');
    expect(component.statusLabel('inactive')).toBe('Inactive');
    expect(component.statusLabel('all')).toBe('All');
  });

  it('maps nullable contact fields correctly', () => {
    fixture.detectChanges();
    flushGuardians([inactiveGuardianApi]);
    fixture.detectChanges();

    const guardian = component.guardians[0];
    expect(guardian.email).toBeNull();
    expect(guardian.phone).toBeNull();
    expect(guardian.isActive).toBe(false);
    expect(guardian.deactivatedAt).toBe('2025-01-15T10:00:00Z');
  });

  it('inactive guardian shows deactivated fields', () => {
    fixture.detectChanges();
    flushGuardians([inactiveGuardianApi]);
    fixture.detectChanges();

    const guardian = component.guardians[0];
    expect(guardian.deactivationReasonCode).toBe('moved_away');
    expect(guardian.deactivationReasonNote).toBe('Family relocated');
  });
});
