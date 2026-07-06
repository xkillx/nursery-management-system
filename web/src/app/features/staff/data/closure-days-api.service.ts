import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { apiUrl } from '../../../core/config/api.config';
import { ClosureDay } from '../models/closure-day.models';

interface ApiClosureDay {
  id: string;
  branch_id: string;
  date: string;
  reason: string | null;
  created_at: string;
}

interface ApiListResponse {
  items: ApiClosureDay[];
  total: number;
  page: number;
  page_size: number;
}

@Injectable({ providedIn: 'root' })
export class ClosureDaysApiService {
  private readonly http = inject(HttpClient);

  list(branchId: string, from: string, to: string): Observable<ClosureDay[]> {
    const params = new HttpParams().set('from', from).set('to', to);
    return this.http
      .get<ApiListResponse>(apiUrl(`/sites/${branchId}/closure-days`), { params })
      .pipe(map((res) => res.items.map((c) => this.toClosureDay(c))));
  }

  create(branchId: string, date: string, reason?: string): Observable<ClosureDay> {
    const body: Record<string, string> = { date };
    if (reason) {
      body['reason'] = reason;
    }
    return this.http
      .post<{ closure_day: ApiClosureDay }>(apiUrl(`/sites/${branchId}/closure-days`), body)
      .pipe(map((res) => this.toClosureDay(res.closure_day)));
  }

  delete(branchId: string, id: string): Observable<void> {
    return this.http.delete<void>(apiUrl(`/sites/${branchId}/closure-days/${id}`));
  }

  private toClosureDay(c: ApiClosureDay): ClosureDay {
    return {
      id: c.id,
      branch_id: c.branch_id,
      date: c.date,
      reason: c.reason,
      created_at: c.created_at,
    };
  }
}
