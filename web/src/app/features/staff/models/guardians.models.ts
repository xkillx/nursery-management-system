import { StatusFilter } from './children.models';

export interface GuardianRecord {
  id: string;
  fullName: string;
  email: string | null;
  phone: string | null;
  notes: string | null;
  isActive: boolean;
  deactivatedAt: string | null;
  deactivationReasonCode: string | null;
  deactivationReasonNote: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface GuardianWritePayload {
  full_name: string;
  email?: string;
  phone?: string;
  notes?: string;
}

export interface GuardianListQuery {
  status: StatusFilter;
  limit: number;
  offset: number;
}
