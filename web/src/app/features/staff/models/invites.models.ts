export type InviteRole = 'practitioner' | 'parent';

export type InviteStatus = 'pending' | 'accepted' | 'revoked' | 'expired';

export type InviteStatusFilter = InviteStatus | 'all';

export interface InviteRecord {
  id: string;
  email: string;
  role: InviteRole;
  status: InviteStatus;
  expiresAt: string;
  acceptedAt: string | null;
  revokedAt: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface InviteCreatePayload {
  email: string;
  role: InviteRole;
}
