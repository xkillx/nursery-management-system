export type HourlyBookingStatus = 'active' | 'cancelled';

export interface HourlyBooking {
  id: string;
  child_id: string;
  calendar_date: string;
  start_time_minutes: number;
  duration_minutes: number;
  session_type_id: string | null;
  status: HourlyBookingStatus;
  created_at: string;
  updated_at: string;
}

export interface CreateHourlyBookingRequest {
  child_id: string;
  calendar_date: string;
  start_time_minutes: number;
  duration_minutes: number;
  session_type_id?: string;
}
