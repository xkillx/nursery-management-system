import { Routes } from '@angular/router';

import { authGuard } from './core/guards/auth.guard';
import { roleDefaultRedirectGuard } from './core/guards/role-default-redirect.guard';
import { roleGuard } from './core/guards/role.guard';
import { ManagerDashboardComponent } from './features/staff/pages/manager-dashboard/manager-dashboard.component';
import { ManagerChildrenComponent } from './features/staff/pages/manager-children/manager-children.component';
import { ManagerChildDetailComponent } from './features/staff/pages/manager-child-detail/manager-child-detail.component';
import { ManagerGuardiansComponent } from './features/staff/pages/manager-guardians/manager-guardians.component';
import { ManagerInvitesComponent } from './features/staff/pages/manager-invites/manager-invites.component';
import { ManagerAttendanceCorrectionsComponent } from './features/staff/pages/manager-attendance-corrections/manager-attendance-corrections.component';
import { ManagerFundingOverviewComponent } from './features/staff/pages/manager-funding-overview/manager-funding-overview.component';
import { ManagerInvoiceRunComponent } from './features/staff/pages/manager-invoice-run/manager-invoice-run.component';
import { ManagerInvoicesComponent } from './features/staff/pages/manager-invoices/manager-invoices.component';
import { ManagerInvoiceDetailComponent } from './features/staff/pages/manager-invoice-detail/manager-invoice-detail.component';
import { PractitionerAttendanceChildrenComponent } from './features/staff/pages/practitioner-attendance-children/practitioner-attendance-children.component';
import { ParentInvoicesComponent } from './features/parent-portal/pages/parent-invoices/parent-invoices.component';
import { ParentInvoiceDetailComponent } from './features/parent-portal/pages/parent-invoice-detail/parent-invoice-detail.component';
import { SignInComponent } from './pages/auth-pages/sign-in/sign-in.component';
import { SignUpComponent } from './pages/auth-pages/sign-up/sign-up.component';
import { ForgotPasswordComponent } from './pages/auth-pages/forgot-password/forgot-password.component';
import { ResetPasswordComponent } from './pages/auth-pages/reset-password/reset-password.component';
import { InviteAcceptComponent } from './pages/auth-pages/invite-accept/invite-accept.component';
import { NotFoundComponent } from './pages/other-page/not-found/not-found.component';
import { AppLayoutComponent } from './shared/layout/app-layout/app-layout.component';
import { ParentPortalLayoutComponent } from './shared/layout/parent-portal-layout/parent-portal-layout.component';

export const routes: Routes = [
  {
    path: '',
    component: AppLayoutComponent,
    canActivate: [authGuard, roleDefaultRedirectGuard],
    pathMatch: 'full',
  },
  {
    path: '',
    component: AppLayoutComponent,
    children: [
      {
        path: 'staff/manager/dashboard',
        component: ManagerDashboardComponent,
        canActivate: [authGuard, roleGuard],
        data: { roles: ['manager'] },
        title: 'Manager Dashboard | Nursery Management',
      },
      {
        path: 'staff/manager/children',
        component: ManagerChildrenComponent,
        canActivate: [authGuard, roleGuard],
        data: { roles: ['manager'] },
        title: 'Manager Children | Nursery Management',
      },
      {
        path: 'staff/manager/children/:childId',
        component: ManagerChildDetailComponent,
        canActivate: [authGuard, roleGuard],
        data: { roles: ['manager'] },
        title: 'Child Enrollment | Nursery Management',
      },
      {
        path: 'staff/manager/guardians',
        component: ManagerGuardiansComponent,
        canActivate: [authGuard, roleGuard],
        data: { roles: ['manager'] },
        title: 'Manager Guardians | Nursery Management',
      },
      {
        path: 'staff/manager/invites',
        component: ManagerInvitesComponent,
        canActivate: [authGuard, roleGuard],
        data: { roles: ['manager'] },
        title: 'User Invites | Nursery Management',
      },
      {
        path: 'staff/manager/attendance-corrections',
        component: ManagerAttendanceCorrectionsComponent,
        canActivate: [authGuard, roleGuard],
        data: { roles: ['manager'] },
        title: 'Attendance Corrections | Nursery Management',
      },
      {
        path: 'staff/manager/funding',
        component: ManagerFundingOverviewComponent,
        canActivate: [authGuard, roleGuard],
        data: { roles: ['manager'] },
        title: 'Funding Overview | Nursery Management',
      },
      {
        path: 'staff/manager/invoice-run',
        component: ManagerInvoiceRunComponent,
        canActivate: [authGuard, roleGuard],
        data: { roles: ['manager'] },
        title: 'Invoice Run | Nursery Management',
      },
      {
        path: 'staff/manager/invoices',
        component: ManagerInvoicesComponent,
        canActivate: [authGuard, roleGuard],
        data: { roles: ['manager'] },
        title: 'Invoices | Nursery Management',
      },
      {
        path: 'staff/manager/invoices/:invoiceId',
        component: ManagerInvoiceDetailComponent,
        canActivate: [authGuard, roleGuard],
        data: { roles: ['manager'] },
        title: 'Invoice Detail | Nursery Management',
      },
      {
        path: 'staff/practitioner/attendance',
        component: PractitionerAttendanceChildrenComponent,
        canActivate: [authGuard, roleGuard],
        data: { roles: ['manager', 'practitioner'] },
        title: 'Attendance | Nursery Management',
      },
      {
        path: 'staff/practitioner/attendance-children',
        pathMatch: 'full',
        redirectTo: 'staff/practitioner/attendance',
      },
    ],
  },
  {
    path: '',
    component: ParentPortalLayoutComponent,
    children: [
      {
        path: 'parent/invoices',
        component: ParentInvoicesComponent,
        canActivate: [authGuard, roleGuard],
        data: { roles: ['parent'] },
        title: 'Invoices | Nursery Management',
      },
      {
        path: 'parent/invoices/:invoiceId',
        component: ParentInvoiceDetailComponent,
        canActivate: [authGuard, roleGuard],
        data: { roles: ['parent'] },
        title: 'Invoice Detail | Nursery Management',
      },
    ],
  },
  {
    path: 'signin',
    component: SignInComponent,
    title: 'Sign In | Nursery Management',
  },
  {
    path: 'signup',
    component: SignUpComponent,
    title: 'Invitation Only | Nursery Management',
  },
  {
    path: 'forgot-password',
    component: ForgotPasswordComponent,
    title: 'Forgot Password | Nursery Management',
  },
  {
    path: 'reset-password',
    component: ResetPasswordComponent,
    title: 'Reset Password | Nursery Management',
  },
  {
    path: 'invite-accept',
    component: InviteAcceptComponent,
    title: 'Accept Invitation | Nursery Management',
  },
  {
    path: '**',
    component: NotFoundComponent,
    title: 'Not Found | Nursery Management',
  },
];
