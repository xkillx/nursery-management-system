export type BookingType = 'recurring' | 'ad_hoc' | 'hourly';

export type BookingStatus = 'active' | 'paused' | 'cancelled';

export interface UnifiedBooking {
  bookingType: BookingType;
  id: string;
  childId: string;
  childFirstName: string;
  childLastName: string;
  startDate: string;
  endDate: string | null;
  roomId: string | null;
  roomName: string | null;
  sessionTemplateId: string;
  status: BookingStatus;
  createdAt: string;
  updatedAt: string;
}

export interface UnifiedBookingListResult {
  items: UnifiedBooking[];
  total: number;
  page: number;
  pageSize: number;
}

export interface BookingListFilters {
  childId?: string;
  roomId?: string;
  sessionTypeId?: string;
  status?: string;
  fundingType?: string;
  search?: string;
  from?: string;
  to?: string;
}

export interface CreateRecurringBookingRequest {
  child_id: string;
  session_template_id: string;
  room_id: string;
  days_of_week: number[];
  effective_start_date: string;
  effective_end_date?: string;
  funding_type?: string;
  funding_hours_per_week?: number;
  la_reference?: string;
}

export interface CreateAdHocBookingRequest {
  child_id: string;
  calendar_date: string;
  session_type_id: string;
}

export interface CreateHourlyBookingRequest {
  child_id: string;
  calendar_date: string;
  start_time_minutes: number;
  duration_minutes: number;
  session_type_id?: string;
}

export interface UpdateRecurringBookingRequest {
  room_id?: string;
  days_of_week?: number[];
  effective_start_date?: string;
  effective_end_date?: string;
  funding_type?: string;
  funding_hours_per_week?: number;
  la_reference?: string;
}
