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
