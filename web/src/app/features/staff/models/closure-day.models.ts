export interface ClosureDay {
  id: string;
  branch_id: string;
  date: string;
  reason: string | null;
  created_at: string;
}

export interface ClosureDayInput {
  date: string;
  reason?: string;
}
