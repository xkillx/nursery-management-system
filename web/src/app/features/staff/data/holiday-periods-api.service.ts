import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { apiUrl } from '../../../core/config/api.config';
import { HolidayPeriod } from '../models/holiday-period.models';

interface ApiHolidayPeriod {
  id: string;
  branch_id: string;
  name: string;
  start_date: string;
  end_date: string;
  type: string;
  created_at: string;
  updated_at: string;
}

interface ApiListResponse {
  items: ApiHolidayPeriod[];
}

@Injectable({ providedIn: 'root' })
export class HolidayPeriodsApiService {
  private readonly http = inject(HttpClient);

  list(branchId: string): Observable<HolidayPeriod[]> {
    return this.http
      .get<ApiListResponse>(apiUrl(`/sites/${branchId}/holiday-periods`))
      .pipe(map((res) => res.items.map((c) => this.toHolidayPeriod(c))));
  }

  create(branchId: string, body: { name: string; type: string; start_date: string; end_date: string }): Observable<HolidayPeriod> {
    return this.http
      .post<{ holiday_period: ApiHolidayPeriod }>(apiUrl(`/sites/${branchId}/holiday-periods`), body)
      .pipe(map((res) => this.toHolidayPeriod(res.holiday_period)));
  }

  update(branchId: string, id: string, body: Partial<{ name: string; type: string; start_date: string; end_date: string }>): Observable<HolidayPeriod> {
    return this.http
      .patch<{ holiday_period: ApiHolidayPeriod }>(apiUrl(`/sites/${branchId}/holiday-periods/${id}`), body)
      .pipe(map((res) => this.toHolidayPeriod(res.holiday_period)));
  }

  delete(branchId: string, id: string): Observable<void> {
    return this.http.delete<void>(apiUrl(`/sites/${branchId}/holiday-periods/${id}`));
  }

  private toHolidayPeriod(c: ApiHolidayPeriod): HolidayPeriod {
    return {
      id: c.id,
      branch_id: c.branch_id,
      name: c.name,
      start_date: c.start_date,
      end_date: c.end_date,
      type: c.type as HolidayPeriod['type'],
      created_at: c.created_at,
      updated_at: c.updated_at,
    };
  }
}
