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

export function isMembershipSelectionRequired(body: any): body is MembershipSelectionRequiredResponse {
  if (body?.code !== 'membership_selection_required') {
    return false;
  }

  const memberships: unknown[] = body.available_memberships;
  if (!Array.isArray(memberships) || memberships.length === 0) {
    return false;
  }

  return memberships.every(
    (m: any) =>
      typeof m?.membership_id === 'string' &&
      typeof m?.tenant_name === 'string' &&
      (m?.branch_name === null || typeof m?.branch_name === 'string') &&
      typeof m?.role === 'string',
  );
}
