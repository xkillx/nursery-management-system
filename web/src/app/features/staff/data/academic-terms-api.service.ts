import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { apiUrl } from '../../../core/config/api.config';
import { AcademicTerm, AcademicTermInput, AcademicTermUpdateInput } from '../models/academic-term.models';

interface ApiAcademicTerm {
  id: string;
  name: string;
  kind: string;
  start_date: string;
  end_date: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

interface ApiListResponse {
  academic_terms: ApiAcademicTerm[];
}

@Injectable({ providedIn: 'root' })
export class AcademicTermsApiService {
  private readonly http = inject(HttpClient);

  listTerms(siteId: string, options: { includeArchived?: boolean } = {}): Observable<AcademicTerm[]> {
    let params = new HttpParams();
    if (options.includeArchived) {
      params = params.set('include_archived', 'true');
    }
    return this.http
      .get<ApiListResponse>(apiUrl(`/sites/${siteId}/academic-terms`), { params })
      .pipe(map((res) => res.academic_terms.map((t) => this.toTerm(t))));
  }

  createTerm(siteId: string, payload: AcademicTermInput): Observable<AcademicTerm> {
    return this.http
      .post<ApiAcademicTerm>(apiUrl(`/sites/${siteId}/academic-terms`), payload)
      .pipe(map((t) => this.toTerm(t)));
  }

  updateTerm(siteId: string, termId: string, payload: AcademicTermUpdateInput): Observable<AcademicTerm> {
    return this.http
      .patch<ApiAcademicTerm>(apiUrl(`/sites/${siteId}/academic-terms/${termId}`), payload)
      .pipe(map((t) => this.toTerm(t)));
  }

  archiveTerm(siteId: string, termId: string): Observable<void> {
    return this.http.post<void>(
      apiUrl(`/sites/${siteId}/academic-terms/${termId}/actions/archive`),
      {},
    );
  }

  private toTerm(t: ApiAcademicTerm): AcademicTerm {
    return {
      id: t.id,
      name: t.name,
      kind: t.kind as AcademicTerm['kind'],
      start_date: t.start_date,
      end_date: t.end_date,
      is_active: t.is_active,
      created_at: t.created_at,
      updated_at: t.updated_at,
    };
  }
}
