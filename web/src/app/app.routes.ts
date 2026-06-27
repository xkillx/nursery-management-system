import { Routes } from '@angular/router';

import { authGuard } from './core/guards/auth.guard';
import { roleDefaultRedirectGuard } from './core/guards/role-default-redirect.guard';
import { roleGuard } from './core/guards/role.guard';
import { ManagerDashboardComponent } from './features/staff/pages/manager-dashboard/manager-dashboard.component';
import { ManagerChildrenComponent } from './features/staff/pages/manager-children/manager-children.component';
import { ManagerChildDetailComponent } from './features/staff/pages/manager-child-detail/manager-child-detail.component';
import { ManagerChildEditComponent } from './features/staff/pages/manager-child-edit/manager-child-edit.component';
import { ManagerInvitesComponent } from './features/staff/pages/manager-invites/manager-invites.component';
import { ManagerAttendanceCorrectionsComponent } from './features/staff/pages/manager-attendance-corrections/manager-attendance-corrections.component';
import { ManagerRoomsComponent } from './features/staff/pages/manager-rooms/manager-rooms.component';
import { ManagerSessionTypesComponent } from './features/staff/pages/manager-session-types/manager-session-types.component';
import { ManagerSessionTemplatesComponent } from './features/staff/pages/manager-session-templates/manager-session-templates.component';
import { ManagerBookingPatternComponent } from './features/staff/pages/manager-booking-pattern/manager-booking-pattern.component';
import { ManagerFundingOverviewComponent } from './features/staff/pages/manager-funding-overview/manager-funding-overview.component';
import { ManagerInvoiceRunComponent } from './features/staff/pages/manager-invoice-run/manager-invoice-run.component';
import { ManagerInvoicesComponent } from './features/staff/pages/manager-invoices/manager-invoices.component';
import { ManagerInvoiceDetailComponent } from './features/staff/pages/manager-invoice-detail/manager-invoice-detail.component';
import { PractitionerAttendanceChildrenComponent } from './features/staff/pages/practitioner-attendance-children/practitioner-attendance-children.component';
import { ParentInvoicesComponent } from './features/parent-portal/pages/parent-invoices/parent-invoices.component';
import { ParentInvoiceDetailComponent } from './features/parent-portal/pages/parent-invoice-detail/parent-invoice-detail.component';
import { OwnerOverviewComponent } from './features/owner/pages/owner-overview/owner-overview.component';
import { OwnerManagerAccessComponent } from './features/owner/pages/owner-manager-access/owner-manager-access.component';
import { OwnerRoomsComponent } from './features/owner/pages/owner-rooms/owner-rooms.component';
import { OwnerRoomFormComponent } from './features/owner/pages/owner-room-form/owner-room-form.component';
import { OwnerSessionTypesComponent } from './features/owner/pages/owner-session-types/owner-session-types.component';
import { OwnerSessionTypeFormComponent } from './features/owner/pages/owner-session-type-form/owner-session-type-form.component';
import { SignInComponent } from './pages/auth-pages/sign-in/sign-in.component';
import { SignUpComponent } from './pages/auth-pages/sign-up/sign-up.component';
import { ForgotPasswordComponent } from './pages/auth-pages/forgot-password/forgot-password.component';
import { ResetPasswordComponent } from './pages/auth-pages/reset-password/reset-password.component';
import { InviteAcceptComponent } from './pages/auth-pages/invite-accept/invite-accept.component';
import { NotFoundComponent } from './pages/other-page/not-found/not-found.component';
import { AppLayoutComponent } from './shared/layout/app-layout/app-layout.component';
import {
  childNameResolver,
  invoiceNumberResolver,
  ownerRoomNameResolver,
  parentInvoiceNumberResolver,
} from './core/navigation/breadcrumb-resolvers';

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
            path: 'manager/dashboard',
            component: ManagerDashboardComponent,
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['manager'],
              breadcrumb: { label: 'Manager' },
            },
            title: 'Manager Dashboard | Nursery Management',
          },
          {
            path: 'manager/children',
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['manager'],
              breadcrumb: { label: 'Children', link: ['/manager/children'] },
            },
            children: [
              {
                path: '',
                component: ManagerChildrenComponent,
                title: 'Manager Children | Nursery Management',
              },
              {
                path: 'new',
                component: ManagerChildEditComponent,
                data: {
                  breadcrumb: { label: 'Add child' },
                },
                title: 'Add Child | Nursery Management',
              },
              {
                path: ':childId',
                component: ManagerChildDetailComponent,
                data: {
                  breadcrumb: { label: 'Child', resolve: childNameResolver },
                },
                title: 'Child Enrollment | Nursery Management',
              },
              {
                path: ':childId/edit',
                component: ManagerChildEditComponent,
                data: {
                  breadcrumb: { label: 'Edit' },
                },
                title: 'Edit Child | Nursery Management',
              },
              {
                path: ':childId/booking-pattern',
                component: ManagerBookingPatternComponent,
                data: {
                  breadcrumb: { label: 'Session pattern' },
                },
                title: 'Session pattern (booking pattern) | Nursery Management',
              },
            ],
          },
          {
            path: 'manager/registrations',
            redirectTo: 'manager/children',
            pathMatch: 'full',
          },
          {
            path: 'manager/registrations/new',
            redirectTo: 'manager/children/new',
            pathMatch: 'full',
          },
          {
            path: 'manager/registrations/:childId/intake',
            redirectTo: 'manager/children/:childId/edit',
            pathMatch: 'full',
          },
          {
            path: 'manager/invites',
            component: ManagerInvitesComponent,
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['manager'],
              breadcrumb: { label: 'Invites' },
            },
            title: 'User Invites | Nursery Management',
          },
          {
            path: 'manager/attendance-corrections',
            component: ManagerAttendanceCorrectionsComponent,
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['manager'],
              breadcrumb: { label: 'Attendance corrections' },
            },
            title: 'Attendance Corrections | Nursery Management',
          },
          {
            path: 'manager/rooms',
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['manager'],
              breadcrumb: { label: 'Rooms', link: ['/manager/rooms'] },
            },
            children: [
              {
                path: '',
                component: ManagerRoomsComponent,
                title: 'Rooms | Nursery Management',
              },
              {
                path: 'new',
                component: OwnerRoomFormComponent,
                data: {
                  breadcrumb: { label: 'New room' },
                },
                title: 'New Room | Nursery Management',
              },
              {
                path: ':roomId/edit',
                component: OwnerRoomFormComponent,
                data: {
                  breadcrumb: { label: 'Edit room', resolve: ownerRoomNameResolver },
                },
                title: 'Edit Room | Nursery Management',
              },
            ],
          },
          {
            path: 'manager/session-types',
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['manager'],
              breadcrumb: { label: 'Session types', link: ['/manager/session-types'] },
            },
            children: [
              {
                path: '',
                component: ManagerSessionTypesComponent,
                title: 'Session types | Nursery Management',
              },
              {
                path: 'new',
                component: OwnerSessionTypeFormComponent,
                data: { breadcrumb: { label: 'New session type' } },
                title: 'New session type | Nursery Management',
              },
              {
                path: ':sessionTypeId/edit',
                component: OwnerSessionTypeFormComponent,
                data: { breadcrumb: { label: 'Edit session type' } },
                title: 'Edit session type | Nursery Management',
              },
            ],
          },
          {
            path: 'manager/session-templates',
            component: ManagerSessionTemplatesComponent,
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['manager'],
              breadcrumb: { label: 'Session templates' },
            },
            title: 'Session templates | Nursery Management',
          },
          {
            path: 'manager/funding',
            component: ManagerFundingOverviewComponent,
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['manager'],
              breadcrumb: { label: 'Funding' },
            },
            title: 'Funding Overview | Nursery Management',
          },
          {
            path: 'manager/invoice-run',
            component: ManagerInvoiceRunComponent,
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['manager'],
              breadcrumb: { label: 'Invoice run' },
            },
            title: 'Invoice Run | Nursery Management',
          },
          {
            path: 'manager/invoices',
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['manager'],
              breadcrumb: { label: 'Invoices', link: ['/manager/invoices'] },
            },
            children: [
              {
                path: '',
                component: ManagerInvoicesComponent,
                title: 'Invoices | Nursery Management',
              },
              {
                path: ':invoiceId',
                component: ManagerInvoiceDetailComponent,
                data: {
                  breadcrumb: { label: 'Invoice', resolve: invoiceNumberResolver },
                },
                title: 'Invoice Detail | Nursery Management',
              },
            ],
          },
          {
            path: 'practitioner/attendance',
            component: PractitionerAttendanceChildrenComponent,
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['practitioner'],
              breadcrumb: { label: 'Practitioner' },
            },
            title: 'Attendance | Nursery Management',
          },
          {
            path: 'manager/attendance',
            component: PractitionerAttendanceChildrenComponent,
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['manager'],
              breadcrumb: { label: 'Attendance' },
            },
            title: 'Attendance | Nursery Management',
          },
          {
            path: 'practitioner/attendance-children',
            pathMatch: 'full',
            redirectTo: 'practitioner/attendance',
          },
          {
            path: 'owner/rooms',
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['owner'],
              breadcrumb: { label: 'Rooms', link: ['/owner/rooms'] },
            },
            children: [
              {
                path: '',
                component: OwnerRoomsComponent,
                title: 'Room Management | Nursery Management',
              },
              {
                path: 'new',
                component: OwnerRoomFormComponent,
                data: {
                  breadcrumb: { label: 'New room' },
                },
                title: 'New Room | Nursery Management',
              },
              {
                path: ':roomId/edit',
                component: OwnerRoomFormComponent,
                data: {
                  breadcrumb: { label: 'Edit room', resolve: ownerRoomNameResolver },
                },
                title: 'Edit Room | Nursery Management',
              },
            ],
          },
          {
            path: 'owner/session-types',
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['owner'],
              breadcrumb: { label: 'Session types', link: ['/owner/session-types'] },
            },
            children: [
              {
                path: '',
                component: OwnerSessionTypesComponent,
                title: 'Session types | Nursery Management',
              },
              {
                path: 'new',
                component: OwnerSessionTypeFormComponent,
                data: { breadcrumb: { label: 'New session type' } },
                title: 'New session type | Nursery Management',
              },
              {
                path: ':sessionTypeId/edit',
                component: OwnerSessionTypeFormComponent,
                data: { breadcrumb: { label: 'Edit session type' } },
                title: 'Edit session type | Nursery Management',
              },
            ],
          },
          {
            path: 'owner',
            component: OwnerOverviewComponent,
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['owner'],
              breadcrumb: { label: 'Owner' },
            },
            title: 'Owner Overview | Nursery Management',
          },
          {
            path: 'owner/manager-access',
            component: OwnerManagerAccessComponent,
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['owner'],
              breadcrumb: { label: 'Manager access' },
            },
            title: 'Manager Access | Nursery Management',
          },
          {
            path: 'parent/invoices',
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['parent'],
              breadcrumb: { label: 'Billing', link: ['/parent/invoices'] },
            },
            children: [
              {
                path: '',
                component: ParentInvoicesComponent,
                title: 'Invoices | Nursery Management',
              },
              {
                path: ':invoiceId',
                component: ParentInvoiceDetailComponent,
                data: {
                  breadcrumb: { label: 'Invoice', resolve: parentInvoiceNumberResolver },
                },
                title: 'Invoice Detail | Nursery Management',
              },
            ],
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
