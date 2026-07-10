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
  fundingProfileId: string | null;
  fundedAllowanceMinutes: number | null;
  fundingUpdatedAt: string | null;
  photoUrl: string | null;
  flags: FundingOverviewFlag[];
}

export interface FundingOverviewRecord {
  billingMonth: string;
  summary: FundingOverviewSummary;
  items: FundingOverviewItem[];
}

export interface FundingProfileRecord {
  id: string;
  childId: string;
  billingMonth: string;
  fundedAllowanceMinutes: number;
  createdAt: string;
  updatedAt: string;
}

export interface FundingProfileWritePayload {
  billing_month: string;
  funded_allowance_minutes: number;
}
