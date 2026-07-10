import { MembershipModel } from './auth.models';

export interface ApiErrorBody {
  code: string;
  message: string;
  details?: unknown;
  request_id?: string;
}

export interface MappedApiError {
  code: string;
  message: string;
  requestId: string | null;
  fieldErrors: Record<string, string>;
}

export interface MembershipSelectionRequiredResponse {
  code: 'membership_selection_required';
  message: string;
  request_id?: string;
  available_memberships: MembershipModel[];
}

export function isMembershipSelectionRequired(body: unknown): body is MembershipSelectionRequiredResponse {
  if (!body || typeof body !== 'object' || (body as Record<string, unknown>).code !== 'membership_selection_required') {
    return false;
  }

  const memberships: unknown[] = (body as MembershipSelectionRequiredResponse).available_memberships ?? [];
  if (!Array.isArray(memberships) || memberships.length === 0) {
    return false;
  }

  return memberships.every(
    (m: unknown) => {
      if (!m || typeof m !== 'object') return false;
      const obj = m as Record<string, unknown>;
      return typeof obj.membership_id === 'string' &&
        typeof obj.tenant_name === 'string' &&
        (obj.branch_name === null || typeof obj.branch_name === 'string') &&
        typeof obj.role === 'string';
    },
  );
}
