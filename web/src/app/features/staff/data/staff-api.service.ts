import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { apiUrl } from '../../../core/config/api.config';
import { AbsenceMarkerRecord, AttendanceChildRecord, AttendanceCorrectionPayload, AttendanceSessionRecord, AttendanceState, CorrectionHistory, CorrectionHistoryEvent, CorrectionSessionContext, IssuedInvoiceWarning } from '../models/attendance-child.models';
import { ChildRecord, ChildWritePayload, StaffListQuery, StatusFilter } from '../models/children.models';
import { GuardianRecord, GuardianWritePayload, ChildGuardianLinkRecord, GuardianChildLinkWritePayload } from '../models/guardians.models';
import { FundingProfileRecord, FundingProfileWritePayload, FundingOverviewRecord, FundingOverviewItem, FundingOverviewFlag } from '../models/funding.models';
import { InviteCreatePayload, InviteRecord, InviteRole, InviteStatus, InviteStatusFilter } from '../models/invites.models';
import {
  RegistrationProfileResponse, CollectionPasswordPayload,
  ConsentRecord, ConsentWithCompletenessResponse, ConsentWritePayload, RegistrationWorkflowStatus,
  CompleteRegistrationPayload, CompleteRegistrationResponse,
} from '../models/registration-profile.models';
import { formatChildName } from '../utils/manager-list-formatters';

interface StaffListResponse<T> {
  items: T[];
}

interface ChildApiModel {
  id: string;
  first_name: string;
  middle_name?: string | null;
  last_name?: string | null;
  date_of_birth: string;
  start_date: string;
  end_date?: string;
  core_hourly_rate_minor: number | null;
  site_core_hourly_rate_minor: number | null;
  notes?: string;
  is_active: boolean;
  left_at?: string;
  left_reason_code?: string;
  left_reason_note?: string;
  primary_room_id?: string | null;
  enrollment_complete: boolean;
  missing_requirements?: string[];
  created_at: string;
  updated_at: string;
}

interface GuardianApiModel {
  id: string;
  full_name: string;
  email?: string;
  phone?: string;
  notes?: string;
  is_active: boolean;
  deactivated_at?: string;
  deactivation_reason_code?: string;
  deactivation_reason_note?: string;
  created_at: string;
  updated_at: string;
}

interface AbsenceMarkerApiModel {
  id: string;
  child_id: string;
  local_date: string;
  marked_at: string;
  cleared_at?: string;
  created_at: string;
  updated_at: string;
}

interface AttendanceChildApiModel {
  id: string;
  first_name: string;
  middle_name?: string | null;
  last_name?: string | null;
  enrollment_complete: boolean;
  attendance_state: string;
  open_session_id?: string;
  checked_in_at?: string;
  has_incomplete_session: boolean;
  absence_marker_id?: string;
  absence_marked_at?: string;
}

interface AttendanceSessionApiModel {
  id: string;
  child_id: string;
  status: string;
  check_in_at: string;
  check_out_at?: string;
  check_in_local_date: string;
  check_out_local_date?: string;
  duration_minutes?: number;
  created_at: string;
  updated_at: string;
}

interface CorrectionSessionContextApiModel {
  child_id: string;
  selected_local_date: string;
  invoice_warning?: InvoiceWarningApiModel;
  items: AttendanceSessionApiModel[];
}

interface InvoiceWarningApiModel {
  billing_month: string;
  invoice_id: string;
  invoice_number: string;
  status: string;
}

interface CorrectionHistoryApiModel {
  session: AttendanceSessionApiModel;
  items: CorrectionHistoryEventApiModel[];
}

interface CorrectionHistoryEventApiModel {
  id: string;
  event_type: string;
  occurred_at: string;
  local_date: string;
  recorded_by_user_id: string;
  recorded_by_membership_id: string;
  recorded_by_label?: string;
  reason_code?: string;
  reason_note?: string;
  previous_check_in_at?: string;
  previous_check_out_at?: string;
  corrected_check_in_at?: string;
  corrected_check_out_at?: string;
  created_by_correction: boolean;
}

interface FundingProfileApiModel {
  id: string;
  child_id: string;
  billing_month: string;
  funded_allowance_minutes: number;
  created_at: string;
  updated_at: string;
}

interface RegistrationProfileApiModel {
  child: {
    id: string;
    first_name: string;
    middle_name?: string | null;
    last_name?: string | null;
    date_of_birth: string;
  };
  profile_exists: boolean;
  profile: { id: string; created_at: string; updated_at: string } | null;
  demographics_home: Record<string, unknown> | null;
  medical_dietary: Record<string, unknown> | null;
  health_contacts: Record<string, unknown> | null;
  social_development: Record<string, unknown> | null;
  parent_carers: Record<string, unknown>[];
  emergency_contacts: Record<string, unknown>[];
  authorised_collectors: Record<string, unknown>[];
  collection: Record<string, unknown> | null;
  funding_support: Record<string, unknown> | null;
  routine_care: Record<string, unknown> | null;
  gdpr_declaration: Record<string, unknown> | null;
  completeness: {
    is_complete: boolean;
    missing_sections: string[];
    sections: { code: string; status: string; missing_fields: string[] }[];
  };
}

interface FundingOverviewApiModel {
  billing_month: string;
  summary: {
    included_child_count: number;
    flagged_child_count: number;
    missing_profile_count: number;
    explicit_zero_count: number;
    under_one_hour_count: number;
    above_160_hours_count: number;
  };
  items: FundingOverviewItemApiModel[];
}

interface FundingOverviewItemApiModel {
  child_id: string;
  child_first_name: string;
  child_middle_name?: string | null;
  child_last_name?: string | null;
  is_active: boolean;
  start_date: string;
  end_date?: string | null;
  funding_profile_id?: string | null;
  funded_allowance_minutes?: number | null;
  funding_updated_at?: string | null;
  flags: string[];
}

interface InviteApiModel {
  id: string;
  email: string;
  role: string;
  status: string;
  expires_at: string;
  accepted_at?: string | null;
  revoked_at?: string | null;
  created_at: string;
  updated_at: string;
}

interface LinkedGuardianSummaryApiModel {
  id: string;
  full_name: string;
  email?: string;
  phone?: string;
  is_active: boolean;
}

interface ChildGuardianLinkApiModel {
  id: string;
  guardian_id: string;
  child_id: string;
  guardian: LinkedGuardianSummaryApiModel;
  created_at: string;
  updated_at: string;
}

@Injectable({ providedIn: 'root' })
export class StaffApiService {
  private readonly http = inject(HttpClient);

  listChildren(query: StaffListQuery): Observable<ChildRecord[]> {
    return this.http
      .get<StaffListResponse<ChildApiModel>>(apiUrl('/children'), {
        params: this.buildListParams(query.status, query.limit, query.offset),
      })
      .pipe(map((response) => response.items.map((child) => this.toChildRecord(child))));
  }

  createChild(payload: ChildWritePayload): Observable<ChildRecord> {
    return this.http
      .post<ChildApiModel>(apiUrl('/children'), payload)
      .pipe(map((child) => this.toChildRecord(child)));
  }

  updateChild(childId: string, payload: ChildWritePayload): Observable<ChildRecord> {
    return this.http
      .patch<ChildApiModel>(apiUrl(`/children/${childId}`), payload)
      .pipe(map((child) => this.toChildRecord(child)));
  }

  getChild(childId: string): Observable<ChildRecord> {
    return this.http
      .get<ChildApiModel>(apiUrl(`/children/${childId}`))
      .pipe(map((child) => this.toChildRecord(child)));
  }

  listChildGuardianLinks(childId: string): Observable<ChildGuardianLinkRecord[]> {
    return this.http
      .get<StaffListResponse<ChildGuardianLinkApiModel>>(apiUrl(`/children/${childId}/guardian-child-links`))
      .pipe(map((response) => response.items.map((link) => this.toChildGuardianLinkRecord(link))));
  }

  createGuardianChildLink(payload: GuardianChildLinkWritePayload): Observable<ChildGuardianLinkRecord> {
    return this.http
      .post<ChildGuardianLinkApiModel>(apiUrl('/guardian-child-links'), payload)
      .pipe(map((link) => this.toChildGuardianLinkRecord(link)));
  }

  listGuardians(query: StaffListQuery): Observable<GuardianRecord[]> {
    return this.http
      .get<StaffListResponse<GuardianApiModel>>(apiUrl('/guardians'), {
        params: this.buildListParams(query.status, query.limit, query.offset),
      })
      .pipe(map((response) => response.items.map((guardian) => this.toGuardianRecord(guardian))));
  }

  createGuardian(payload: GuardianWritePayload): Observable<GuardianRecord> {
    return this.http
      .post<GuardianApiModel>(apiUrl('/guardians'), payload)
      .pipe(map((guardian) => this.toGuardianRecord(guardian)));
  }

  updateGuardian(guardianId: string, payload: GuardianWritePayload): Observable<GuardianRecord> {
    return this.http
      .patch<GuardianApiModel>(apiUrl(`/guardians/${guardianId}`), payload)
      .pipe(map((guardian) => this.toGuardianRecord(guardian)));
  }

  listAttendanceChildren(): Observable<AttendanceChildRecord[]> {
    return this.http
      .get<StaffListResponse<AttendanceChildApiModel>>(apiUrl('/children/attendance'))
      .pipe(
        map((response) =>
          response.items.map((child) => ({
            id: child.id,
            firstName: child.first_name,
            middleName: child.middle_name ?? null,
            lastName: child.last_name ?? null,
            fullName: this.childDisplayName(child),
            enrollmentComplete: child.enrollment_complete,
            attendanceState: child.attendance_state as AttendanceState,
            openSessionId: child.open_session_id ?? null,
            checkedInAt: child.checked_in_at ?? null,
            hasIncompleteSession: child.has_incomplete_session,
            absenceMarkerId: child.absence_marker_id ?? null,
            absenceMarkedAt: child.absence_marked_at ?? null,
          })),
        ),
      );
  }

  checkInChild(childId: string): Observable<AttendanceSessionRecord> {
    return this.http
      .post<AttendanceSessionApiModel>(apiUrl('/attendance/check-ins'), { child_id: childId })
      .pipe(map((session) => this.toAttendanceSessionRecord(session)));
  }

  checkOutChild(childId: string): Observable<AttendanceSessionRecord> {
    return this.http
      .post<AttendanceSessionApiModel>(apiUrl('/attendance/check-outs'), { child_id: childId })
      .pipe(map((session) => this.toAttendanceSessionRecord(session)));
  }

  markChildAbsent(childId: string): Observable<AbsenceMarkerRecord> {
    return this.http
      .post<AbsenceMarkerApiModel>(apiUrl('/attendance/absence-markers'), { child_id: childId })
      .pipe(map((marker) => this.toAbsenceMarkerRecord(marker)));
  }

  clearAbsenceMarker(absenceMarkerId: string): Observable<AbsenceMarkerRecord> {
    return this.http
      .post<AbsenceMarkerApiModel>(apiUrl(`/attendance/absence-markers/${absenceMarkerId}/clear`), null)
      .pipe(map((marker) => this.toAbsenceMarkerRecord(marker)));
  }

  listCorrectionSessions(childId: string, localDate: string): Observable<CorrectionSessionContext> {
    return this.http
      .get<CorrectionSessionContextApiModel>(apiUrl('/attendance/sessions'), {
        params: new HttpParams({ fromObject: { child_id: childId, local_date: localDate } }),
      })
      .pipe(map((ctx) => this.toCorrectionSessionContext(ctx)));
  }

  getCorrectionHistory(sessionId: string): Observable<CorrectionHistory> {
    return this.http
      .get<CorrectionHistoryApiModel>(apiUrl(`/attendance/sessions/${sessionId}/history`))
      .pipe(map((history) => this.toCorrectionHistory(history)));
  }

  correctAttendance(payload: AttendanceCorrectionPayload): Observable<AttendanceSessionRecord> {
    const apiPayload: Record<string, string | undefined> = {
      check_in_at: payload.checkInAt,
      check_out_at: payload.checkOutAt,
      reason_code: payload.reasonCode,
      reason_note: payload.reasonNote,
    };
    if (payload.sessionId) {
      apiPayload['session_id'] = payload.sessionId;
    }
    if (payload.childId) {
      apiPayload['child_id'] = payload.childId;
    }
    return this.http
      .post<AttendanceSessionApiModel>(apiUrl('/attendance/corrections'), apiPayload)
      .pipe(map((session) => this.toAttendanceSessionRecord(session)));
  }

  listInvites(status: InviteStatusFilter = 'pending'): Observable<InviteRecord[]> {
    return this.http
      .get<StaffListResponse<InviteApiModel>>(apiUrl('/invites'), {
        params: new HttpParams({ fromObject: { status } }),
      })
      .pipe(map((response) => response.items.map((invite) => this.toInviteRecord(invite))));
  }

  createInvite(payload: InviteCreatePayload): Observable<InviteRecord> {
    return this.http
      .post<InviteApiModel>(apiUrl('/invites'), payload)
      .pipe(map((invite) => this.toInviteRecord(invite)));
  }

  resendInvite(inviteId: string): Observable<InviteRecord> {
    return this.http
      .post<InviteApiModel>(apiUrl(`/invites/${inviteId}/resend`), null)
      .pipe(map((invite) => this.toInviteRecord(invite)));
  }

  revokeInvite(inviteId: string): Observable<InviteRecord> {
    return this.http
      .post<InviteApiModel>(apiUrl(`/invites/${inviteId}/revoke`), null)
      .pipe(map((invite) => this.toInviteRecord(invite)));
  }

  getFundingProfile(childId: string, billingMonth: string): Observable<FundingProfileRecord> {
    return this.http
      .get<FundingProfileApiModel>(apiUrl(`/funding/children/${childId}`), {
        params: new HttpParams({ fromObject: { billing_month: billingMonth } }),
      })
      .pipe(map((profile) => this.toFundingProfileRecord(profile)));
  }

  upsertFundingProfile(childId: string, payload: FundingProfileWritePayload): Observable<FundingProfileRecord> {
    return this.http
      .put<FundingProfileApiModel>(apiUrl(`/funding/children/${childId}`), payload)
      .pipe(map((profile) => this.toFundingProfileRecord(profile)));
  }

  getFundingOverview(billingMonth: string): Observable<FundingOverviewRecord> {
    return this.http
      .get<FundingOverviewApiModel>(apiUrl('/funding/overview'), {
        params: new HttpParams({ fromObject: { billing_month: billingMonth } }),
      })
      .pipe(map((overview) => this.toFundingOverviewRecord(overview)));
  }

  getRegistrationProfile(childId: string): Observable<RegistrationProfileResponse> {
    return this.http.get<RegistrationProfileApiModel>(apiUrl(`/children/${childId}/registration-profile`))
      .pipe(map((profile) => this.toRegistrationProfileRecord(profile)));
  }

  patchRegistrationProfile(childId: string, patch: Record<string, unknown>): Observable<RegistrationProfileResponse> {
    return this.http.patch<RegistrationProfileApiModel>(apiUrl(`/children/${childId}/registration-profile`), patch)
      .pipe(map((profile) => this.toRegistrationProfileRecord(profile)));
  }

  setRegistrationCollectionPassword(childId: string, password: string): Observable<RegistrationProfileResponse> {
    return this.http.put<RegistrationProfileApiModel>(
      apiUrl(`/children/${childId}/registration-profile/collection-password`),
      { password } as CollectionPasswordPayload,
    ).pipe(map((profile) => this.toRegistrationProfileRecord(profile)));
  }

  private buildListParams(status: StatusFilter, limit: number, offset: number): HttpParams {
    return new HttpParams({
      fromObject: {
        status,
        limit,
        offset,
      },
    });
  }

  private toAttendanceSessionRecord(session: AttendanceSessionApiModel): AttendanceSessionRecord {
    return {
      id: session.id,
      childId: session.child_id,
      status: session.status,
      checkInAt: session.check_in_at,
      checkOutAt: session.check_out_at ?? null,
      checkInLocalDate: session.check_in_local_date,
      checkOutLocalDate: session.check_out_local_date ?? null,
      durationMinutes: session.duration_minutes ?? null,
      createdAt: session.created_at,
      updatedAt: session.updated_at,
    };
  }

  private toChildRecord(child: ChildApiModel): ChildRecord {
    return {
      id: child.id,
      firstName: child.first_name,
      middleName: child.middle_name ?? null,
      lastName: child.last_name ?? null,
      fullName: this.childDisplayName(child),
      dateOfBirth: child.date_of_birth,
      startDate: child.start_date,
      endDate: child.end_date ?? null,
      coreHourlyRateMinor: child.core_hourly_rate_minor,
      siteCoreHourlyRateMinor: child.site_core_hourly_rate_minor ?? null,
      notes: child.notes ?? null,
      isActive: child.is_active,
      leftAt: child.left_at ?? null,
      leftReasonCode: child.left_reason_code ?? null,
      leftReasonNote: child.left_reason_note ?? null,
      primaryRoomId: child.primary_room_id ?? null,
      enrollmentComplete: child.enrollment_complete,
      missingRequirements: child.missing_requirements ?? [],
      createdAt: child.created_at,
      updatedAt: child.updated_at,
    };
  }

  private toGuardianRecord(guardian: GuardianApiModel): GuardianRecord {
    return {
      id: guardian.id,
      fullName: guardian.full_name,
      email: guardian.email ?? null,
      phone: guardian.phone ?? null,
      notes: guardian.notes ?? null,
      isActive: guardian.is_active,
      deactivatedAt: guardian.deactivated_at ?? null,
      deactivationReasonCode: guardian.deactivation_reason_code ?? null,
      deactivationReasonNote: guardian.deactivation_reason_note ?? null,
      createdAt: guardian.created_at,
      updatedAt: guardian.updated_at,
    };
  }

  private toInviteRecord(invite: InviteApiModel): InviteRecord {
    return {
      id: invite.id,
      email: invite.email,
      role: invite.role as InviteRole,
      status: invite.status as InviteStatus,
      expiresAt: invite.expires_at,
      acceptedAt: invite.accepted_at ?? null,
      revokedAt: invite.revoked_at ?? null,
      createdAt: invite.created_at,
      updatedAt: invite.updated_at,
    };
  }

  private toChildGuardianLinkRecord(link: ChildGuardianLinkApiModel): ChildGuardianLinkRecord {
    return {
      id: link.id,
      guardianId: link.guardian_id,
      childId: link.child_id,
      guardian: {
        id: link.guardian.id,
        fullName: link.guardian.full_name,
        email: link.guardian.email ?? null,
        phone: link.guardian.phone ?? null,
        isActive: link.guardian.is_active,
      },
      createdAt: link.created_at,
      updatedAt: link.updated_at,
    };
  }

  private toCorrectionSessionContext(ctx: CorrectionSessionContextApiModel): CorrectionSessionContext {
    return {
      childId: ctx.child_id,
      selectedLocalDate: ctx.selected_local_date,
      invoiceWarning: ctx.invoice_warning ? this.toInvoiceWarning(ctx.invoice_warning) : null,
      items: ctx.items.map((s) => this.toAttendanceSessionRecord(s)),
    };
  }

  private toInvoiceWarning(w: InvoiceWarningApiModel): IssuedInvoiceWarning {
    return {
      billingMonth: w.billing_month,
      invoiceId: w.invoice_id,
      invoiceNumber: w.invoice_number,
      status: w.status,
    };
  }

  private toCorrectionHistory(history: CorrectionHistoryApiModel): CorrectionHistory {
    return {
      session: this.toAttendanceSessionRecord(history.session),
      items: history.items.map((e) => this.toCorrectionHistoryEvent(e)),
    };
  }

  private toCorrectionHistoryEvent(e: CorrectionHistoryEventApiModel): CorrectionHistoryEvent {
    return {
      id: e.id,
      eventType: e.event_type as CorrectionHistoryEvent['eventType'],
      occurredAt: e.occurred_at,
      localDate: e.local_date,
      recordedByUserId: e.recorded_by_user_id,
      recordedByMembershipId: e.recorded_by_membership_id,
      recordedByLabel: e.recorded_by_label ?? null,
      reasonCode: e.reason_code ?? null,
      reasonNote: e.reason_note ?? null,
      previousCheckInAt: e.previous_check_in_at ?? null,
      previousCheckOutAt: e.previous_check_out_at ?? null,
      correctedCheckInAt: e.corrected_check_in_at ?? null,
      correctedCheckOutAt: e.corrected_check_out_at ?? null,
      createdByCorrection: e.created_by_correction,
    };
  }

  private toAbsenceMarkerRecord(marker: AbsenceMarkerApiModel): AbsenceMarkerRecord {
    return {
      id: marker.id,
      childId: marker.child_id,
      localDate: marker.local_date,
      markedAt: marker.marked_at,
      clearedAt: marker.cleared_at ?? null,
      createdAt: marker.created_at,
      updatedAt: marker.updated_at,
    };
  }

  private toFundingOverviewRecord(overview: FundingOverviewApiModel): FundingOverviewRecord {
    return {
      billingMonth: overview.billing_month,
      summary: {
        includedChildCount: overview.summary.included_child_count,
        flaggedChildCount: overview.summary.flagged_child_count,
        missingProfileCount: overview.summary.missing_profile_count,
        explicitZeroCount: overview.summary.explicit_zero_count,
        underOneHourCount: overview.summary.under_one_hour_count,
        above160HoursCount: overview.summary.above_160_hours_count,
      },
      items: overview.items.map((item) => this.toFundingOverviewItem(item)),
    };
  }

  private toFundingOverviewItem(item: FundingOverviewItemApiModel): FundingOverviewItem {
    return {
      childId: item.child_id,
      childName: formatChildName({
        firstName: item.child_first_name,
        middleName: item.child_middle_name,
        lastName: item.child_last_name,
      }),
      isActive: item.is_active,
      startDate: item.start_date,
      endDate: item.end_date ?? null,
      fundingProfileId: item.funding_profile_id ?? null,
      fundedAllowanceMinutes: item.funded_allowance_minutes ?? null,
      fundingUpdatedAt: item.funding_updated_at ?? null,
      flags: item.flags as FundingOverviewFlag[],
    };
  }

  private toFundingProfileRecord(profile: FundingProfileApiModel): FundingProfileRecord {
    return {
      id: profile.id,
      childId: profile.child_id,
      billingMonth: profile.billing_month,
      fundedAllowanceMinutes: profile.funded_allowance_minutes,
      createdAt: profile.created_at,
      updatedAt: profile.updated_at,
    };
  }

  private toRegistrationProfileRecord(profile: RegistrationProfileApiModel): RegistrationProfileResponse {
    return {
      child: {
        id: profile.child.id,
        fullName: formatChildName({
          firstName: profile.child.first_name,
          middleName: profile.child.middle_name,
          lastName: profile.child.last_name,
        }),
        dateOfBirth: profile.child.date_of_birth,
      },
      profileExists: profile.profile_exists,
      profile: profile.profile ? {
        id: profile.profile.id,
        createdAt: profile.profile.created_at,
        updatedAt: profile.profile.updated_at,
      } : null,
      demographicsHome: profile.demographics_home as RegistrationProfileResponse['demographicsHome'],
      medicalDietary: profile.medical_dietary as RegistrationProfileResponse['medicalDietary'],
      healthContacts: profile.health_contacts as RegistrationProfileResponse['healthContacts'],
      socialDevelopment: profile.social_development as RegistrationProfileResponse['socialDevelopment'],
      parentCarers: (profile.parent_carers ?? []).map((c) => ({
        fullName: (c as Record<string, unknown>)['full_name'] as string,
        relationshipToChild: (c as Record<string, unknown>)['relationship_to_child'] as string | null ?? null,
        address: (c as Record<string, unknown>)['address'] as Record<string, unknown> | null ?? null,
        telephone: (c as Record<string, unknown>)['telephone'] as string | null ?? null,
        email: (c as Record<string, unknown>)['email'] as string | null ?? null,
        workAddress: (c as Record<string, unknown>)['work_address'] as Record<string, unknown> | null ?? null,
        hasParentalResponsibility: (c as Record<string, unknown>)['has_parental_responsibility'] as boolean | null ?? null,
      })),
      emergencyContacts: (profile.emergency_contacts ?? []).map((c) => ({
        fullName: (c as Record<string, unknown>)['full_name'] as string,
        relationshipToChild: (c as Record<string, unknown>)['relationship_to_child'] as string | null ?? null,
        address: (c as Record<string, unknown>)['address'] as Record<string, unknown> | null ?? null,
        telephone: (c as Record<string, unknown>)['telephone'] as string | null ?? null,
        email: (c as Record<string, unknown>)['email'] as string | null ?? null,
        workAddress: (c as Record<string, unknown>)['work_address'] as Record<string, unknown> | null ?? null,
        hasParentalResponsibility: (c as Record<string, unknown>)['has_parental_responsibility'] as boolean | null ?? null,
      })),
      authorisedCollectors: (profile.authorised_collectors ?? []).map((c) => ({
        fullName: (c as Record<string, unknown>)['full_name'] as string,
        relationshipToChild: (c as Record<string, unknown>)['relationship_to_child'] as string | null ?? null,
        address: (c as Record<string, unknown>)['address'] as Record<string, unknown> | null ?? null,
        telephone: (c as Record<string, unknown>)['telephone'] as string | null ?? null,
        email: (c as Record<string, unknown>)['email'] as string | null ?? null,
        workAddress: (c as Record<string, unknown>)['work_address'] as Record<string, unknown> | null ?? null,
        hasParentalResponsibility: (c as Record<string, unknown>)['has_parental_responsibility'] as boolean | null ?? null,
      })),
      collection: profile.collection as RegistrationProfileResponse['collection'],
      fundingSupport: profile.funding_support as RegistrationProfileResponse['fundingSupport'],
      routineCare: profile.routine_care as RegistrationProfileResponse['routineCare'],
      gdprDeclaration: profile.gdpr_declaration as RegistrationProfileResponse['gdprDeclaration'],
      completeness: {
        isComplete: profile.completeness.is_complete,
        missingSections: profile.completeness.missing_sections,
        sections: profile.completeness.sections.map((s) => ({
          code: s.code,
          status: s.status as 'complete' | 'incomplete',
          missingFields: s.missing_fields,
        })),
      },
    };
  }

  private childDisplayName(child: { first_name: string; middle_name?: string | null; last_name?: string | null }): string {
    return formatChildName({
      firstName: child.first_name,
      middleName: child.middle_name,
      lastName: child.last_name,
    });
  }

  getRegistrationWorkflowStatus(childId: string): Observable<RegistrationWorkflowStatus> {
    return this.http.get<RegistrationWorkflowStatus>(apiUrl(`/children/${childId}/registration-workflow-status`));
  }

  getRegistrationConsents(childId: string): Observable<ConsentWithCompletenessResponse> {
    return this.http.get<ConsentWithCompletenessResponse>(apiUrl(`/children/${childId}/registration-consents`));
  }

  createRegistrationConsent(childId: string, payload: ConsentWritePayload): Observable<ConsentRecord> {
    return this.http.post<ConsentRecord>(apiUrl(`/children/${childId}/registration-consents`), payload);
  }

  createRegistrationCompletionAttestation(childId: string): Observable<unknown> {
    return this.http.post(apiUrl(`/children/${childId}/registration-completion-attestations`), null);
  }

  submitCompleteRegistration(payload: CompleteRegistrationPayload): Observable<CompleteRegistrationResponse> {
    return this.http.post<CompleteRegistrationResponse>(apiUrl('/children/with-registration'), payload);
  }

}
