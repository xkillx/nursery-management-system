export interface ParentFundingEntitlement {
  childId: string;
  childFirstName: string;
  childMiddleName: string | null;
  childLastName: string | null;
  fundingType: string | null;
  fundedHoursPerWeek: number | null;
  fundedAllowanceMinutes: number;
  bookedHoursThisWeek: number;
}

export interface ParentFundingBreakdown {
  record: ParentFundingRecord;
  fundedAllowanceMinutes: number;
  allocation: ParentAllocationEntry[];
  history: ParentFundingHistoryEntry[];
}

export interface ParentFundingRecord {
  id: string;
  childId: string;
  fundingEnabled: boolean;
  fundingType: string;
  fundingModel: string;
  fundedHoursPerWeek: number | null;
  fundingStartDate: string | null;
  fundingEndDate: string | null;
  eligibilityCode: string | null;
}

export interface ParentAllocationEntry {
  bookingId: string;
  effectiveStartDate: string;
  effectiveEndDate: string | null;
  daysOfWeek: number[];
  sessionTypeName: string;
  sessionDurationMinutes: number;
}

export interface ParentFundingHistoryEntry {
  id: string;
  fundingType: string | null;
  fundingModel: string | null;
  fundedHoursPerWeek: number | null;
  fundingStartDate: string | null;
  fundingEndDate: string | null;
  changedAt: string;
}
