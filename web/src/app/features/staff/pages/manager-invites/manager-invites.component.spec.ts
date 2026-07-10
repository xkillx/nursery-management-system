import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';

import { InviteRecord } from '../../models/invites.models';
import { ToastService } from '../../../../shared/services/toast.service';
import { ManagerInvitesComponent } from './manager-invites.component';

describe('ManagerInvitesComponent', () => {
  let fixture: ComponentFixture<ManagerInvitesComponent>;
  let component: ManagerInvitesComponent;
  let httpMock: HttpTestingController;
  let toastService: ToastService;

  const pendingInviteRecord: InviteRecord = {
    id: 'invite-1',
    email: 'staff@example.com',
    role: 'practitioner',
    status: 'pending',
    expiresAt: '2026-06-13T00:00:00Z',
    acceptedAt: null,
    revokedAt: null,
    createdAt: '2026-06-06T10:00:00Z',
    updatedAt: '2026-06-06T10:00:00Z',
  };

  const pendingInviteApi = {
    id: 'invite-1',
    email: 'staff@example.com',
    role: 'practitioner',
    status: 'pending',
    expires_at: '2026-06-13T00:00:00Z',
    accepted_at: null,
    revoked_at: null,
    created_at: '2026-06-06T10:00:00Z',
    updated_at: '2026-06-06T10:00:00Z',
  };

  const acceptedInviteRecord: InviteRecord = {
    ...pendingInviteRecord,
    id: 'invite-2',
    email: 'parent@example.com',
    role: 'parent',
    status: 'accepted',
    acceptedAt: '2026-06-07T09:00:00Z',
  };

  const acceptedInviteApi = {
    ...pendingInviteApi,
    id: 'invite-2',
    email: 'parent@example.com',
    role: 'parent',
    status: 'accepted',
    accepted_at: '2026-06-07T09:00:00Z',
  };

  const revokedInvite: InviteRecord = {
    ...pendingInviteRecord,
    id: 'invite-3',
    status: 'revoked',
    revokedAt: '2026-06-07T10:00:00Z',
  };

  const expiredInvite: InviteRecord = {
    ...pendingInviteRecord,
    id: 'invite-4',
    status: 'expired',
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ManagerInvitesComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    }).compileComponents();

    fixture = TestBed.createComponent(ManagerInvitesComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
    toastService = TestBed.inject(ToastService);
  });

  afterEach(() => {
    httpMock.verify();
  });

  function flushInvites(items: Record<string, unknown>[] = []): void {
    const req = httpMock.expectOne((r) => r.url === '/api/v1/invites' && r.params.get('status') === 'pending');
    req.flush({ items });
  }

  it('loads pending invites on init', () => {
    fixture.detectChanges();

    flushInvites([pendingInviteApi]);
    fixture.detectChanges();

    expect(component.invites).toEqual([pendingInviteRecord]);
    expect(component.isLoading).toBe(false);
  });

  it('role options contain practitioner and parent but not manager', () => {
    const roleValues = component.roleOptions.map((o) => o.value) as string[];
    expect(roleValues).toContain('practitioner');
    expect(roleValues).toContain('parent');
    expect(roleValues).not.toContain('manager');
  });

  it('submits invite with trimmed email and selected role', () => {
    component.email = '  new@example.com  ';
    component.role = 'parent';
    fixture.detectChanges();
    flushInvites();

    component.submitInvite();

    const createReq = httpMock.expectOne((r) => r.url === '/api/v1/invites' && r.method === 'POST');
    expect(createReq.request.body).toEqual({ email: 'new@example.com', role: 'parent' });

    spyOn(toastService, 'success');
    createReq.flush(pendingInviteApi);

    const listReq = httpMock.expectOne((r) => r.url === '/api/v1/invites');
    listReq.flush({ items: [pendingInviteApi] });

    expect(component.email).toBe('');
    expect(component.role).toBe('practitioner');
    expect(toastService.success).toHaveBeenCalledWith('Invitation pending for new@example.com.');
  });

  it('status filter change reloads invites with selected status', () => {
    fixture.detectChanges();
    flushInvites();

    component.setStatusFilter('all');

    const req = httpMock.expectOne((r) => r.url === '/api/v1/invites' && r.params.get('status') === 'all');
    req.flush({ items: [pendingInviteApi, acceptedInviteApi] });

    expect(component.invites.length).toBe(2);
  });

  it('pending invite rows allow resend and revoke', () => {
    expect(component.canAct(pendingInviteRecord)).toBe(true);
  });

  it('accepted invite rows do not allow actions', () => {
    expect(component.canAct(acceptedInviteRecord)).toBe(false);
  });

  it('revoked invite rows do not allow actions', () => {
    expect(component.canAct(revokedInvite)).toBe(false);
  });

  it('expired invite rows do not allow actions', () => {
    expect(component.canAct(expiredInvite)).toBe(false);
  });

  it('resend calls service only for pending rows', () => {
    fixture.detectChanges();
    flushInvites([pendingInviteApi]);

    spyOn(toastService, 'success');
    component.resend(pendingInviteRecord);

    const req = httpMock.expectOne('/api/v1/invites/invite-1/resend');
    req.flush(pendingInviteApi);

    const listReq = httpMock.expectOne((r) => r.url === '/api/v1/invites');
    listReq.flush({ items: [pendingInviteApi] });

    expect(toastService.success).toHaveBeenCalledWith('Invitation resent to staff@example.com.');
  });

  it('resend no-ops for non-pending rows', () => {
    component.resend(acceptedInviteRecord);
    httpMock.expectNone('/api/v1/invites/invite-2/resend');
    expect(component.pendingInviteIds.has(acceptedInviteRecord.id)).toBe(false);
  });

  it('revoke opens confirmation and calls service after confirm', () => {
    fixture.detectChanges();
    flushInvites([pendingInviteApi]);

    component.openRevoke(pendingInviteRecord);
    expect(component.inviteToRevoke).toEqual(pendingInviteRecord);

    spyOn(toastService, 'success');
    component.confirmRevoke();

    const req = httpMock.expectOne('/api/v1/invites/invite-1/revoke');
    req.flush({ ...pendingInviteApi, status: 'revoked', revoked_at: '2026-06-06T12:00:00Z' });

    const listReq = httpMock.expectOne((r) => r.url === '/api/v1/invites');
    listReq.flush({ items: [] });

    expect(component.inviteToRevoke).toBeNull();
    expect(toastService.success).toHaveBeenCalledWith('Invitation revoked for staff@example.com.');
  });

  it('revoke no-ops for non-pending rows', () => {
    component.openRevoke(acceptedInviteRecord);
    expect(component.inviteToRevoke).toBeNull();
  });

  it('cancelRevoke clears inviteToRevoke', () => {
    component.inviteToRevoke = pendingInviteRecord;
    component.cancelRevoke();
    expect(component.inviteToRevoke).toBeNull();
  });

  it('renders API error message with request ID', () => {
    fixture.detectChanges();
    flushInvites([pendingInviteApi]);

    component.resend(pendingInviteRecord);

    const req = httpMock.expectOne('/api/v1/invites/invite-1/resend');
    req.flush(
      { code: 'internal_error', message: 'Server error', request_id: 'req-abc' },
      { status: 500, statusText: 'Internal Server Error' },
    );

    expect(component.rowErrors['invite-1']).toContain('Something went wrong');
    expect(component.rowErrors['invite-1']).toContain('Request: req-abc');
  });

  it('renders field-level errors on create', () => {
    fixture.detectChanges();
    flushInvites();

    component.email = 'taken@example.com';
    component.submitInvite();

    const req = httpMock.expectOne((r) => r.url === '/api/v1/invites' && r.method === 'POST');
    req.flush(
      { code: 'validation_error', message: 'Email already has a pending invite.', details: { field: 'email' } },
      { status: 422, statusText: 'Unprocessable Entity' },
    );

    expect(component.fieldErrors['email']).toBe('Email already has a pending invite.');
  });
});
