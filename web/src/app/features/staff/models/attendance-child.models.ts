export type AttendanceState = 'not_checked_in' | 'checked_in';

export interface AttendanceChildRecord {
  id: string;
  fullName: string;
  enrollmentComplete: boolean;
  attendanceState: AttendanceState;
  openSessionId: string | null;
  checkedInAt: string | null;
  hasIncompleteSession: boolean;
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
