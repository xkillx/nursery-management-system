import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { apiUrl } from '../../../core/config/api.config';
import { AdHocBooking, AdHocBookingInput } from '../models/ad-hoc-booking.models';

interface ApiAdHocBooking {
  id: string;
  child_id: string;
  calendar_date: string;
  session_type_id: string;
  session_type_name: string;
  status: string;
  created_at: string;
  updated_at: string;
}

interface ApiListResponse {
  items: ApiAdHocBooking[];
  total: number;
  page: number;
  page_size: number;
}

@Injectable({ providedIn: 'root' })
export class AdHocBookingsApiService {
  private readonly http = inject(HttpClient);

  listBookings(
    siteId: string,
    params: { childId?: string; from?: string; to?: string },
  ): Observable<AdHocBooking[]> {
    let httpParams = new HttpParams();
    if (params.childId) httpParams = httpParams.set('child_id', params.childId);
    if (params.from) httpParams = httpParams.set('from', params.from);
    if (params.to) httpParams = httpParams.set('to', params.to);
    return this.http
      .get<ApiListResponse>(apiUrl(`/sites/${siteId}/ad-hoc-bookings`), { params: httpParams })
      .pipe(map((res) => res.items.map((b) => this.toBooking(b))));
  }

  createBooking(siteId: string, payload: AdHocBookingInput): Observable<AdHocBooking> {
    return this.http
      .post<ApiAdHocBooking>(apiUrl(`/sites/${siteId}/ad-hoc-bookings`), payload)
      .pipe(map((b) => this.toBooking(b)));
  }

  cancelBooking(siteId: string, bookingId: string): Observable<void> {
    return this.http.post<void>(
      apiUrl(`/sites/${siteId}/ad-hoc-bookings/${bookingId}/actions/cancel`),
      {},
    );
  }

  private toBooking(b: ApiAdHocBooking): AdHocBooking {
    return {
      id: b.id,
      child_id: b.child_id,
      calendar_date: b.calendar_date,
      session_type_id: b.session_type_id,
      session_type_name: b.session_type_name,
      status: b.status as AdHocBooking['status'],
      created_at: b.created_at,
      updated_at: b.updated_at,
    };
  }
}
