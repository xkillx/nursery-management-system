export interface ParentRecord {
  id: string;
  first_name: string;
  last_name?: string | null;
  email?: string | null;
  phone?: string | null;
  address_line1?: string | null;
  address_line2?: string | null;
  address_city?: string | null;
  address_postcode?: string | null;
  relationship_to_child?: string | null;
  has_parental_responsibility: boolean;
  can_pick_up: boolean;
  is_emergency_contact: boolean;
  notes?: string | null;
  user_id?: string | null;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface ParentChildLink {
  id: string;
  child_id: string;
  ended_at?: string | null;
  ended_reason_code?: string | null;
  ended_reason_note?: string | null;
  created_at: string;
  updated_at: string;
}

export interface ParentWithChildren extends ParentRecord {
  children: ParentChildLink[];
}

export interface ParentListResponse {
  parents: ParentRecord[];
  total_count: number;
  page: number;
  page_size: number;
}

export type ParentStatusFilter = 'active' | 'inactive' | 'all';
