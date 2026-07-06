import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { apiUrl } from '../../../core/config/api.config';
import {
  SessionTemplate,
  SessionTemplateInput,
  SessionTemplateListItem,
  SessionTemplateUpdateInput,
} from '../models/session-template.models';

interface ApiSessionTypeRef {
  id: string;
  name: string;
  start_time: string;
  end_time: string;
  is_active: boolean;
}

interface ApiSessionTemplateEntry {
  id: string;
  day_of_week: number;
  session_type: ApiSessionTypeRef;
}

interface ApiSessionTemplate {
  id: string;
  branch_id: string;
  name: string;
  description?: string | null;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  entries: ApiSessionTemplateEntry[];
}

interface ApiListResponse {
  items: ApiSessionTemplate[];
  total: number;
  page: number;
  page_size: number;
}

@Injectable({ providedIn: 'root' })
export class StaffSessionTemplatesApiService {
  private readonly http = inject(HttpClient);

  listSessionTemplates(
    siteId: string,
    options: { includeArchived?: boolean } = {},
  ): Observable<SessionTemplateListItem[]> {
    let params = new HttpParams();
    if (options.includeArchived) {
      params = params.set('include_archived', 'true');
    }
    return this.http
      .get<ApiListResponse>(apiUrl(`/sites/${siteId}/session-templates`), { params })
      .pipe(map((res) => res.items.map((s) => this.toListItem(s))));
  }

  getSessionTemplate(siteId: string, templateId: string): Observable<SessionTemplate> {
    return this.http
      .get<ApiSessionTemplate>(apiUrl(`/sites/${siteId}/session-templates/${templateId}`))
      .pipe(map((s) => this.toTemplate(s)));
  }

  createSessionTemplate(siteId: string, payload: SessionTemplateInput): Observable<SessionTemplate> {
    return this.http
      .post<ApiSessionTemplate>(apiUrl(`/sites/${siteId}/session-templates`), payload)
      .pipe(map((s) => this.toTemplate(s)));
  }

  updateSessionTemplate(
    siteId: string,
    templateId: string,
    payload: SessionTemplateUpdateInput,
  ): Observable<SessionTemplate> {
    return this.http
      .patch<ApiSessionTemplate>(
        apiUrl(`/sites/${siteId}/session-templates/${templateId}`),
        payload,
      )
      .pipe(map((s) => this.toTemplate(s)));
  }

  archiveSessionTemplate(siteId: string, templateId: string): Observable<void> {
    return this.http.post<void>(
      apiUrl(`/sites/${siteId}/session-templates/${templateId}/actions/archive`),
      {},
    );
  }

  reactivateSessionTemplate(siteId: string, templateId: string): Observable<SessionTemplate> {
    return this.http
      .post<ApiSessionTemplate>(
        apiUrl(`/sites/${siteId}/session-templates/${templateId}/actions/reactivate`),
        {},
      )
      .pipe(map((s) => this.toTemplate(s)));
  }

  private toListItem(s: ApiSessionTemplate): SessionTemplateListItem {
    return {
      id: s.id,
      branchId: s.branch_id,
      name: s.name,
      description: s.description ?? null,
      isActive: s.is_active,
      createdAt: s.created_at,
      updatedAt: s.updated_at,
      entries: [] as never[],
    };
  }

  private toTemplate(s: ApiSessionTemplate): SessionTemplate {
    return {
      id: s.id,
      branchId: s.branch_id,
      name: s.name,
      description: s.description ?? null,
      isActive: s.is_active,
      createdAt: s.created_at,
      updatedAt: s.updated_at,
      entries: s.entries.map((e) => ({
        id: e.id,
        dayOfWeek: e.day_of_week,
        sessionType: {
          id: e.session_type.id,
          name: e.session_type.name,
          startTime: e.session_type.start_time,
          endTime: e.session_type.end_time,
          isActive: e.session_type.is_active,
        },
      })),
    };
  }
}
