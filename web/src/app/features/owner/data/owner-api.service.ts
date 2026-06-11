import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { apiUrl } from '../../../core/config/api.config';
import {
  OwnerGrantManagerAccessResult,
  OwnerManagerAccessRecord,
  OwnerSiteSummariesResponse,
} from '../models/owner.models';

interface ApiSiteSummariesResponse {
  billing_month: string;
  attendance_local_date: string;
  currency_code: string;
  totals: ApiSiteSummaryTotals;
  sites: ApiSiteSummary[];
}

interface ApiSiteSummaryTotals {
  active_manager_count: number;
  pending_manager_invite_count: number;
  active_children_count: number;
  checked_in_today_count: number;
  incomplete_attendance_count: number;
  draft_count: number;
  issued_count: number;
  overdue_count: number;
  payment_failed_count: number;
  paid_count: number;
  total_issued_minor: number;
  total_paid_minor: number;
  outstanding_minor: number;
  overdue_outstanding_minor: number;
}

interface ApiAttendanceSummary {
  checked_in_today_count: number;
  incomplete_attendance_count: number;
}

interface ApiFundingReadiness {
  included_child_count: number;
  flagged_child_count: number;
  missing_profile_count: number;
  explicit_zero_count: number;
  under_one_hour_count: number;
  above_160_hours_count: number;
}

interface ApiInvoicePaymentHealth {
  draft_count: number;
  issued_count: number;
  overdue_count: number;
  payment_failed_count: number;
  paid_count: number;
  total_issued_minor: number;
  total_paid_minor: number;
  outstanding_minor: number;
  overdue_outstanding_minor: number;
  failed_payment_count: number;
}

interface ApiSiteSummary {
  site_id: string;
  site_name: string;
  setup_status: string;
  active_manager_count: number;
  pending_manager_invite_count: number;
  active_children_count: number;
  attendance: ApiAttendanceSummary;
  funding_readiness: ApiFundingReadiness;
  invoice_payment_health: ApiInvoicePaymentHealth;
}

interface ApiManagerAccessRecord {
  membership_id: string;
  user_id: string;
  email: string;
  is_active: boolean;
}

interface ApiGrantResult {
  outcome: string;
  membership_id: string | null;
  invite: { email: string; expires_at: string } | null;
}

@Injectable({ providedIn: 'root' })
export class OwnerApiService {
  private readonly http = inject(HttpClient);

  getSiteSummaries(params?: { billingMonth?: string; siteId?: string }): Observable<OwnerSiteSummariesResponse> {
    let httpParams = new HttpParams();
    if (params?.billingMonth) {
      httpParams = httpParams.set('billing_month', params.billingMonth);
    }
    if (params?.siteId) {
      httpParams = httpParams.set('site_id', params.siteId);
    }

    return this.http
      .get<ApiSiteSummariesResponse>(apiUrl('/owner/site-summaries'), { params: httpParams })
      .pipe(map((response) => this.mapSiteSummaries(response)));
  }

  listManagerAccess(siteId: string, status: 'active' | 'inactive' | 'all' = 'active'): Observable<OwnerManagerAccessRecord[]> {
    const params = new HttpParams()
      .set('site_id', siteId)
      .set('status', status);

    return this.http
      .get<ApiManagerAccessRecord[]>(apiUrl('/owner/manager-access'), { params })
      .pipe(map((records) => records.map((r) => this.mapManagerAccessRecord(r))));
  }

  grantManagerAccess(siteId: string, email: string): Observable<OwnerGrantManagerAccessResult> {
    return this.http
      .post<ApiGrantResult>(apiUrl(`/owner/sites/${siteId}/manager-access`), { email })
      .pipe(map((result) => this.mapGrantResult(result)));
  }

  deactivateManagerAccess(siteId: string, membershipId: string): Observable<void> {
    return this.http.post<void>(
      apiUrl(`/owner/sites/${siteId}/manager-access/${membershipId}/actions/deactivate`),
      {},
    );
  }

  reactivateManagerAccess(siteId: string, membershipId: string): Observable<void> {
    return this.http.post<void>(
      apiUrl(`/owner/sites/${siteId}/manager-access/${membershipId}/actions/activate`),
      {},
    );
  }

  private mapSiteSummaries(response: ApiSiteSummariesResponse): OwnerSiteSummariesResponse {
    return {
      billingMonth: response.billing_month,
      attendanceLocalDate: response.attendance_local_date,
      currencyCode: response.currency_code,
      totals: this.mapTotals(response.totals),
      sites: response.sites.map((s) => this.mapSite(s)),
    };
  }

  private mapTotals(t: ApiSiteSummaryTotals) {
    return {
      activeManagerCount: t.active_manager_count,
      pendingManagerInviteCount: t.pending_manager_invite_count,
      activeChildrenCount: t.active_children_count,
      checkedInTodayCount: t.checked_in_today_count,
      incompleteAttendanceCount: t.incomplete_attendance_count,
      draftCount: t.draft_count,
      issuedCount: t.issued_count,
      overdueCount: t.overdue_count,
      paymentFailedCount: t.payment_failed_count,
      paidCount: t.paid_count,
      totalIssuedMinor: t.total_issued_minor,
      totalPaidMinor: t.total_paid_minor,
      outstandingMinor: t.outstanding_minor,
      overdueOutstandingMinor: t.overdue_outstanding_minor,
    };
  }

  private mapSite(s: ApiSiteSummary) {
    return {
      siteId: s.site_id,
      siteName: s.site_name,
      setupStatus: s.setup_status,
      activeManagerCount: s.active_manager_count,
      pendingManagerInviteCount: s.pending_manager_invite_count,
      activeChildrenCount: s.active_children_count,
      attendance: {
        checkedInTodayCount: s.attendance.checked_in_today_count,
        incompleteAttendanceCount: s.attendance.incomplete_attendance_count,
      },
      fundingReadiness: {
        includedChildCount: s.funding_readiness.included_child_count,
        flaggedChildCount: s.funding_readiness.flagged_child_count,
        missingProfileCount: s.funding_readiness.missing_profile_count,
        explicitZeroCount: s.funding_readiness.explicit_zero_count,
        underOneHourCount: s.funding_readiness.under_one_hour_count,
        above160HoursCount: s.funding_readiness.above_160_hours_count,
      },
      invoicePaymentHealth: {
        draftCount: s.invoice_payment_health.draft_count,
        issuedCount: s.invoice_payment_health.issued_count,
        overdueCount: s.invoice_payment_health.overdue_count,
        paymentFailedCount: s.invoice_payment_health.payment_failed_count,
        paidCount: s.invoice_payment_health.paid_count,
        totalIssuedMinor: s.invoice_payment_health.total_issued_minor,
        totalPaidMinor: s.invoice_payment_health.total_paid_minor,
        outstandingMinor: s.invoice_payment_health.outstanding_minor,
        overdueOutstandingMinor: s.invoice_payment_health.overdue_outstanding_minor,
        failedPaymentCount: s.invoice_payment_health.failed_payment_count,
      },
    };
  }

  private mapManagerAccessRecord(r: ApiManagerAccessRecord): OwnerManagerAccessRecord {
    return {
      membershipId: r.membership_id,
      userId: r.user_id,
      email: r.email,
      isActive: r.is_active,
    };
  }

  private mapGrantResult(r: ApiGrantResult): OwnerGrantManagerAccessResult {
    return {
      outcome: r.outcome as OwnerGrantManagerAccessResult['outcome'],
      membershipId: r.membership_id,
      invite: r.invite ? { email: r.invite.email, expiresAt: r.invite.expires_at } : null,
    };
  }
}
