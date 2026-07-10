import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable, computed, inject, signal } from '@angular/core';
import { Observable, catchError, map, of, tap } from 'rxjs';

import { apiUrl } from '../config/api.config';
import { AppRole } from '../constants/roles';
import { AuthResponse, AuthState, LoginRequest, MembershipModel, PasswordResetAcceptedResponse, UserModel } from '../models/auth.models';
import { getCookie } from '../utils/cookie.util';

const initialState: AuthState = {
  accessToken: null,
  user: null,
  activeMembership: null,
  availableMemberships: [],
};

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly state = signal<AuthState>(initialState);

  readonly user = computed<UserModel | null>(() => this.state().user);
  readonly activeMembership = computed<MembershipModel | null>(() => this.state().activeMembership);
  readonly role = computed<AppRole | null>(() => this.state().activeMembership?.role ?? null);

  login(email: string, password: string, membershipId?: string, rememberMe = true): Observable<AuthResponse> {
    const payload: LoginRequest = { email, password, remember_me: rememberMe };
    if (membershipId) {
      payload.membership_id = membershipId;
    }

    return this.http.post<AuthResponse>(apiUrl('/auth/login'), payload, { withCredentials: true }).pipe(
      tap((response) => this.applySession(response)),
    );
  }

  refresh(): Observable<AuthResponse> {
    return this.http
      .post<AuthResponse>(apiUrl('/auth/refresh'), {}, {
        withCredentials: true,
        headers: this.csrfHeaders(),
      })
      .pipe(tap((response) => this.applySession(response)));
  }

  logout(): Observable<void> {
    return this.http
      .post<void>(apiUrl('/auth/logout'), {}, {
        withCredentials: true,
        headers: this.csrfHeaders(),
      })
      .pipe(
        map(() => void 0),
        catchError(() => of(void 0)),
        tap(() => this.clearSession()),
      );
  }

  bootstrapSession(): Promise<void> {
    return new Promise((resolve) => {
      this.refresh().subscribe({
        next: () => resolve(),
        error: () => {
          this.clearSession();
          resolve();
        },
      });
    });
  }

  accessToken(): string | null {
    return this.state().accessToken;
  }

  isAuthenticated(): boolean {
    return !!this.state().accessToken;
  }

  currentRole(): AppRole | null {
    return this.role();
  }

  clearSession(): void {
    this.state.set(initialState);
  }

  requestPasswordReset(email: string): Observable<PasswordResetAcceptedResponse> {
    return this.http.post<PasswordResetAcceptedResponse>(
      apiUrl('/auth/password-reset-requests'),
      { email },
    );
  }

  resetPassword(token: string, newPassword: string): Observable<void> {
    return this.http.post<void>(
      apiUrl('/auth/password-resets'),
      { token, new_password: newPassword },
    );
  }

  acceptInvite(token: string, newPassword: string): Observable<void> {
    return this.http.post<void>(
      apiUrl('/invites/accept'),
      { token, new_password: newPassword },
    );
  }

  private applySession(response: AuthResponse): void {
    this.state.set({
      accessToken: response.access_token,
      user: response.user,
      activeMembership: response.active_membership,
      availableMemberships: response.available_memberships,
    });
  }

  private csrfHeaders(): HttpHeaders {
    const token = getCookie('csrf_token');
    if (!token) {
      return new HttpHeaders();
    }

    return new HttpHeaders({
      'X-CSRF-Token': token,
    });
  }
}
