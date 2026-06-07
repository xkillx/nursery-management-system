import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { apiUrl } from '../../../core/config/api.config';
import { AttendanceChildRecord, AttendanceSessionRecord, AttendanceState } from '../models/attendance-child.models';
import { ChildRecord, ChildWritePayload, StaffListQuery, StatusFilter } from '../models/children.models';
import { GuardianRecord, GuardianWritePayload, ChildGuardianLinkRecord, GuardianChildLinkWritePayload } from '../models/guardians.models';
import { InviteCreatePayload, InviteRecord, InviteRole, InviteStatus, InviteStatusFilter } from '../models/invites.models';

interface StaffListResponse<T> {
  items: T[];
}

interface ChildApiModel {
  id: string;
  full_name: string;
  date_of_birth: string;
  start_date: string;
  end_date?: string;
  core_hourly_rate_minor: number;
  notes?: string;
  is_active: boolean;
  left_at?: string;
  left_reason_code?: string;
  left_reason_note?: string;
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

interface AttendanceChildApiModel {
  id: string;
  full_name: string;
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
            fullName: child.full_name,
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
      fullName: child.full_name,
      dateOfBirth: child.date_of_birth,
      startDate: child.start_date,
      endDate: child.end_date ?? null,
      coreHourlyRateMinor: child.core_hourly_rate_minor,
      notes: child.notes ?? null,
      isActive: child.is_active,
      leftAt: child.left_at ?? null,
      leftReasonCode: child.left_reason_code ?? null,
      leftReasonNote: child.left_reason_note ?? null,
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
}
