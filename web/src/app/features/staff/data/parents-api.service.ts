import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { apiUrl } from '../../../core/config/api.config';
import {
  ParentRecord,
  ParentWithChildren,
  ParentListResponse,
  ParentStatusFilter,
} from '../models/parents.models';

@Injectable({ providedIn: 'root' })
export class ParentsApiService {
  private readonly http = inject(HttpClient);

  list(
    page = 1,
    pageSize = 20,
    status: ParentStatusFilter = 'active',
    search?: string,
  ): Observable<ParentListResponse> {
    let params = new HttpParams()
      .set('page', page.toString())
      .set('page_size', pageSize.toString());

    if (status !== 'all') {
      params = params.set('is_active', (status === 'active').toString());
    }
    if (search?.trim()) {
      params = params.set('search', search.trim());
    }

    return this.http.get<ParentListResponse>(apiUrl('/parents'), { params });
  }

  get(parentId: string): Observable<ParentWithChildren> {
    return this.http.get<ParentWithChildren>(apiUrl(`/parents/${parentId}`));
  }

  create(body: {
    first_name: string;
    last_name?: string;
    email?: string;
    phone?: string;
    address_line1?: string;
    address_line2?: string;
    address_city?: string;
    address_postcode?: string;
    relationship_to_child?: string;
    has_parental_responsibility?: boolean;
    can_pick_up?: boolean;
    is_emergency_contact?: boolean;
    notes?: string;
  }): Observable<ParentRecord> {
    return this.http.post<ParentRecord>(apiUrl('/parents'), body);
  }

  update(
    parentId: string,
    body: Partial<{
      first_name: string;
      last_name: string | null;
      email: string | null;
      phone: string | null;
      address_line1: string | null;
      address_line2: string | null;
      address_city: string | null;
      address_postcode: string | null;
      relationship_to_child: string | null;
      has_parental_responsibility: boolean;
      can_pick_up: boolean;
      is_emergency_contact: boolean;
      notes: string | null;
      is_active: boolean;
    }>,
  ): Observable<ParentRecord> {
    return this.http.put<ParentRecord>(apiUrl(`/parents/${parentId}`), body);
  }

  delete(parentId: string): Observable<{ message: string }> {
    return this.http.delete<{ message: string }>(apiUrl(`/parents/${parentId}`));
  }

  linkChild(parentId: string, childId: string): Observable<{ id: string; child_id: string }> {
    return this.http.post<{ id: string; child_id: string }>(
      apiUrl(`/parents/${parentId}/link-child`),
      { child_id: childId },
    );
  }

  unlinkChild(parentId: string, childId: string, reasonCode: string, reasonNote?: string): Observable<{ message: string }> {
    return this.http.delete<{ message: string }>(
      apiUrl(`/parents/${parentId}/link-child/${childId}`),
      { body: { reason_code: reasonCode, reason_note: reasonNote || '' } },
    );
  }

  inviteToPortal(parentId: string): Observable<{ message: string }> {
    return this.http.post<{ message: string }>(
      apiUrl(`/parents/${parentId}/invite`),
      {},
    );
  }

  revokeAccess(parentId: string): Observable<{ message: string }> {
    return this.http.post<{ message: string }>(
      apiUrl(`/parents/${parentId}/revoke-access`),
      {},
    );
  }
}
