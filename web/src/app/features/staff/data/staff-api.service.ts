import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { apiUrl } from '../../../core/config/api.config';
import { AttendanceChildRecord } from '../models/attendance-child.models';
import { ChildRecord, ChildWritePayload, StaffListQuery, StatusFilter } from '../models/children.models';
import { GuardianRecord, GuardianWritePayload } from '../models/guardians.models';

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
            attendanceState: child.attendance_state,
            openSessionId: child.open_session_id ?? null,
            checkedInAt: child.checked_in_at ?? null,
          })),
        ),
      );
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
}
