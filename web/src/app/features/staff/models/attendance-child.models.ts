export interface AttendanceChildRecord {
  id: string;
  fullName: string;
  enrollmentComplete: boolean;
  attendanceState: string;
  openSessionId: string | null;
  checkedInAt: string | null;
}
