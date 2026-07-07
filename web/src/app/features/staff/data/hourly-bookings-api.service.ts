import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { apiUrl } from '../../../core/config/api.config';
import { HourlyBooking, CreateHourlyBookingRequest } from '../models/hourly-booking.models';

interface ApiHourlyBooking {
  id: string;
  child_id: string;
  calendar_date: string;
  start_time_minutes: number;
  duration_minutes: number;
  session_type_id: string | null;
  status: string;
  created_at: string;
  updated_at: string;
}

interface ApiListResponse {
  items: ApiHourlyBooking[];
  total: number;
  page: number;
  page_size: number;
}

@Injectable({ providedIn: 'root' })
export class HourlyBookingsApiService {
  private readonly http = inject(HttpClient);

  listBookings(
    siteId: string,
    params: { childId?: string; from?: string; to?: string },
  ): Observable<HourlyBooking[]> {
    let httpParams = new HttpParams();
    if (params.childId) httpParams = httpParams.set('child_id', params.childId);
    if (params.from) httpParams = httpParams.set('from', params.from);
    if (params.to) httpParams = httpParams.set('to', params.to);
    return this.http
      .get<ApiListResponse>(apiUrl(`/sites/${siteId}/hourly-bookings`), { params: httpParams })
      .pipe(map((res) => res.items.map((b) => this.toBooking(b))));
  }

  createBooking(siteId: string, payload: CreateHourlyBookingRequest): Observable<HourlyBooking> {
    return this.http
      .post<{ hourly_booking: ApiHourlyBooking }>(apiUrl(`/sites/${siteId}/hourly-bookings`), payload)
      .pipe(map((r) => this.toBooking(r.hourly_booking)));
  }

  cancelBooking(siteId: string, bookingId: string): Observable<void> {
    return this.http.post<void>(
      apiUrl(`/sites/${siteId}/hourly-bookings/${bookingId}/cancel`),
      {},
    );
  }

  private toBooking(b: ApiHourlyBooking): HourlyBooking {
    return {
      id: b.id,
      child_id: b.child_id,
      calendar_date: b.calendar_date,
      start_time_minutes: b.start_time_minutes,
      duration_minutes: b.duration_minutes,
      session_type_id: b.session_type_id,
      status: b.status as HourlyBooking['status'],
      created_at: b.created_at,
      updated_at: b.updated_at,
    };
  }
}
