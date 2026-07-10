import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { AuthService } from '../../../../core/services/auth.service';
import { ManagerRoomsComponent } from './manager-rooms.component';

describe('ManagerRoomsComponent', () => {
  let component: ManagerRoomsComponent;
  let fixture: ComponentFixture<ManagerRoomsComponent>;
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

  const roomsResponse = {
    items: [
      {
        id: 'room-baby',
        name: 'Baby Room',
        description: 'Calm baby space',
        age_group: 'baby',
        capacity: 10,
        is_active: true,
        assigned_count: 6,
        is_over_capacity: false,
        created_at: '2026-06-01T00:00:00Z',
        updated_at: '2026-06-01T00:00:00Z',
      },
      {
        id: 'room-sunshine',
        name: 'Sunshine Room',
        description: null,
        age_group: 'mixed',
        capacity: 12,
        is_active: true,
        assigned_count: 14,
        is_over_capacity: true,
        created_at: '2026-06-01T00:00:00Z',
        updated_at: '2026-06-01T00:00:00Z',
      },
      {
        id: 'room-empty',
        name: 'Empty Room',
        description: null,
        age_group: 'toddler',
        capacity: 8,
        is_active: true,
        assigned_count: 0,
        is_over_capacity: false,
        created_at: '2026-06-01T00:00:00Z',
        updated_at: '2026-06-01T00:00:00Z',
      },
    ],
  };

  beforeEach(async () => {
    authStub = jasmine.createSpyObj<AuthService>('AuthService', ['activeMembership', 'currentRole']);
    authStub.activeMembership.and.returnValue(membership);
    authStub.currentRole.and.returnValue('manager');

    await TestBed.configureTestingModule({
      imports: [ManagerRoomsComponent],
      providers: [
        provideRouter([]),
        provideHttpClient(),
        provideHttpClientTesting(),
        { provide: AuthService, useValue: authStub },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ManagerRoomsComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  function flushRooms() {
    const req = httpMock.expectOne((r) => r.url === '/api/v1/sites/site-1/rooms');
    expect(req.request.params.get('include_archived')).toBe('true');
    expect(req.request.params.get('include')).toBe('occupancy');
    req.flush(roomsResponse);
    fixture.detectChanges();
  }

  it('loads rooms with the occupancy flag and renders real counts', () => {
    fixture.detectChanges();
    flushRooms();

    expect(component.selectedSiteName).toBe('Oak Lane');
    expect(component.rooms.length).toBe(3);
    expect(component.filteredRows[0].room.id).toBe('room-sunshine');
    expect(component.filteredRows[0].occupancy.current).toBe(14);
    expect(component.filteredRows[0].occupancy.percent).toBe(117);
  });

  it('sorts rows by occupancy descending, ties by name', () => {
    fixture.detectChanges();
    flushRooms();

    const names = component.statusFilteredRows.map((row) => row.room.name);
    expect(names).toEqual(['Sunshine Room', 'Baby Room', 'Empty Room']);
  });

  it('surfaces an archive error with the assigned count message', () => {
    fixture.detectChanges();
    flushRooms();

    spyOn(window, 'confirm').and.returnValue(true);
    component.archiveRoom(component.rooms.find((r) => r.id === 'room-baby')!);

    const req = httpMock.expectOne('/api/v1/sites/site-1/rooms/room-baby/actions/archive');
    expect(req.request.method).toBe('POST');
    req.flush(
      {
        code: 'room_has_children',
        message: 'Room has 3 active children assigned',
        request_id: 'req-1',
        details: { assigned_count: 3 },
      },
      { status: 409, statusText: 'Conflict' },
    );
    fixture.detectChanges();

    expect(component.pageError).toBe('3 children are still assigned — reassign them before archiving.');
  });

  it('shows an empty state when the rooms list is empty', () => {
    fixture.detectChanges();
    httpMock.expectOne((r) => r.url === '/api/v1/sites/site-1/rooms').flush({ items: [] });
    fixture.detectChanges();

    expect(component.filteredRows.length).toBe(0);
    expect(fixture.nativeElement.textContent).toContain('No rooms found');
  });
});
