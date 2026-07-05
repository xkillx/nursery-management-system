export type AcademicTermKind = 'autumn' | 'spring' | 'summer';

export interface AcademicTerm {
  id: string;
  name: string;
  kind: AcademicTermKind;
  start_date: string;
  end_date: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface AcademicTermInput {
  name: string;
  kind: AcademicTermKind;
  start_date: string;
  end_date: string;
}

export interface AcademicTermUpdateInput {
  name?: string;
  kind?: AcademicTermKind;
  start_date?: string;
  end_date?: string;
}
