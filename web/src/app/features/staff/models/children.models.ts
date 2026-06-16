export type StatusFilter = 'active' | 'inactive' | 'all';

export interface ChildRecord {
  id: string;
  firstName?: string;
  middleName?: string | null;
  lastName?: string | null;
  fullName: string;
  dateOfBirth: string;
  startDate: string;
  endDate: string | null;
  coreHourlyRateMinor?: number | null;
  siteCoreHourlyRateMinor: number | null;
  notes: string | null;
  isActive: boolean;
  leftAt: string | null;
  leftReasonCode: string | null;
  leftReasonNote: string | null;
  primaryRoomId: string | null;
  enrollmentComplete: boolean;
  missingRequirements: string[];
  createdAt: string;
  updatedAt: string;
}

export interface ChildWritePayload {
  first_name: string;
  middle_name?: string | null;
  last_name?: string | null;
  date_of_birth: string;
  start_date: string;
  end_date?: string;
  notes?: string;
  primary_room_id?: string | null;
}

export interface StaffListQuery {
  status: StatusFilter;
  limit: number;
  offset: number;
}
