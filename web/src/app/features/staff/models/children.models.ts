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
  siteCoreHourlyRateMinor: number | null;
  notes: string | null;
  isActive: boolean;
  hasCurrentRoom?: boolean;
  hasBookingPattern?: boolean;
  enrollmentComplete: boolean;
  missingRequirements: string[];
  photoUrl: string | null;
  createdAt: string;
  updatedAt: string;
  // Legacy alias for the still-imported manager-registration-intake stepper.
  primaryRoomId?: string | null;
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
