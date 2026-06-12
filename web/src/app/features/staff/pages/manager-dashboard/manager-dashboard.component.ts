import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';

import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { TableShellComponent } from '../../../../shared/components/ui/table/table-shell.component';
import {
  MANAGER_DASHBOARD_MOCK,
  ManagerDashboardSnapshot,
  PaymentFollowUpInvoice,
  formatGbp,
  sortPaymentFollowUp,
} from './manager-dashboard.models';

type AttendanceTileTone = 'success' | 'warning' | 'neutral';

interface AttendanceTile {
  label: string;
  value: number;
  meta: string;
  tone: AttendanceTileTone;
  description: string;
}

@Component({
  selector: 'app-manager-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    EmptyStateComponent,
    PageHeaderComponent,
    StatusBadgeComponent,
    TableShellComponent,
  ],
  template: `
    <div class="space-y-6">
      <app-page-header
        title="Manager operations"
        description="Live branch health for attendance, registrations, billing, and payment follow-up."
        [hasActions]="true"
      >
        <a
          actions
          [routerLink]="attendanceRoute"
          class="inline-flex min-h-11 w-full items-center justify-center rounded-lg bg-brand-500 px-4 py-2.5 text-sm font-medium text-white shadow-theme-xs transition hover:bg-brand-600 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-500 sm:w-auto"
        >
          Open attendance
        </a>
      </app-page-header>

      <section aria-labelledby="attendance-heading" class="space-y-3">
        <div class="flex flex-col gap-1 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p class="text-xs font-medium uppercase text-gray-500 dark:text-gray-400">Today</p>
            <h2 id="attendance-heading" class="text-lg font-semibold text-gray-800 dark:text-white/90">Attendance snapshot</h2>
          </div>
          <p class="text-sm text-gray-500 dark:text-gray-400">
            {{ attendanceCompletionRate }}% attendance completion across {{ totalChildrenToday }} expected children.
          </p>
        </div>

        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
          @for (tile of attendanceTiles; track tile.label) {
            <article class="rounded-2xl border border-gray-200 bg-white p-5 shadow-theme-xs transition hover:-translate-y-0.5 hover:shadow-theme-sm dark:border-gray-800 dark:bg-white/[0.03]">
              <div class="flex items-start justify-between gap-3">
                <div>
                  <p class="text-3xl font-semibold tabular-nums text-gray-900 dark:text-white">{{ tile.value }}</p>
                  <p class="mt-1 text-sm font-medium text-gray-700 dark:text-gray-200">{{ tile.label }}</p>
                </div>
                <span
                  class="rounded-full px-2.5 py-1 text-xs font-medium"
                  [ngClass]="tile.tone === 'warning'
                    ? 'bg-warning-50 text-warning-700 dark:bg-warning-500/15 dark:text-warning-400'
                    : tile.tone === 'success'
                      ? 'bg-success-50 text-success-700 dark:bg-success-500/15 dark:text-success-500'
                      : 'bg-gray-100 text-gray-700 dark:bg-white/5 dark:text-white/80'"
                >
                  {{ tile.meta }}
                </span>
              </div>
              <p class="mt-4 text-sm text-gray-500 dark:text-gray-400">{{ tile.description }}</p>
            </article>
          }
        </div>
      </section>

      <section aria-labelledby="invoice-heading" class="rounded-2xl border border-gray-200 bg-white p-5 shadow-theme-xs dark:border-gray-800 dark:bg-white/[0.03]">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <p class="text-xs font-medium uppercase text-gray-500 dark:text-gray-400">Billing month</p>
            <h2 id="invoice-heading" class="text-lg font-semibold text-gray-800 dark:text-white/90">
              {{ snapshot.invoiceRunStatus.billingMonthLabel }} invoice run
            </h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ snapshot.invoiceRunStatus.nextStep }}</p>
          </div>
          <a
            [routerLink]="invoiceRunRoute"
            class="inline-flex min-h-11 items-center justify-center rounded-lg border border-gray-300 bg-white px-4 py-2.5 text-sm font-medium text-gray-700 transition hover:bg-gray-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-500 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-white/[0.05]"
          >
            Review invoice run
          </a>
        </div>

        <dl class="mt-5 grid grid-cols-2 gap-3 md:grid-cols-5">
          <div class="rounded-xl bg-gray-50 p-4 dark:bg-white/[0.04]">
            <dt class="text-xs font-medium uppercase text-gray-500 dark:text-gray-400">Eligible</dt>
            <dd class="mt-1 text-2xl font-semibold tabular-nums text-gray-900 dark:text-white">{{ snapshot.invoiceRunStatus.eligibleChildren }}</dd>
          </div>
          <div class="rounded-xl bg-gray-50 p-4 dark:bg-white/[0.04]">
            <dt class="text-xs font-medium uppercase text-gray-500 dark:text-gray-400">Blocked</dt>
            <dd class="mt-1 text-2xl font-semibold tabular-nums text-warning-700 dark:text-warning-400">{{ snapshot.invoiceRunStatus.blockedChildren }}</dd>
          </div>
          <div class="rounded-xl bg-gray-50 p-4 dark:bg-white/[0.04]">
            <dt class="text-xs font-medium uppercase text-gray-500 dark:text-gray-400">Drafts</dt>
            <dd class="mt-1 text-2xl font-semibold tabular-nums text-gray-900 dark:text-white">{{ snapshot.invoiceRunStatus.draftInvoices }}</dd>
          </div>
          <div class="rounded-xl bg-gray-50 p-4 dark:bg-white/[0.04]">
            <dt class="text-xs font-medium uppercase text-gray-500 dark:text-gray-400">Issued</dt>
            <dd class="mt-1 text-2xl font-semibold tabular-nums text-brand-600 dark:text-brand-400">{{ snapshot.invoiceRunStatus.issuedInvoices }}</dd>
          </div>
          <div class="col-span-2 rounded-xl bg-gray-50 p-4 md:col-span-1 dark:bg-white/[0.04]">
            <dt class="text-xs font-medium uppercase text-gray-500 dark:text-gray-400">Last run</dt>
            <dd class="mt-1 text-sm font-medium text-gray-900 dark:text-white/90">{{ snapshot.invoiceRunStatus.lastRunLabel }}</dd>
          </div>
        </dl>
      </section>

      <div class="grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,1.15fr)_minmax(360px,0.85fr)]">
        <section aria-labelledby="incomplete-heading">
          <h2 id="incomplete-heading" class="sr-only">Incomplete attendance</h2>
          <app-table-shell
            title="Incomplete attendance"
            description="Sessions that need manager review before billing stays clean."
          >
            @if (snapshot.incompleteAttendance.length === 0) {
              <app-empty-state
                title="No incomplete sessions"
                message="Attendance is complete for this billing month."
              />
            } @else {
              <table class="min-w-full text-left text-sm">
                <thead>
                  <tr class="border-b border-gray-200 text-gray-500 dark:border-gray-800 dark:text-gray-400">
                    <th class="py-2 pr-3 font-medium">Child</th>
                    <th class="py-2 pr-3 font-medium">Date</th>
                    <th class="py-2 pr-3 font-medium">Issue</th>
                    <th class="py-2 font-medium">Action</th>
                  </tr>
                </thead>
                <tbody>
                  @for (item of snapshot.incompleteAttendance; track item.id) {
                    <tr class="border-b border-gray-100 last:border-b-0 hover:bg-gray-50 dark:border-gray-800/60 dark:hover:bg-white/[0.03]">
                      <td class="py-3 pr-3 font-medium text-gray-800 dark:text-white/90">{{ item.childName }}</td>
                      <td class="py-3 pr-3 text-gray-600 dark:text-gray-300">
                        <span class="inline-flex items-center gap-2">
                          {{ item.localDateLabel }}
                          @if (item.isToday) {
                            <app-status-badge status="due" label="Today" size="sm" />
                          }
                        </span>
                      </td>
                      <td class="py-3 pr-3 text-gray-600 dark:text-gray-300">{{ item.issue }}</td>
                      <td class="py-3">
                        @if (item.childId && item.localDate) {
                          <a
                            class="inline-flex min-h-9 items-center rounded-lg px-3 text-sm font-medium text-brand-600 transition hover:bg-brand-50 hover:text-brand-700 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-500 dark:text-brand-400 dark:hover:bg-brand-500/15"
                            [routerLink]="['/staff/manager/attendance-corrections']"
                            [queryParams]="{ child_id: item.childId, local_date: item.localDate, session_id: item.sessionId || undefined }"
                            [attr.aria-label]="'Correct attendance for ' + item.childName"
                          >
                            Correct
                          </a>
                        } @else {
                          <span class="text-gray-500 dark:text-gray-400">{{ item.actionHint }}</span>
                        }
                      </td>
                    </tr>
                  }
                </tbody>
              </table>
            }
          </app-table-shell>
        </section>

        <section aria-labelledby="actions-heading">
          <div class="rounded-2xl border border-gray-200 bg-white p-5 shadow-theme-xs dark:border-gray-800 dark:bg-white/[0.03]">
            <div class="mb-4">
              <p class="text-xs font-medium uppercase text-gray-500 dark:text-gray-400">Workflow</p>
              <h2 id="actions-heading" class="text-lg font-semibold text-gray-800 dark:text-white/90">Quick actions</h2>
            </div>
            <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-1">
              @for (action of snapshot.quickActions; track action.label) {
                @if (action.disabled) {
                  <div
                    class="rounded-xl border border-gray-200 bg-gray-50 p-4 opacity-70 dark:border-gray-800 dark:bg-white/[0.03]"
                    aria-disabled="true"
                  >
                    <p class="text-sm font-medium text-gray-700 dark:text-white/80">{{ action.label }}</p>
                    <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ action.description }}</p>
                  </div>
                } @else {
                  <a
                    [routerLink]="action.route!"
                    class="group block min-h-20 rounded-xl border border-gray-200 bg-white p-4 transition hover:border-brand-200 hover:bg-brand-50/60 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-500 dark:border-gray-800 dark:bg-white/[0.02] dark:hover:border-brand-500/30 dark:hover:bg-brand-500/10"
                  >
                    <span class="flex items-start justify-between gap-3">
                      <span>
                        <span class="block text-sm font-medium text-gray-800 group-hover:text-brand-700 dark:text-white/90 dark:group-hover:text-brand-300">{{ action.label }}</span>
                        <span class="mt-1 block text-xs text-gray-500 dark:text-gray-400">{{ action.description }}</span>
                      </span>
                      <span aria-hidden="true" class="text-gray-400 transition group-hover:translate-x-0.5 group-hover:text-brand-500">-&gt;</span>
                    </span>
                  </a>
                }
              }
            </div>
          </div>
        </section>
      </div>

      <section aria-labelledby="payment-heading">
        <h2 id="payment-heading" class="sr-only">Payment follow-up</h2>
        <app-table-shell
          title="Payment follow-up"
          description="Sorted by urgency so overdue and failed payments are reviewed first."
        >
          <div shell-actions class="text-sm font-medium text-gray-700 dark:text-gray-200">
            {{ formatGbp(outstandingTotalMinor) }} outstanding
          </div>
          @if (sortedPayments.length === 0) {
            <app-empty-state title="No outstanding payments" message="All issued invoices are currently settled." />
          } @else {
            <table class="min-w-full text-left text-sm">
              <thead>
                <tr class="border-b border-gray-200 text-gray-500 dark:border-gray-800 dark:text-gray-400">
                  <th class="py-2 pr-3 font-medium">Invoice</th>
                  <th class="py-2 pr-3 font-medium">Child</th>
                  <th class="py-2 pr-3 font-medium">Payer</th>
                  <th class="py-2 pr-3 font-medium">Due</th>
                  <th class="py-2 pr-3 font-medium">Status</th>
                  <th class="py-2 text-right font-medium">Outstanding</th>
                </tr>
              </thead>
              <tbody>
                @for (item of sortedPayments; track item.id) {
                  <tr class="border-b border-gray-100 last:border-b-0 hover:bg-gray-50 dark:border-gray-800/60 dark:hover:bg-white/[0.03]">
                    <td class="py-3 pr-3 font-medium text-gray-800 dark:text-white/90">{{ item.invoiceNumber }}</td>
                    <td class="py-3 pr-3 text-gray-800 dark:text-white/90">{{ item.childName }}</td>
                    <td class="py-3 pr-3 text-gray-600 dark:text-gray-300">{{ item.payerName }}</td>
                    <td class="py-3 pr-3 text-gray-600 dark:text-gray-300">{{ item.dueDateLabel }}</td>
                    <td class="py-3 pr-3"><app-status-badge [status]="item.status" size="sm" /></td>
                    <td class="py-3 text-right font-semibold tabular-nums text-gray-900 dark:text-white">{{ formatGbp(item.outstandingMinor) }}</td>
                  </tr>
                }
              </tbody>
            </table>
          }
        </app-table-shell>
      </section>
    </div>
  `,
})
export class ManagerDashboardComponent {
  readonly snapshot: ManagerDashboardSnapshot = MANAGER_DASHBOARD_MOCK;
  readonly attendanceRoute = '/staff/practitioner/attendance';
  readonly invoiceRunRoute = '/staff/manager/invoice-run';

  get attendanceTiles(): AttendanceTile[] {
    const s = this.snapshot.attendanceSummary;
    return [
      {
        label: 'Checked in today',
        value: s.checkedInToday,
        meta: 'On site',
        tone: 'success',
        description: 'Children currently present and visible to practitioners.',
      },
      {
        label: 'Not in yet',
        value: s.notInYet,
        meta: 'Pending',
        tone: 'neutral',
        description: 'Expected children without a check-in record yet.',
      },
      {
        label: 'Enrollment incomplete',
        value: s.enrollmentIncomplete,
        meta: 'Review',
        tone: 'warning',
        description: 'Profiles missing registration requirements.',
      },
      {
        label: 'Incomplete attendance',
        value: s.incompleteAttendance,
        meta: 'Blocks billing',
        tone: 'warning',
        description: 'Attendance records that need correction before invoicing.',
      },
    ];
  }

  get totalChildrenToday(): number {
    const s = this.snapshot.attendanceSummary;
    return s.checkedInToday + s.notInYet;
  }

  get attendanceCompletionRate(): number {
    if (this.totalChildrenToday === 0) {
      return 0;
    }
    return Math.round((this.snapshot.attendanceSummary.checkedInToday / this.totalChildrenToday) * 100);
  }

  get sortedPayments(): PaymentFollowUpInvoice[] {
    return sortPaymentFollowUp(this.snapshot.paymentFollowUp);
  }

  get outstandingTotalMinor(): number {
    return this.snapshot.paymentFollowUp.reduce((total, item) => total + item.outstandingMinor, 0);
  }

  formatGbp(minorUnits: number): string {
    return formatGbp(minorUnits);
  }
}
