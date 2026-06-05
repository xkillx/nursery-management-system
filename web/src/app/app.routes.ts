import { Routes } from '@angular/router';

import { authGuard } from './core/guards/auth.guard';
import { roleDefaultRedirectGuard } from './core/guards/role-default-redirect.guard';
import { roleGuard } from './core/guards/role.guard';
import { ManagerDashboardComponent } from './features/staff/pages/manager-dashboard/manager-dashboard.component';
import { ManagerChildrenComponent } from './features/staff/pages/manager-children/manager-children.component';
import { ManagerGuardiansComponent } from './features/staff/pages/manager-guardians/manager-guardians.component';
import { PractitionerAttendanceChildrenComponent } from './features/staff/pages/practitioner-attendance-children/practitioner-attendance-children.component';
import { ParentInvoicesPlaceholderComponent } from './features/parent-portal/pages/parent-invoices-placeholder/parent-invoices-placeholder.component';
import { SignInComponent } from './pages/auth-pages/sign-in/sign-in.component';
import { SignUpComponent } from './pages/auth-pages/sign-up/sign-up.component';
import { ForgotPasswordComponent } from './pages/auth-pages/forgot-password/forgot-password.component';
import { ResetPasswordComponent } from './pages/auth-pages/reset-password/reset-password.component';
import { NotFoundComponent } from './pages/other-page/not-found/not-found.component';
import { AppLayoutComponent } from './shared/layout/app-layout/app-layout.component';

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
        path: 'staff/manager/guardians',
        component: ManagerGuardiansComponent,
        canActivate: [authGuard, roleGuard],
        data: { roles: ['manager'] },
        title: 'Manager Guardians | Nursery Management',
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
      {
        path: 'parent/invoices',
        component: ParentInvoicesPlaceholderComponent,
        canActivate: [authGuard, roleGuard],
        data: { roles: ['parent'] },
        title: 'Invoices | Nursery Management',
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
    path: '**',
    component: NotFoundComponent,
    title: 'Not Found | Nursery Management',
  },
];
