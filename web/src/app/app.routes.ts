import { Routes } from '@angular/router';

import { authGuard } from './core/guards/auth.guard';
import { roleGuard } from './core/guards/role.guard';
import { ManagerChildrenComponent } from './features/staff/pages/manager-children/manager-children.component';
import { ManagerGuardiansComponent } from './features/staff/pages/manager-guardians/manager-guardians.component';
import { PractitionerAttendanceChildrenComponent } from './features/staff/pages/practitioner-attendance-children/practitioner-attendance-children.component';
import { SignInComponent } from './pages/auth-pages/sign-in/sign-in.component';
import { SignUpComponent } from './pages/auth-pages/sign-up/sign-up.component';
import { BlankComponent } from './pages/blank/blank.component';
import { CalenderComponent } from './pages/calender/calender.component';
import { BarChartComponent } from './pages/charts/bar-chart/bar-chart.component';
import { LineChartComponent } from './pages/charts/line-chart/line-chart.component';
import { EcommerceComponent } from './pages/dashboard/ecommerce/ecommerce.component';
import { FormElementsComponent } from './pages/forms/form-elements/form-elements.component';
import { InvoicesComponent } from './pages/invoices/invoices.component';
import { NotFoundComponent } from './pages/other-page/not-found/not-found.component';
import { ProfileComponent } from './pages/profile/profile.component';
import { BasicTablesComponent } from './pages/tables/basic-tables/basic-tables.component';
import { AlertsComponent } from './pages/ui-elements/alerts/alerts.component';
import { AvatarElementComponent } from './pages/ui-elements/avatar-element/avatar-element.component';
import { BadgesComponent } from './pages/ui-elements/badges/badges.component';
import { ButtonsComponent } from './pages/ui-elements/buttons/buttons.component';
import { ImagesComponent } from './pages/ui-elements/images/images.component';
import { VideosComponent } from './pages/ui-elements/videos/videos.component';
import { AppLayoutComponent } from './shared/layout/app-layout/app-layout.component';

export const routes: Routes = [
  {
    path: '',
    component: AppLayoutComponent,
    children: [
      {
        path: '',
        component: EcommerceComponent,
        pathMatch: 'full',
        title: 'Angular Ecommerce Dashboard | TailAdmin - Angular Admin Dashboard Template',
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
        path: 'staff/practitioner/attendance-children',
        component: PractitionerAttendanceChildrenComponent,
        canActivate: [authGuard, roleGuard],
        data: { roles: ['manager', 'practitioner'] },
        title: 'Attendance Children | Nursery Management',
      },
      {
        path: 'calendar',
        component: CalenderComponent,
        title: 'Angular Calender | TailAdmin - Angular Admin Dashboard Template',
      },
      {
        path: 'profile',
        component: ProfileComponent,
        title: 'Angular Profile Dashboard | TailAdmin - Angular Admin Dashboard Template',
      },
      {
        path: 'form-elements',
        component: FormElementsComponent,
        title: 'Angular Form Elements Dashboard | TailAdmin - Angular Admin Dashboard Template',
      },
      {
        path: 'basic-tables',
        component: BasicTablesComponent,
        title: 'Angular Basic Tables Dashboard | TailAdmin - Angular Admin Dashboard Template',
      },
      {
        path: 'blank',
        component: BlankComponent,
        title: 'Angular Blank Dashboard | TailAdmin - Angular Admin Dashboard Template',
      },
      {
        path: 'invoice',
        component: InvoicesComponent,
        title: 'Angular Invoice Details Dashboard | TailAdmin - Angular Admin Dashboard Template',
      },
      {
        path: 'line-chart',
        component: LineChartComponent,
        title: 'Angular Line Chart Dashboard | TailAdmin - Angular Admin Dashboard Template',
      },
      {
        path: 'bar-chart',
        component: BarChartComponent,
        title: 'Angular Bar Chart Dashboard | TailAdmin - Angular Admin Dashboard Template',
      },
      {
        path: 'alerts',
        component: AlertsComponent,
        title: 'Angular Alerts Dashboard | TailAdmin - Angular Admin Dashboard Template',
      },
      {
        path: 'avatars',
        component: AvatarElementComponent,
        title: 'Angular Avatars Dashboard | TailAdmin - Angular Admin Dashboard Template',
      },
      {
        path: 'badge',
        component: BadgesComponent,
        title: 'Angular Badges Dashboard | TailAdmin - Angular Admin Dashboard Template',
      },
      {
        path: 'buttons',
        component: ButtonsComponent,
        title: 'Angular Buttons Dashboard | TailAdmin - Angular Admin Dashboard Template',
      },
      {
        path: 'images',
        component: ImagesComponent,
        title: 'Angular Images Dashboard | TailAdmin - Angular Admin Dashboard Template',
      },
      {
        path: 'videos',
        component: VideosComponent,
        title: 'Angular Videos Dashboard | TailAdmin - Angular Admin Dashboard Template',
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
    title: 'Angular Sign Up Dashboard | TailAdmin - Angular Admin Dashboard Template',
  },
  {
    path: '**',
    component: NotFoundComponent,
    title: 'Angular NotFound Dashboard | TailAdmin - Angular Admin Dashboard Template',
  },
];
