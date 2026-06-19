export interface BookedSession {
  dayOfWeek: number;
  sessionType: SessionTypeRef;
}

export interface SessionTypeRef {
  id: string;
  name: string;
  startTime: string;
  endTime: string;
  isActive: boolean;
}

export interface BookingPattern {
  id: string;
  childId: string;
  effectiveFrom: string;
  effectiveTo: string | null;
  isCurrent: boolean;
  createdAt: string;
  entries: BookedSession[];
}

export interface BookingPatternInput {
  effectiveFrom: string;
  entries: BookingPatternEntryInput[];
}

export interface BookingPatternEntryInput {
  dayOfWeek: number;
  sessionTypeId: string;
}
