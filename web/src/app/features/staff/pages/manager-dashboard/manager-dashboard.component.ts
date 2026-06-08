import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';

import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import {
  MANAGER_DASHBOARD_MOCK,
  ManagerDashboardSnapshot,
  PaymentFollowUpInvoice,
  formatGbp,
  sortPaymentFollowUp,
} from './manager-dashboard.models';

@Component({
  selector: 'app-manager-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    PageHeaderComponent,
    StatusBadgeComponent,
  ],
  template: `
    <div class="p-6 space-y-6">
      <app-page-header
        title="Manager operations"
        description="Today at the active branch."
        [hasActions]="true"
      >
        <a
          actions
          [routerLink]="attendanceRoute"
          class="inline-flex items-center justify-center gap-2 rounded-lg bg-brand-500 px-4 py-2.5 text-sm font-medium text-white shadow-theme-xs hover:bg-brand-600"
        >
          Open attendance
        </a>
      </app-page-header>

      <!-- Attendance summary -->
      <section aria-labelledby="attendance-heading">
        <h2 id="attendance-heading" class="text-sm font-medium text-gray-500 dark:text-gray-400 mb-3">Today's attendance</h2>
        <div class="grid grid-cols-2 gap-4 md:grid-cols-4">
          @for (tile of attendanceTiles; track tile.label) {
            <div class="rounded-xl border border-gray-200 bg-white p-4 dark:border-gray-800 dark:bg-white/[0.03]">
              <p class="text-2xl font-semibold text-gray-800 dark:text-white/90">{{ tile.value }}</p>
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ tile.label }}</p>
            </div>
          }
        </div>
      </section>

      <!-- Incomplete attendance triage -->
      <section aria-labelledby="incomplete-heading">
        <h2 id="incomplete-heading" class="text-sm font-medium text-gray-500 dark:text-gray-400 mb-3">Incomplete attendance</h2>
        @if (snapshot.incompleteAttendance.length === 0) {
          <p class="text-sm text-gray-500 dark:text-gray-400">No incomplete sessions this billing month.</p>
        } @else {
          <div class="overflow-x-auto rounded-xl border border-gray-200 bg-white dark:border-gray-800 dark:bg-white/[0.03]">
            <table class="w-full text-sm text-left">
              <thead class="bg-gray-50 dark:bg-gray-800/50">
                <tr>
                  <th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Child</th>
                  <th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Date</th>
                  <th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Issue</th>
                  <th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Action</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
                @for (item of snapshot.incompleteAttendance; track item.id) {
                  <tr>
                    <td class="px-4 py-3 text-gray-800 dark:text-white/90">{{ item.childName }}</td>
                    <td class="px-4 py-3 text-gray-500 dark:text-gray-400">{{ item.localDateLabel }}</td>
                    <td class="px-4 py-3 text-gray-500 dark:text-gray-400">{{ item.issue }}</td>
                    <td class="px-4 py-3 text-gray-500 dark:text-gray-400">
                      @if (item.childId && item.localDate) {
                        <a
                          class="text-blue-600 hover:text-blue-800 dark:text-blue-400"
                          [routerLink]="['/staff/manager/attendance-corrections']"
                          [queryParams]="{ child_id: item.childId, local_date: item.localDate, session_id: item.sessionId || undefined }"
                        >
                          Correct
                        </a>
                      } @else {
                        {{ item.actionHint }}
                      }
                    </td>
                  </tr>
                }
              </tbody>
            </table>
          </div>
        }
      </section>

      <!-- Invoice run status -->
      <section aria-labelledby="invoice-heading">
        <h2 id="invoice-heading" class="text-sm font-medium text-gray-500 dark:text-gray-400 mb-3">Invoice run status</h2>
        <div class="rounded-xl border border-gray-200 bg-white p-5 dark:border-gray-800 dark:bg-white/[0.03]">
          <dl class="grid grid-cols-2 gap-x-6 gap-y-3 sm:grid-cols-3 md:grid-cols-4">
            <div>
              <dt class="text-sm text-gray-500 dark:text-gray-400">Billing month</dt>
              <dd class="text-sm font-medium text-gray-800 dark:text-white/90">{{ snapshot.invoiceRunStatus.billingMonthLabel }}</dd>
            </div>
            <div>
              <dt class="text-sm text-gray-500 dark:text-gray-400">Eligible children</dt>
              <dd class="text-sm font-medium text-gray-800 dark:text-white/90">{{ snapshot.invoiceRunStatus.eligibleChildren }}</dd>
            </div>
            <div>
              <dt class="text-sm text-gray-500 dark:text-gray-400">Blocked children</dt>
              <dd class="text-sm font-medium text-gray-800 dark:text-white/90">{{ snapshot.invoiceRunStatus.blockedChildren }}</dd>
            </div>
            <div>
              <dt class="text-sm text-gray-500 dark:text-gray-400">Draft invoices</dt>
              <dd class="text-sm font-medium text-gray-800 dark:text-white/90">{{ snapshot.invoiceRunStatus.draftInvoices }}</dd>
            </div>
            <div>
              <dt class="text-sm text-gray-500 dark:text-gray-400">Issued invoices</dt>
              <dd class="text-sm font-medium text-gray-800 dark:text-white/90">{{ snapshot.invoiceRunStatus.issuedInvoices }}</dd>
            </div>
            <div>
              <dt class="text-sm text-gray-500 dark:text-gray-400">Last run</dt>
              <dd class="text-sm font-medium text-gray-800 dark:text-white/90">{{ snapshot.invoiceRunStatus.lastRunLabel }}</dd>
            </div>
          </dl>
          <p class="mt-4 text-sm text-gray-600 dark:text-gray-300">{{ snapshot.invoiceRunStatus.nextStep }}</p>
        </div>
      </section>

      <!-- Payment follow-up -->
      <section aria-labelledby="payment-heading">
        <h2 id="payment-heading" class="text-sm font-medium text-gray-500 dark:text-gray-400 mb-3">Payment follow-up</h2>
        @if (sortedPayments.length === 0) {
          <p class="text-sm text-gray-500 dark:text-gray-400">No outstanding payments.</p>
        } @else {
          <div class="overflow-x-auto rounded-xl border border-gray-200 bg-white dark:border-gray-800 dark:bg-white/[0.03]">
            <table class="w-full text-sm text-left">
              <thead class="bg-gray-50 dark:bg-gray-800/50">
                <tr>
                  <th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Invoice</th>
                  <th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Child</th>
                  <th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Payer</th>
                  <th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Due</th>
                  <th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Status</th>
                  <th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Outstanding</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
                @for (item of sortedPayments; track item.id) {
                  <tr>
                    <td class="px-4 py-3 text-gray-800 dark:text-white/90">{{ item.invoiceNumber }}</td>
                    <td class="px-4 py-3 text-gray-800 dark:text-white/90">{{ item.childName }}</td>
                    <td class="px-4 py-3 text-gray-500 dark:text-gray-400">{{ item.payerName }}</td>
                    <td class="px-4 py-3 text-gray-500 dark:text-gray-400">{{ item.dueDateLabel }}</td>
                    <td class="px-4 py-3"><app-status-badge [status]="item.status" size="sm"></app-status-badge></td>
                    <td class="px-4 py-3 text-gray-800 dark:text-white/90">{{ formatGbp(item.outstandingMinor) }}</td>
                  </tr>
                }
              </tbody>
            </table>
          </div>
        }
      </section>

      <!-- Quick actions -->
      <section aria-labelledby="actions-heading">
        <h2 id="actions-heading" class="text-sm font-medium text-gray-500 dark:text-gray-400 mb-3">Quick actions</h2>
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 md:grid-cols-3">
          @for (action of snapshot.quickActions; track action.label) {
            @if (action.disabled) {
              <div
                class="rounded-xl border border-gray-200 bg-white p-4 opacity-50 cursor-not-allowed dark:border-gray-800 dark:bg-white/[0.03]"
                [attr.aria-disabled]="true"
              >
                <p class="text-sm font-medium text-gray-800 dark:text-white/90">{{ action.label }}</p>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ action.description }}</p>
              </div>
            } @else {
              <a
                [routerLink]="action.route!"
                class="block rounded-xl border border-gray-200 bg-white p-4 hover:bg-gray-50 dark:border-gray-800 dark:bg-white/[0.03] dark:hover:bg-white/[0.06]"
              >
                <p class="text-sm font-medium text-gray-800 dark:text-white/90">{{ action.label }}</p>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ action.description }}</p>
              </a>
            }
          }
        </div>
      </section>
    </div>
  `,
})
export class ManagerDashboardComponent {
  readonly snapshot: ManagerDashboardSnapshot = MANAGER_DASHBOARD_MOCK;
  readonly attendanceRoute = '/staff/practitioner/attendance';

  get attendanceTiles(): { label: string; value: number }[] {
    const s = this.snapshot.attendanceSummary;
    return [
      { label: 'Checked in today', value: s.checkedInToday },
      { label: 'Not in yet', value: s.notInYet },
      { label: 'Enrollment incomplete', value: s.enrollmentIncomplete },
      { label: 'Incomplete attendance', value: s.incompleteAttendance },
    ];
  }

  get sortedPayments(): PaymentFollowUpInvoice[] {
    return sortPaymentFollowUp(this.snapshot.paymentFollowUp);
  }

  formatGbp(minorUnits: number): string {
    return formatGbp(minorUnits);
  }
}
