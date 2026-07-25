import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { apiUrl } from '../../../core/config/api.config';
import { CalendarDayItem } from '../models/parent-calendar.models';

interface ApiCalendarDayItem {
  date: string;
  is_open: boolean;
  reason: string;
}

interface ApiCalendarListResponse {
  items: ApiCalendarDayItem[];
}

@Injectable({ providedIn: 'root' })
export class ParentCalendarApiService {
  private readonly http = inject(HttpClient);

  listClosureDays(from?: string, to?: string): Observable<CalendarDayItem[]> {
    let params = new HttpParams();
    if (from) params = params.set('from', from);
    if (to) params = params.set('to', to);

    return this.http
      .get<ApiCalendarListResponse>(apiUrl('/parent/closure-days'), { params })
      .pipe(map((res) => res.items.map((item) => this.toCalendarDay(item))));
  }

  listHolidayPeriods(from?: string, to?: string): Observable<CalendarDayItem[]> {
    let params = new HttpParams();
    if (from) params = params.set('from', from);
    if (to) params = params.set('to', to);

    return this.http
      .get<ApiCalendarListResponse>(apiUrl('/parent/holiday-periods'), { params })
      .pipe(map((res) => res.items.map((item) => this.toCalendarDay(item))));
  }

  private toCalendarDay(item: ApiCalendarDayItem): CalendarDayItem {
    return {
      date: item.date,
      is_open: item.is_open,
      reason: item.reason as CalendarDayItem['reason'],
    };
  }
}
