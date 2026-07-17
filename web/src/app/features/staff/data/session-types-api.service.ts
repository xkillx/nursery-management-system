import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { apiUrl } from '../../../core/config/api.config';

export type SessionTypeKind = 'standard' | 'wraparound_before' | 'wraparound_after' | 'core' | 'extended';

export interface StaffSessionType {
  id: string;
  name: string;
  startTime: string;
  endTime: string;
  isActive: boolean;
  kind: SessionTypeKind;
  createdAt: string;
  updatedAt: string;
}

export interface StaffSessionTypeInput {
  name: string;
  start_time: string;
  end_time: string;
  kind?: SessionTypeKind;
}

interface ApiSessionType {
  id: string;
  name: string;
  start_time: string;
  end_time: string;
  is_active: boolean;
  kind?: string;
  created_at: string;
  updated_at: string;
}

interface ApiListResponse {
  items: ApiSessionType[];
  total: number;
  page: number;
  page_size: number;
}

@Injectable({ providedIn: 'root' })
export class StaffSessionTypesApiService {
  private readonly http = inject(HttpClient);

  listSessionTypes(
    siteId: string,
    options: { includeArchived?: boolean } = {},
  ): Observable<StaffSessionType[]> {
    let params = new HttpParams();
    if (options.includeArchived) {
      params = params.set('include_archived', 'true');
    }
    return this.http
      .get<ApiListResponse>(apiUrl(`/sites/${siteId}/session-types`), { params })
      .pipe(map((res) => res.items.map((s) => this.toSessionType(s))));
  }

  createSessionType(siteId: string, payload: StaffSessionTypeInput): Observable<StaffSessionType> {
    return this.http
      .post<ApiSessionType>(apiUrl(`/sites/${siteId}/session-types`), payload)
      .pipe(map((s) => this.toSessionType(s)));
  }

  updateSessionType(
    siteId: string,
    sessionTypeId: string,
    payload: Partial<StaffSessionTypeInput>,
  ): Observable<StaffSessionType> {
    return this.http
      .patch<ApiSessionType>(apiUrl(`/sites/${siteId}/session-types/${sessionTypeId}`), payload)
      .pipe(map((s) => this.toSessionType(s)));
  }

  archiveSessionType(siteId: string, sessionTypeId: string): Observable<void> {
    return this.http.post<void>(
      apiUrl(`/sites/${siteId}/session-types/${sessionTypeId}/actions/archive`),
      {},
    );
  }

  reactivateSessionType(siteId: string, sessionTypeId: string): Observable<StaffSessionType> {
    return this.http
      .post<ApiSessionType>(apiUrl(`/sites/${siteId}/session-types/${sessionTypeId}/actions/activate`), {})
      .pipe(map((s) => this.toSessionType(s)));
  }

  private toSessionType(s: ApiSessionType): StaffSessionType {
    return {
      id: s.id,
      name: s.name,
      startTime: s.start_time,
      endTime: s.end_time,
      isActive: s.is_active,
      kind: (s.kind ?? 'standard') as SessionTypeKind,
      createdAt: s.created_at,
      updatedAt: s.updated_at,
    };
  }
}
