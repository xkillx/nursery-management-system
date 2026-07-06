import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { apiUrl } from '../../../core/config/api.config';

export interface StaffRoom {
  id: string;
  name: string;
  description: string | null;
  ageGroup: string;
  capacity: number;
  isActive: boolean;
  assignedCount?: number;
  isOverCapacity?: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface StaffListRoomsOptions {
  includeArchived?: boolean;
  includeOccupancy?: boolean;
}

interface ApiRoom {
  id: string;
  name: string;
  description: string | null;
  age_group: string;
  capacity: number;
  is_active: boolean;
  assigned_count?: number;
  is_over_capacity?: boolean;
  created_at: string;
  updated_at: string;
}

interface ApiRoomListResponse {
  items: ApiRoom[];
  total: number;
  page: number;
  page_size: number;
}

@Injectable({ providedIn: 'root' })
export class StaffRoomsApiService {
  private readonly http = inject(HttpClient);

  listRooms(siteId: string, options: StaffListRoomsOptions = {}): Observable<StaffRoom[]> {
    let params = new HttpParams();
    if (options.includeArchived) {
      params = params.set('include_archived', 'true');
    }
    if (options.includeOccupancy) {
      params = params.set('include', 'occupancy');
    }
    return this.http
      .get<ApiRoomListResponse>(apiUrl(`/sites/${siteId}/rooms`), { params })
      .pipe(map((res) => res.items.map((room) => this.toRoom(room))));
  }

  archiveRoom(siteId: string, roomId: string): Observable<void> {
    return this.http.post<void>(apiUrl(`/sites/${siteId}/rooms/${roomId}/actions/archive`), {});
  }

  reactivateRoom(siteId: string, roomId: string): Observable<StaffRoom> {
    return this.http
      .post<ApiRoom>(apiUrl(`/sites/${siteId}/rooms/${roomId}/actions/activate`), {})
      .pipe(map((room) => this.toRoom(room)));
  }

  private toRoom(room: ApiRoom): StaffRoom {
    return {
      id: room.id,
      name: room.name,
      description: room.description,
      ageGroup: room.age_group,
      capacity: room.capacity,
      isActive: room.is_active,
      assignedCount: room.assigned_count,
      isOverCapacity: room.is_over_capacity,
      createdAt: room.created_at,
      updatedAt: room.updated_at,
    };
  }
}
