export type AdHocBookingStatus = 'active' | 'cancelled';

export interface AdHocBooking {
  id: string;
  child_id: string;
  calendar_date: string;
  session_type_id: string;
  session_type_name: string;
  status: AdHocBookingStatus;
  created_at: string;
  updated_at: string;
}

export interface AdHocBookingInput {
  child_id: string;
  calendar_date: string;
  session_type_id: string;
}
