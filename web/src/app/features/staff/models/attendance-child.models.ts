export type AttendanceState = 'not_checked_in' | 'checked_in' | 'absent';

export interface AttendanceChildRecord {
  id: string;
  fullName: string;
  enrollmentComplete: boolean;
  attendanceState: AttendanceState;
  openSessionId: string | null;
  checkedInAt: string | null;
  hasIncompleteSession: boolean;
  absenceMarkerId: string | null;
  absenceMarkedAt: string | null;
}

export interface AttendanceSessionRecord {
  id: string;
  childId: string;
  status: string;
  checkInAt: string;
  checkOutAt: string | null;
  checkInLocalDate: string;
  checkOutLocalDate: string | null;
  durationMinutes: number | null;
  createdAt: string;
  updatedAt: string;
}

export type AttendanceCorrectionReasonCode =
  | 'missed_check_in'
  | 'missed_check_out'
  | 'incorrect_time'
  | 'duplicate_entry'
  | 'other';

export interface AttendanceCorrectionPayload {
  sessionId?: string;
  childId?: string;
  checkInAt: string;
  checkOutAt: string;
  reasonCode: AttendanceCorrectionReasonCode;
  reasonNote?: string;
}

export interface IssuedInvoiceWarning {
  billingMonth: string;
  invoiceId: string;
  invoiceNumber: string;
  status: string;
}

export interface CorrectionSessionContext {
  childId: string;
  selectedLocalDate: string;
  invoiceWarning: IssuedInvoiceWarning | null;
  items: AttendanceSessionRecord[];
}

export interface CorrectionHistoryEvent {
  id: string;
  eventType: 'check_in' | 'check_out' | 'correction';
  occurredAt: string;
  localDate: string;
  recordedByUserId: string;
  recordedByMembershipId: string;
  recordedByLabel: string | null;
  reasonCode: string | null;
  reasonNote: string | null;
  previousCheckInAt: string | null;
  previousCheckOutAt: string | null;
  correctedCheckInAt: string | null;
  correctedCheckOutAt: string | null;
  createdByCorrection: boolean;
}

export interface CorrectionHistory {
  session: AttendanceSessionRecord;
  items: CorrectionHistoryEvent[];
}
