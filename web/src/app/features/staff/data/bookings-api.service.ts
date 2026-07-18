import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { apiUrl } from '../../../core/config/api.config';
import {
  UnifiedBooking,
  UnifiedBookingListResult,
  BookingListFilters,
  BookingType,
  BookingStatus,
  CreateRecurringBookingRequest,
  CreateAdHocBookingRequest,
  CreateHourlyBookingRequest,
  UpdateRecurringBookingRequest,
} from '../models/booking.models';

interface UnifiedBookingApi {
  booking_type: string;
  id: string;
  child_id: string;
  child_first_name: string;
  child_last_name: string;
  start_date: string;
  end_date: string | null;
  room_id: string | null;
  room_name: string | null;
  session_template_id?: string;
  status: string;
  created_at: string;
  updated_at: string;
}

interface UnifiedListResponseApi {
  items: UnifiedBookingApi[];
  total: number;
  page: number;
  page_size: number;
}

@Injectable({ providedIn: 'root' })
export class BookingsApiService {
  private readonly http = inject(HttpClient);

  listBookings(
    siteId: string,
    filters: BookingListFilters,
    page: number,
    pageSize: number,
  ): Observable<UnifiedBookingListResult> {
    let params = new HttpParams()
      .set('page', String(page))
      .set('page_size', String(pageSize));

    if (filters.childId) params = params.set('child_id', filters.childId);
    if (filters.roomId) params = params.set('room_id', filters.roomId);
    if (filters.sessionTypeId) params = params.set('session_type_id', filters.sessionTypeId);
    if (filters.status) params = params.set('status', filters.status);
    if (filters.fundingType) params = params.set('funding_type', filters.fundingType);
    if (filters.search) params = params.set('search', filters.search);
    if (filters.from) params = params.set('from', filters.from);
    if (filters.to) params = params.set('to', filters.to);

    return this.http
      .get<UnifiedListResponseApi>(apiUrl(`/sites/${siteId}/bookings`), { params })
      .pipe(map((res) => this.toListResult(res)));
  }

  getBooking(siteId: string, bookingId: string): Observable<UnifiedBooking> {
    return this.http
      .get<{ booking: UnifiedBookingApi }>(apiUrl(`/sites/${siteId}/bookings/${bookingId}`))
      .pipe(map((res) => this.toBooking(res.booking)));
  }

  createRecurringBooking(siteId: string, data: CreateRecurringBookingRequest): Observable<unknown> {
    return this.http.post(apiUrl(`/sites/${siteId}/bookings`), data);
  }

  createAdHocBooking(siteId: string, data: CreateAdHocBookingRequest): Observable<unknown> {
    return this.http.post(apiUrl(`/sites/${siteId}/ad-hoc-bookings`), data);
  }

  createHourlyBooking(siteId: string, data: CreateHourlyBookingRequest): Observable<unknown> {
    return this.http.post(apiUrl(`/sites/${siteId}/hourly-bookings`), data);
  }

  updateRecurringBooking(siteId: string, bookingId: string, data: UpdateRecurringBookingRequest): Observable<unknown> {
    return this.http.patch(apiUrl(`/sites/${siteId}/bookings/${bookingId}`), data);
  }

  cancelBooking(siteId: string, bookingType: BookingType, bookingId: string): Observable<void> {
    switch (bookingType) {
      case 'recurring':
        return this.http.post<void>(apiUrl(`/sites/${siteId}/bookings/${bookingId}/cancel`), {});
      case 'ad_hoc':
        return this.http.post<void>(apiUrl(`/sites/${siteId}/ad-hoc-bookings/${bookingId}/actions/cancel`), {});
      case 'hourly':
        return this.http.post<void>(apiUrl(`/sites/${siteId}/hourly-bookings/${bookingId}/cancel`), {});
    }
  }

  private toListResult(res: UnifiedListResponseApi): UnifiedBookingListResult {
    return {
      items: res.items.map((b) => this.toBooking(b)),
      total: res.total,
      page: res.page,
      pageSize: res.page_size,
    };
  }

  private toBooking(b: UnifiedBookingApi): UnifiedBooking {
    return {
      bookingType: b.booking_type as BookingType,
      id: b.id,
      childId: b.child_id,
      childFirstName: b.child_first_name,
      childLastName: b.child_last_name,
      startDate: b.start_date,
      endDate: b.end_date ?? null,
      roomId: b.room_id ?? null,
      roomName: b.room_name ?? null,
      sessionTemplateId: b.session_template_id ?? undefined,
      status: b.status as BookingStatus,
      createdAt: b.created_at,
      updatedAt: b.updated_at,
    };
  }
}
