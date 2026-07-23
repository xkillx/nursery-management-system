import { Routes } from '@angular/router';

import { authGuard } from './core/guards/auth.guard';
import { roleDefaultRedirectGuard } from './core/guards/role-default-redirect.guard';
import { roleGuard } from './core/guards/role.guard';
import { ManagerDashboardComponent } from './features/staff/pages/manager-dashboard/manager-dashboard.component';
import { ManagerChildrenComponent } from './features/staff/pages/manager-children/manager-children.component';
import { ManagerChildDetailComponent } from './features/staff/pages/manager-child-detail/manager-child-detail.component';
import { ManagerChildEditComponent } from './features/staff/pages/manager-child-edit/manager-child-edit.component';
import { ManagerParentsComponent } from './features/staff/pages/manager-parents/manager-parents.component';
import { ManagerParentDetailComponent } from './features/staff/pages/manager-parent-detail/manager-parent-detail.component';
import { ManagerParentNewComponent } from './features/staff/pages/manager-parent-new/manager-parent-new.component';
import { ManagerParentEditComponent } from './features/staff/pages/manager-parent-edit/manager-parent-edit.component';
import { ManagerInvitesComponent } from './features/staff/pages/manager-invites/manager-invites.component';
import { ManagerAttendanceCorrectionsComponent } from './features/staff/pages/manager-attendance-corrections/manager-attendance-corrections.component';
import { ManagerRoomsComponent } from './features/staff/pages/manager-rooms/manager-rooms.component';
import { ManagerSessionTypesComponent } from './features/staff/pages/manager-session-types/manager-session-types.component';
import { ManagerSessionTemplatesComponent } from './features/staff/pages/manager-session-templates/manager-session-templates.component';
import { ManagerSessionTemplateFormComponent } from './features/staff/pages/manager-session-template-form/manager-session-template-form.component';
import { ManagerFundingOverviewComponent } from './features/staff/pages/manager-funding-overview/manager-funding-overview.component';
import { ManagerInvoiceCreateComponent } from './features/staff/pages/manager-invoice-create/manager-invoice-create.component';
import { ManagerInvoiceEditComponent } from './features/staff/pages/manager-invoice-edit/manager-invoice-edit.component';
import { ManagerInvoicesComponent } from './features/staff/pages/manager-invoices/manager-invoices.component';
import { ManagerBillingSetupComponent } from './features/staff/pages/manager-billing-setup/manager-billing-setup.component';
import { ManagerSiteSettingsComponent } from './features/staff/pages/manager-site-settings/manager-site-settings.component';
import { ManagerSiteProfileComponent } from './features/staff/pages/manager-site-profile/manager-site-profile.component';
import { ManagerInvoiceDetailComponent } from './features/staff/pages/manager-invoice-detail/manager-invoice-detail.component';
import { PractitionerAttendanceChildrenComponent } from './features/staff/pages/practitioner-attendance-children/practitioner-attendance-children.component';
import { ManagerTermCalendarComponent } from './features/staff/pages/manager-term-calendar/manager-term-calendar.component';
import { ManagerClosureDaysComponent } from './features/staff/pages/manager-closure-days/manager-closure-days.component';
import { ManagerBookingsComponent } from './features/staff/pages/manager-bookings/manager-bookings.component';
import { CreateRecurringBookingComponent } from './features/staff/pages/create-recurring-booking/create-recurring-booking.component';
import { CreateAdHocBookingComponent } from './features/staff/pages/create-ad-hoc-booking/create-ad-hoc-booking.component';
import { CreateHourlyBookingComponent } from './features/staff/pages/create-hourly-booking/create-hourly-booking.component';
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
                data: {
                  breadcrumb: { label: 'Child', resolve: childNameResolver },
                },
                children: [
                  {
                    path: '',
                    component: ManagerChildDetailComponent,
                    title: 'Child Enrollment | Nursery Management',
                  },
                  {
                    path: 'edit',
                    component: ManagerChildEditComponent,
                    data: { breadcrumb: { label: 'Edit' } },
                    title: 'Edit Child | Nursery Management',
                  },
                  {
                    path: ':tab',
                    component: ManagerChildDetailComponent,
                    title: 'Child Enrollment | Nursery Management',
                  },
                ],
              },
            ],
          },
          {
            path: 'manager/parents',
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['manager'],
              breadcrumb: { label: 'Parents', link: ['/manager/parents'] },
            },
            children: [
              {
                path: '',
                component: ManagerParentsComponent,
                title: 'Parents | Nursery Management',
              },
              {
                path: 'new',
                component: ManagerParentNewComponent,
                data: {
                  breadcrumb: { label: 'Add parent' },
                },
                title: 'Add Parent | Nursery Management',
              },
              {
                path: ':parentId',
                children: [
                  {
                    path: '',
                    component: ManagerParentDetailComponent,
                    data: {
                      breadcrumb: { label: 'Parent' },
                    },
                    title: 'Parent Details | Nursery Management',
                  },
                  {
                    path: 'edit',
                    component: ManagerParentEditComponent,
                    data: { breadcrumb: { label: 'Edit' } },
                    title: 'Edit Parent | Nursery Management',
                  },
                ],
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
            path: 'manager/site-settings',
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['manager'],
              breadcrumb: { label: 'Site settings', link: ['/manager/site-settings'] },
            },
            children: [
              {
                path: '',
                component: ManagerSiteSettingsComponent,
                title: 'Site Settings | Nursery Management',
              },
              {
                path: 'profile',
                component: ManagerSiteProfileComponent,
                data: { breadcrumb: { label: 'Site profile' } },
                title: 'Site Profile | Nursery Management',
              },
              {
                path: 'rooms',
                data: { breadcrumb: { label: 'Rooms & capacity' } },
                children: [
                  {
                    path: '',
                    component: ManagerRoomsComponent,
                    title: 'Rooms | Nursery Management',
                  },
                  {
                    path: 'new',
                    component: OwnerRoomFormComponent,
                    data: { breadcrumb: { label: 'New room' } },
                    title: 'New Room | Nursery Management',
                  },
                  {
                    path: ':roomId/edit',
                    component: OwnerRoomFormComponent,
                    data: { breadcrumb: { label: 'Edit room', resolve: ownerRoomNameResolver } },
                    title: 'Edit Room | Nursery Management',
                  },
                ],
              },
              {
                path: 'session-types',
                data: { breadcrumb: { label: 'Session types' } },
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
                path: 'billing-setup',
                component: ManagerBillingSetupComponent,
                canActivate: [authGuard, roleGuard],
                data: {
                  roles: ['manager'],
                  breadcrumb: { label: 'Fees & billing' },
                },
                title: 'Billing Setup | Nursery Management',
              },
              {
                path: 'term-calendar',
                component: ManagerTermCalendarComponent,
                data: { breadcrumb: { label: 'Term calendar' } },
                title: 'Term Calendar | Nursery Management',
              },
              {
                path: 'closure-days',
                component: ManagerClosureDaysComponent,
                data: { breadcrumb: { label: 'Closure days' } },
                title: 'Closure Days | Nursery Management',
              },
              {
                path: 'session-templates',
                data: { breadcrumb: { label: 'Session templates' } },
                children: [
                  {
                    path: '',
                    component: ManagerSessionTemplatesComponent,
                    title: 'Session templates | Nursery Management',
                  },
                  {
                    path: 'new',
                    component: ManagerSessionTemplateFormComponent,
                    data: { breadcrumb: { label: 'New template' } },
                    title: 'New template | Nursery Management',
                  },
                  {
                    path: ':templateId/edit',
                    component: ManagerSessionTemplateFormComponent,
                    data: { breadcrumb: { label: 'Edit template' } },
                    title: 'Edit template | Nursery Management',
                  },
                ],
              },
            ],
          },
          {
            path: 'manager/bookings/new/recurring',
            component: CreateRecurringBookingComponent,
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['manager'],
              breadcrumb: { label: 'New Recurring Booking', parent: '/manager/bookings' },
            },
            title: 'New Recurring Booking | Nursery Management',
          },
          {
            path: 'manager/bookings/new/ad_hoc',
            component: CreateAdHocBookingComponent,
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['manager'],
              breadcrumb: { label: 'New Ad-Hoc Booking', parent: '/manager/bookings' },
            },
            title: 'New Ad-Hoc Booking | Nursery Management',
          },
          {
            path: 'manager/bookings/new/hourly',
            component: CreateHourlyBookingComponent,
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['manager'],
              breadcrumb: { label: 'New Hourly Booking', parent: '/manager/bookings' },
            },
            title: 'New Hourly Booking | Nursery Management',
          },
          {
            path: 'manager/bookings',
            component: ManagerBookingsComponent,
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['manager'],
              breadcrumb: { label: 'Bookings' },
            },
            title: 'Bookings | Nursery Management',
          },
          {
            path: 'manager/ad-hoc-bookings',
            redirectTo: '/manager/bookings',
            pathMatch: 'full',
          },
          {
            path: 'manager/hourly-bookings',
            redirectTo: '/manager/bookings',
            pathMatch: 'full',
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
                path: 'new',
                component: ManagerInvoiceCreateComponent,
                data: {
                  breadcrumb: { label: 'Create Invoice' },
                },
                title: 'Create Invoice | Nursery Management',
              },
              {
                path: ':invoiceId',
                component: ManagerInvoiceDetailComponent,
                data: {
                  breadcrumb: { label: 'Invoice', resolve: invoiceNumberResolver },
                },
                title: 'Invoice Detail | Nursery Management',
              },
              {
                path: ':invoiceId/edit',
                component: ManagerInvoiceEditComponent,
                data: {
                  breadcrumb: { label: 'Edit invoice' },
                },
                title: 'Edit Invoice | Nursery Management',
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
          // Redirects from old URLs to new nested site-settings structure
          {
            path: 'manager/site-profile',
            redirectTo: '/manager/site-settings/profile',
            pathMatch: 'full',
          },
          {
            path: 'manager/billing-setup',
            redirectTo: '/manager/site-settings/billing-setup',
            pathMatch: 'full',
          },
          {
            path: 'manager/rooms',
            redirectTo: '/manager/site-settings/rooms',
            pathMatch: 'full',
          },
          {
            path: 'manager/rooms/new',
            redirectTo: '/manager/site-settings/rooms/new',
            pathMatch: 'full',
          },
          {
            path: 'manager/rooms/:roomId/edit',
            redirectTo: '/manager/site-settings/rooms/:roomId/edit',
            pathMatch: 'full',
          },
          {
            path: 'manager/session-types',
            redirectTo: '/manager/site-settings/session-types',
            pathMatch: 'full',
          },
          {
            path: 'manager/session-types/new',
            redirectTo: '/manager/site-settings/session-types/new',
            pathMatch: 'full',
          },
          {
            path: 'manager/session-types/:sessionTypeId/edit',
            redirectTo: '/manager/site-settings/session-types/:sessionTypeId/edit',
            pathMatch: 'full',
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
          {
            path: 'parent/funding',
            canActivate: [authGuard, roleGuard],
            data: {
              roles: ['parent'],
              breadcrumb: { label: 'Funding', link: ['/parent/funding'] },
            },
            loadComponent: () => import('./features/parent-portal/pages/parent-funding/parent-funding.component').then((m) => m.ParentFundingComponent),
            title: 'Funding | Nursery Management',
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
