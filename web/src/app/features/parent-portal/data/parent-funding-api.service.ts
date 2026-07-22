import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { apiUrl } from '../../../core/config/api.config';
import {
  ParentFundingEntitlement,
  ParentFundingBreakdown,
  ParentFundingRecord,
} from '../models/parent-funding.models';

interface ParentFundingEntitlementApiModel {
  child_id: string;
  child_first_name: string;
  child_middle_name?: string | null;
  child_last_name?: string | null;
  funding_type?: string | null;
  funded_hours_per_week?: number | null;
  funded_allowance_minutes: number;
  booked_hours_this_week: number;
}

interface ParentFundingBreakdownApiModel {
  record: ParentFundingRecordApiModel;
  funded_allowance_minutes: number;
  allocation: ParentAllocationEntryApiModel[];
  history: ParentFundingHistoryEntryApiModel[];
}

interface ParentFundingRecordApiModel {
  id: string;
  child_id: string;
  funding_enabled: boolean;
  funding_type: string;
  funding_model: string;
  funded_hours_per_week?: number | null;
  funding_start_date?: string | null;
  funding_end_date?: string | null;
  eligibility_code?: string | null;
}

interface ParentAllocationEntryApiModel {
  booking_id: string;
  effective_start_date: string;
  effective_end_date?: string | null;
  days_of_week: number[];
  session_type_name: string;
  session_duration_minutes: number;
}

interface ParentFundingHistoryEntryApiModel {
  id: string;
  funding_type?: string | null;
  funding_model?: string | null;
  funded_hours_per_week?: number | null;
  funding_start_date?: string | null;
  funding_end_date?: string | null;
  changed_at: string;
}

@Injectable({ providedIn: 'root' })
export class ParentFundingApiService {
  private readonly http = inject(HttpClient);

  getFunding(): Observable<ParentFundingEntitlement[]> {
    return this.http
      .get<{ items: ParentFundingEntitlementApiModel[] }>(apiUrl('/parent/funding'))
      .pipe(map((res) => res.items.map((item) => this.toEntitlement(item))));
  }

  getFundingBreakdown(childId: string, billingMonth: string): Observable<ParentFundingBreakdown> {
    return this.http
      .get<ParentFundingBreakdownApiModel>(apiUrl(`/parent/funding/${childId}/breakdown`), {
        params: new HttpParams({ fromObject: { billing_month: billingMonth } }),
      })
      .pipe(map((detail) => this.toBreakdown(detail)));
  }

  private toEntitlement(item: ParentFundingEntitlementApiModel): ParentFundingEntitlement {
    return {
      childId: item.child_id,
      childFirstName: item.child_first_name,
      childMiddleName: item.child_middle_name ?? null,
      childLastName: item.child_last_name ?? null,
      fundingType: item.funding_type ?? null,
      fundedHoursPerWeek: item.funded_hours_per_week ?? null,
      fundedAllowanceMinutes: item.funded_allowance_minutes,
      bookedHoursThisWeek: item.booked_hours_this_week,
    };
  }

  private toBreakdown(detail: ParentFundingBreakdownApiModel): ParentFundingBreakdown {
    return {
      record: this.toRecord(detail.record),
      fundedAllowanceMinutes: detail.funded_allowance_minutes,
      allocation: detail.allocation.map((a) => ({
        bookingId: a.booking_id,
        effectiveStartDate: a.effective_start_date,
        effectiveEndDate: a.effective_end_date ?? null,
        daysOfWeek: a.days_of_week,
        sessionTypeName: a.session_type_name,
        sessionDurationMinutes: a.session_duration_minutes,
      })),
      history: detail.history.map((h) => ({
        id: h.id,
        fundingType: h.funding_type ?? null,
        fundingModel: h.funding_model ?? null,
        fundedHoursPerWeek: h.funded_hours_per_week ?? null,
        fundingStartDate: h.funding_start_date ?? null,
        fundingEndDate: h.funding_end_date ?? null,
        changedAt: h.changed_at,
      })),
    };
  }

  private toRecord(record: ParentFundingRecordApiModel): ParentFundingRecord {
    return {
      id: record.id,
      childId: record.child_id,
      fundingEnabled: record.funding_enabled,
      fundingType: record.funding_type,
      fundingModel: record.funding_model,
      fundedHoursPerWeek: record.funded_hours_per_week ?? null,
      fundingStartDate: record.funding_start_date ?? null,
      fundingEndDate: record.funding_end_date ?? null,
      eligibilityCode: record.eligibility_code ?? null,
    };
  }
}
