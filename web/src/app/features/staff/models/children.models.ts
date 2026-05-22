export type StatusFilter = 'active' | 'inactive' | 'all';

export interface ChildRecord {
  id: string;
  fullName: string;
  dateOfBirth: string;
  startDate: string;
  endDate: string | null;
  coreHourlyRateMinor: number;
  notes: string | null;
  isActive: boolean;
  leftAt: string | null;
  leftReasonCode: string | null;
  leftReasonNote: string | null;
  enrollmentComplete: boolean;
  missingRequirements: string[];
  createdAt: string;
  updatedAt: string;
}

export interface ChildWritePayload {
  full_name: string;
  date_of_birth: string;
  start_date: string;
  core_hourly_rate_minor: number;
  end_date?: string;
  notes?: string;
}

export interface StaffListQuery {
  status: StatusFilter;
  limit: number;
  offset: number;
}
