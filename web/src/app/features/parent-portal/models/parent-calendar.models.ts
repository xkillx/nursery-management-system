export interface CalendarDayItem {
  date: string;
  is_open: boolean;
  reason: 'none' | 'closure_day' | 'holiday_period';
}

export interface ParentCalendarListResult {
  items: CalendarDayItem[];
}
