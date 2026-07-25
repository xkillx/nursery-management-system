export type HolidayPeriodType = 'half_term' | 'christmas' | 'easter' | 'summer' | 'other';

export interface HolidayPeriod {
  id: string;
  branch_id: string;
  name: string;
  start_date: string;
  end_date: string;
  type: HolidayPeriodType;
  created_at: string;
  updated_at: string;
}

export const HOLIDAY_PERIOD_TYPE_LABELS: Record<HolidayPeriodType, string> = {
  half_term: 'Half Term',
  christmas: 'Christmas',
  easter: 'Easter',
  summer: 'Summer',
  other: 'Other',
};

export const HOLIDAY_PERIOD_TYPE_OPTIONS = Object.entries(HOLIDAY_PERIOD_TYPE_LABELS).map(
  ([value, label]) => ({ value, label })
);
