import { inject } from '@angular/core';
import { Observable, catchError, map, of } from 'rxjs';

import { CrumbResolver } from '../../shared/components/common/page-breadcrumb/page-breadcrumb.component';
import { OwnerApiService } from '../../features/owner/data/owner-api.service';
import { ParentInvoicesApiService } from '../../features/parent-portal/data/parent-invoices-api.service';
import { ManagerInvoicesApiService } from '../../features/staff/data/manager-invoices-api.service';
import { StaffApiService } from '../../features/staff/data/staff-api.service';
import { AuthService } from '../../core/services/auth.service';

const FALLBACK = 'Loading…';

function withFallback(value$: Observable<string>): Observable<string> {
  return value$.pipe(catchError(() => of(FALLBACK)));
}

export const childNameResolver: CrumbResolver = (_snapshot, _data, params) => {
  const id = params['childId'];
  if (!id) return of('Child');
  const api = inject(StaffApiService);
  return withFallback(api.getChild(id).pipe(map((c) => c.fullName || 'Child')));
};

export const invoiceNumberResolver: CrumbResolver = (_snapshot, _data, params) => {
  const id = params['invoiceId'];
  if (!id) return of('Invoice');
  const api = inject(ManagerInvoicesApiService);
  return withFallback(
    api.getInvoice(id).pipe(
      map((d) => d.invoiceNumberDisplay || d.invoiceNumber || 'Invoice'),
    ),
  );
};

export const parentInvoiceNumberResolver: CrumbResolver = (_snapshot, _data, params) => {
  const id = params['invoiceId'];
  if (!id) return of('Invoice');
  const api = inject(ParentInvoicesApiService);
  return withFallback(
    api.getInvoice(id).pipe(
      map((d) => d.invoiceNumberDisplay || d.invoiceNumber || 'Invoice'),
    ),
  );
};

export const ownerRoomNameResolver: CrumbResolver = (snapshot, _data, params, queryParams) => {
  const id = params['roomId'];
  if (!id) return of('Room');
  const api = inject(OwnerApiService);
  const auth = inject(AuthService);
  const querySiteId = queryParams['site_id'] ?? '';
  const branchSiteId = auth.activeMembership()?.branch_id ?? '';
  const siteId = querySiteId || branchSiteId;
  if (!siteId) {
    return of('Room');
  }
  return withFallback(api.getRoom(siteId, id).pipe(map((r) => r.name || 'Room')));
};
