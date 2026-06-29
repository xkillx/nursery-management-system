import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { provideIcons } from '@ng-icons/core';
import {
  heroArrowRightCircle,
  heroBanknotes,
  heroBuildingOffice2,
  heroChatBubbleLeftRight,
  heroCheckCircle,
  heroChevronRight,
  heroClock,
  heroCog6Tooth,
  heroReceiptPercent,
  heroRectangleStack,
  heroScale,
  heroShieldCheck,
  heroUserGroup,
} from '@ng-icons/heroicons/outline';

import { apiUrl } from '../../../../core/config/api.config';
import { AuthService } from '../../../../core/services/auth.service';
import { ManagerSiteSettingsComponent } from './manager-site-settings.component';

describe('ManagerSiteSettingsComponent', () => {
  let component: ManagerSiteSettingsComponent;
  let fixture: ComponentFixture<ManagerSiteSettingsComponent>;
  let httpMock: HttpTestingController;
  let authStub: jasmine.SpyObj<AuthService>;

  const membership = {
    membership_id: 'mem-1',
    tenant_id: 'tenant-1',
    tenant_name: 'Little Sprouts',
    branch_id: 'site-1',
    branch_name: 'Oak Lane',
    role: 'manager' as const,
  };

  beforeEach(async () => {
    authStub = jasmine.createSpyObj<AuthService>('AuthService', ['activeMembership', 'currentRole']);
    authStub.activeMembership.and.returnValue(membership);
    authStub.currentRole.and.returnValue('manager');

    await TestBed.configureTestingModule({
      imports: [ManagerSiteSettingsComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: AuthService, useValue: authStub },
        provideIcons({
          heroArrowRightCircle,
          heroBanknotes,
          heroBuildingOffice2,
          heroChatBubbleLeftRight,
          heroCheckCircle,
          heroChevronRight,
          heroClock,
          heroCog6Tooth,
          heroReceiptPercent,
          heroRectangleStack,
          heroScale,
          heroShieldCheck,
          heroUserGroup,
        }),
      ],
    }).compileComponents();

    httpMock = TestBed.inject(HttpTestingController);
    fixture = TestBed.createComponent(ManagerSiteSettingsComponent);
    component = fixture.componentInstance;
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('renders the ten setting cards with real rooms, billing, and session types data', () => {
    fixture.detectChanges();
    const roomsReq = httpMock.expectOne((req) => req.url === apiUrl('/sites/site-1/rooms'));
    roomsReq.flush({
      rooms: [
        {
          id: 'room-1',
          name: 'Baby Room',
          description: null,
          age_group: 'baby',
          capacity: 10,
          is_active: true,
          assigned_count: 4,
          is_over_capacity: false,
          created_at: '2026-06-01T00:00:00Z',
          updated_at: '2026-06-01T00:00:00Z',
        },
        {
          id: 'room-2',
          name: 'Toddler Room',
          description: null,
          age_group: 'toddler',
          capacity: 12,
          is_active: true,
          assigned_count: 9,
          is_over_capacity: false,
          created_at: '2026-06-01T00:00:00Z',
          updated_at: '2026-06-01T00:00:00Z',
        },
        {
          id: 'room-3',
          name: 'Preschool',
          description: null,
          age_group: 'preschool',
          capacity: 16,
          is_active: true,
          assigned_count: 14,
          is_over_capacity: false,
          created_at: '2026-06-01T00:00:00Z',
          updated_at: '2026-06-01T00:00:00Z',
        },
      ],
    });
    const billingReq = httpMock.expectOne(apiUrl('/billing-setup'));
    billingReq.flush({ core_hourly_rate_minor: 700, has_rate: true });

    const sessionTypesReq = httpMock.expectOne((req) => req.url === apiUrl('/sites/site-1/session-types'));
    sessionTypesReq.flush({
      session_types: [
        { id: 'st-1', name: 'Morning', start_time: '08:00', end_time: '12:00', is_active: true, created_at: '', updated_at: '' },
        { id: 'st-2', name: 'Afternoon', start_time: '13:00', end_time: '17:00', is_active: true, created_at: '', updated_at: '' },
      ],
    });

    fixture.detectChanges();

    expect(component.cards().length).toBe(10);
    const roomCard = component.cards().find((c) => c.id === 'rooms') as { detail: string; statusLabel: string };
    expect(roomCard.detail).toContain('Baby');
    expect(roomCard.statusLabel).toBe('3 active');

    const billingCard = component.cards().find((c) => c.id === 'billing');
    expect(billingCard?.statusLabel).toBe('Auto-invoicing: ON');
  });

  it('falls back to a friendly error when the rooms API rejects', () => {
    fixture.detectChanges();
    httpMock.expectOne(apiUrl('/billing-setup')).flush({ core_hourly_rate_minor: 0, has_rate: false });
    httpMock.expectOne(apiUrl('/sites/site-1/rooms')).error(new ProgressEvent('error'));
    httpMock.expectOne((req) => req.url === apiUrl('/sites/site-1/session-types')).flush({ session_types: [] });
    fixture.detectChanges();

    expect(component.pageError()).toContain('Failed to load');
  });
});
