export type FundingOverviewFlag =
  | 'missing_profile'
  | 'explicit_zero_allowance'
  | 'under_one_hour_allowance'
  | 'above_160_hours_allowance';

export interface FundingOverviewSummary {
  includedChildCount: number;
  flaggedChildCount: number;
  missingProfileCount: number;
  explicitZeroCount: number;
  underOneHourCount: number;
  above160HoursCount: number;
}

export interface FundingOverviewItem {
  childId: string;
  childName: string;
  isActive: boolean;
  startDate: string;
  endDate: string | null;
  fundingRecordId: string | null;
  fundingUpdatedAt: string | null;
  photoUrl: string | null;
  flags: FundingOverviewFlag[];
  remainingMinutes: number | null;
}

export interface FundingOverviewRecord {
  billingMonth: string;
  summary: FundingOverviewSummary;
  items: FundingOverviewItem[];
}

export interface FundingRecord {
  id: string;
  childId: string;
  fundingEnabled: boolean;
  fundingType: string;
  fundingModel: string;
  fundedHoursPerWeek: number | null;
  fundingStartDate: string | null;
  fundingEndDate: string | null;
  eligibilityCode: string | null;
  eligibilityCodeValidated: boolean;
  evidenceReceived: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface FundingRecordDetail {
  record: FundingRecord;
  fundedAllowanceMinutes: number;
  allocation: AllocationEntry[];
  history: FundingHistoryEntry[];
}

export interface AllocationEntry {
  bookingId: string;
  effectiveStartDate: string;
  effectiveEndDate: string | null;
  daysOfWeek: number[];
  sessionTypeName: string;
  sessionDurationMinutes: number;
}

export interface FundingHistoryEntry {
  id: string;
  fundingType: string | null;
  fundingModel: string | null;
  fundedHoursPerWeek: number | null;
  fundingStartDate: string | null;
  fundingEndDate: string | null;
  changedAt: string;
}

export interface FundingRecordWritePayload {
  funding_enabled: boolean;
  funding_type: string;
  funding_model: string;
  funded_hours_per_week?: number | null;
  funding_start_date?: string | null;
  funding_end_date?: string | null;
  eligibility_code?: string | null;
  eligibility_code_validated?: boolean;
  evidence_received?: boolean;
}
