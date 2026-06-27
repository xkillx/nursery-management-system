export interface BookedSession {
  day_of_week: number;
  session_type: SessionTypeRef;
}

export interface SessionTypeRef {
  id: string;
  name: string;
  start_time: string;
  end_time: string;
  is_active: boolean;
}

export interface BookingPattern {
  id: string;
  child_id: string;
  effective_from: string;
  effective_to: string | null;
  is_current: boolean;
  created_at: string;
  entries: BookedSession[];
}

export interface BookingPatternInput {
  effective_from: string;
  entries: BookingPatternEntryInput[];
}

export interface BookingPatternEntryInput {
  day_of_week: number;
  session_type_id: string;
}
