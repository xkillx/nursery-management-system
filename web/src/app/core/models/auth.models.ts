import { AppRole } from '../constants/roles';

export interface UserModel {
  id: string;
  email: string;
}

export interface MembershipModel {
  membership_id: string;
  tenant_id: string;
  tenant_name: string;
  branch_id: string;
  branch_name: string;
  role: AppRole;
}

export interface AuthResponse {
  access_token: string;
  token_type: string;
  expires_in_seconds: number;
  user: UserModel;
  active_membership: MembershipModel;
  available_memberships: MembershipModel[];
}

export interface LoginRequest {
  email: string;
  password: string;
  membership_id?: string;
}

export interface AuthState {
  accessToken: string | null;
  user: UserModel | null;
  activeMembership: MembershipModel | null;
  availableMemberships: MembershipModel[];
}
