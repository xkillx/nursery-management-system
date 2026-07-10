import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroUserGroup,
  heroClipboardDocumentCheck,
  heroClipboardDocumentList,
  heroDocumentText,
  heroCheckCircle,
  heroExclamationTriangle,
  heroClock,
  heroArrowRight,
  heroExclamationCircle,
} from '@ng-icons/heroicons/outline';

import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { TableShellComponent } from '../../../../shared/components/ui/table/table-shell.component';
import { ChildAvatarComponent } from '../../../../shared/components/ui/avatar/child-avatar/child-avatar.component';
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
    NgIcon,
    ChildAvatarComponent,
  ],
  providers: [
    provideIcons({
      heroUserGroup,
      heroClipboardDocumentCheck,
      heroClipboardDocumentList,
      heroDocumentText,
      heroCheckCircle,
      heroExclamationTriangle,
      heroClock,
      heroArrowRight,
      heroExclamationCircle,
    }),
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
          class="inline-flex min-h-11 w-full items-center justify-center gap-2 rounded-lg bg-brand-500 px-4 py-2.5 text-sm font-medium text-white shadow-theme-xs transition hover:bg-brand-600 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-500 sm:w-auto"
        >
          <span>Open attendance</span>
          <ng-icon name="heroArrowRight" size="16" class="shrink-0"></ng-icon>
        </a>
      </app-page-header>

      <section aria-labelledby="attendance-heading" class="space-y-3">
        <div class="flex flex-col gap-1 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p class="text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">Today</p>
            <h2 id="attendance-heading" class="text-lg font-bold text-gray-800 dark:text-white/90">Attendance snapshot</h2>
          </div>
          <p class="text-sm text-gray-500 dark:text-gray-400 font-medium">
            {{ attendanceCompletionRate }}% attendance completion across {{ totalChildrenToday }} expected children.
          </p>
        </div>

        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
          @for (tile of attendanceTiles; track tile.label) {
            <article 
              [ngClass]="getCardHoverClass(tile.tone, tile.label)"
              class="rounded-2xl border border-gray-200 bg-white p-5 shadow-theme-xs transition-all duration-200 hover:-translate-y-0.5 hover:shadow-theme-sm dark:border-gray-800 dark:bg-white/[0.03]">
              <div class="flex items-start justify-between gap-3">
                <div class="flex items-center gap-3">
                  <div [ngClass]="getIconBgClass(tile.tone, tile.label)" class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl">
                    <ng-icon [name]="getIconName(tile.tone, tile.label)" size="20" [ngClass]="getIconColorClass(tile.tone, tile.label)"></ng-icon>
                  </div>
                  <div>
                    <h3 class="text-sm font-semibold text-gray-800 dark:text-white/90 leading-tight">{{ tile.label }}</h3>
                    <p class="mt-1 text-xs text-gray-500 dark:text-gray-400 font-medium line-clamp-1">{{ tile.meta }}</p>
                  </div>
                </div>
              </div>
              <div class="mt-5 flex items-baseline gap-2">
                <span class="text-3xl font-bold tracking-tight text-gray-900 dark:text-white tabular-nums">{{ tile.value }}</span>
                <span class="text-xs text-gray-400 dark:text-gray-500 font-medium">children</span>
              </div>
              <p class="mt-2.5 text-xs leading-relaxed text-gray-500 dark:text-gray-400 line-clamp-2 h-9">{{ tile.description }}</p>
            </article>
          }
        </div>
      </section>

      <section aria-labelledby="invoice-heading" class="rounded-2xl border border-gray-200 bg-white p-5 shadow-theme-xs dark:border-gray-800 dark:bg-white/[0.03]">
        <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between border-b border-gray-100 pb-5 dark:border-gray-800/60">
          <div>
            <p class="text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">Billing month</p>
            <h2 id="invoice-heading" class="text-lg font-bold text-gray-800 dark:text-white/90">
              {{ snapshot.invoiceRunStatus.billingMonthLabel }} invoice run
            </h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400 font-medium">
              {{ snapshot.invoiceRunStatus.nextStep }}
            </p>
          </div>
          <div class="flex items-center gap-2 rounded-xl bg-gray-50 px-3 py-2 text-xs font-medium text-gray-600 dark:bg-white/[0.04] dark:text-gray-300 self-start">
            <ng-icon name="heroClock" size="14" class="text-gray-400 shrink-0"></ng-icon>
            <span>Last run: <strong class="font-semibold text-gray-800 dark:text-white">{{ snapshot.invoiceRunStatus.lastRunLabel }}</strong></span>
          </div>
        </div>

        <div class="mt-6">
          <div class="flex items-center justify-between text-xs font-semibold mb-2">
            <span class="text-gray-600 dark:text-gray-300">Run progress</span>
            <span class="text-brand-600 dark:text-brand-400">{{ invoiceProgressPercentage }}% complete</span>
          </div>
          <div class="flex h-3 w-full overflow-hidden rounded-full bg-gray-100 dark:bg-gray-800">
            <div 
              class="bg-brand-500 transition-all duration-500" 
              [style.width.%]="issuedPercentage" 
              title="Issued: {{ snapshot.invoiceRunStatus.issuedInvoices }}">
            </div>
            <div 
              class="bg-blue-light-400 transition-all duration-500" 
              [style.width.%]="draftsPercentage" 
              title="Drafts: {{ snapshot.invoiceRunStatus.draftInvoices }}">
            </div>
            <div 
              class="bg-warning-500 transition-all duration-500" 
              [style.width.%]="blockedPercentage" 
              title="Blocked: {{ snapshot.invoiceRunStatus.blockedChildren }}">
            </div>
          </div>
          <div class="mt-2.5 flex flex-wrap gap-x-4 gap-y-1.5 text-xs text-gray-500 dark:text-gray-400">
            <span class="flex items-center gap-1.5">
              <span class="h-2.5 w-2.5 rounded-full bg-brand-500"></span>
              <span>Issued ({{ snapshot.invoiceRunStatus.issuedInvoices }})</span>
            </span>
            <span class="flex items-center gap-1.5">
              <span class="h-2.5 w-2.5 rounded-full bg-blue-light-400"></span>
              <span>Drafts ({{ snapshot.invoiceRunStatus.draftInvoices }})</span>
            </span>
            <span class="flex items-center gap-1.5">
              <span class="h-2.5 w-2.5 rounded-full bg-warning-500"></span>
              <span>Blocked ({{ snapshot.invoiceRunStatus.blockedChildren }})</span>
            </span>
          </div>
        </div>

        <div class="mt-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
          <div class="rounded-xl border border-gray-100 bg-gray-50/50 p-4 dark:border-gray-800/40 dark:bg-white/[0.01]">
            <dt class="text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">Eligible</dt>
            <dd class="mt-2 flex items-baseline gap-1">
              <span class="text-2xl font-bold tabular-nums text-gray-900 dark:text-white">{{ snapshot.invoiceRunStatus.eligibleChildren }}</span>
              <span class="text-xs text-gray-400 font-medium">children</span>
            </dd>
          </div>
          <div class="rounded-xl border border-warning-100 bg-warning-50/20 p-4 dark:border-warning-500/10 dark:bg-warning-500/5">
            <dt class="text-xs font-semibold uppercase tracking-wider text-warning-700 dark:text-warning-400">Blocked</dt>
            <dd class="mt-2 flex items-baseline gap-1">
              <span class="text-2xl font-bold tabular-nums text-warning-600 dark:text-warning-500">{{ snapshot.invoiceRunStatus.blockedChildren }}</span>
              <span class="text-xs text-warning-400 font-medium">need review</span>
            </dd>
          </div>
          <div class="rounded-xl border border-gray-100 bg-gray-50/50 p-4 dark:border-gray-800/40 dark:bg-white/[0.01]">
            <dt class="text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">Drafts</dt>
            <dd class="mt-2 flex items-baseline gap-1">
              <span class="text-2xl font-bold tabular-nums text-gray-900 dark:text-white">{{ snapshot.invoiceRunStatus.draftInvoices }}</span>
              <span class="text-xs text-gray-400 font-medium">in drafts</span>
            </dd>
          </div>
          <div class="rounded-xl border border-brand-100 bg-brand-50/20 p-4 dark:border-brand-500/10 dark:bg-brand-500/5">
            <dt class="text-xs font-semibold uppercase tracking-wider text-brand-700 dark:text-brand-400">Issued</dt>
            <dd class="mt-2 flex items-baseline gap-1">
              <span class="text-2xl font-bold tabular-nums text-brand-600 dark:text-brand-400">{{ snapshot.invoiceRunStatus.issuedInvoices }}</span>
              <span class="text-xs text-brand-400 font-medium">sent</span>
            </dd>
          </div>
        </div>
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
                    <th class="py-2.5 pr-3 font-semibold">Child</th>
                    <th class="py-2.5 pr-3 font-semibold">Date</th>
                    <th class="py-2.5 pr-3 font-semibold">Issue</th>
                    <th class="py-2.5 font-semibold">Action</th>
                  </tr>
                </thead>
                <tbody>
                  @for (item of snapshot.incompleteAttendance; track item.id) {
                    <tr class="border-b border-gray-100 last:border-b-0 hover:bg-gray-50/60 dark:border-gray-800/60 dark:hover:bg-white/[0.02]">
                      <td class="py-3.5 pr-3 font-medium text-gray-800 dark:text-white/90">
                        <div class="flex items-center gap-3">
                          <span class="shrink-0"><app-child-avatar [photoUrl]="item.photoUrl ?? null" [name]="item.childName" size="sm"></app-child-avatar></span>
                          <span>{{ item.childName }}</span>
                        </div>
                      </td>
                      <td class="py-3.5 pr-3 text-gray-600 dark:text-gray-300">
                        <span class="inline-flex items-center gap-2">
                          @if (item.isToday) {
                            <app-status-badge status="due" label="Today" size="sm" />
                          } @else {
                            {{ item.localDateLabel }}
                          }
                        </span>
                      </td>
                      <td class="py-3.5 pr-3 text-gray-600 dark:text-gray-300">{{ item.issue }}</td>
                      <td class="py-3.5">
                        @if (item.childId && item.localDate) {
                          <a
                            class="inline-flex min-h-8 items-center gap-1.5 rounded-lg border border-brand-100 bg-brand-25/50 px-3 text-xs font-semibold text-brand-600 transition hover:bg-brand-50 hover:text-brand-700 dark:border-brand-500/20 dark:bg-brand-500/10 dark:text-brand-400 dark:hover:bg-brand-500/20"
                            [routerLink]="['/manager/attendance-corrections']"
                            [queryParams]="{ child_id: item.childId, local_date: item.localDate, session_id: item.sessionId || undefined }"
                            [attr.aria-label]="'Correct attendance for ' + item.childName"
                          >
                            <span>Correct</span>
                            <ng-icon name="heroArrowRight" size="12"></ng-icon>
                          </a>
                        } @else {
                          <span class="inline-flex items-center rounded-lg bg-gray-50 border border-gray-100 px-2 py-1 text-xs font-semibold text-gray-500 dark:bg-white/5 dark:border-gray-800 dark:text-gray-400">
                            {{ item.actionHint }}
                          </span>
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
          <div class="rounded-2xl border border-gray-200 bg-white p-5 shadow-theme-xs dark:border-gray-800 dark:bg-white/[0.03] h-full flex flex-col justify-between">
            <div class="mb-4">
              <p class="text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">Workflow</p>
              <h2 id="actions-heading" class="text-lg font-bold text-gray-800 dark:text-white/90">Quick actions</h2>
            </div>
            <div class="grid grid-cols-1 gap-3 flex-1">
              @for (action of snapshot.quickActions; track action.label) {
                @if (action.disabled) {
                  <div
                    class="flex items-start gap-3 rounded-xl border border-gray-100 bg-gray-50/50 p-4 opacity-50 dark:border-gray-800/60 dark:bg-white/[0.01]"
                    aria-disabled="true"
                  >
                    <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-gray-100 text-gray-400 dark:bg-white/5">
                      <ng-icon [name]="getActionIcon(action.label)" size="18"></ng-icon>
                    </div>
                    <div>
                      <p class="text-sm font-semibold text-gray-500 dark:text-gray-400 flex items-center gap-1.5">
                        {{ action.label }}
                        <span class="inline-flex items-center rounded-full bg-gray-100 px-1.5 py-0.5 text-[10px] font-semibold text-gray-500 dark:bg-white/5">Coming soon</span>
                      </p>
                      <p class="mt-0.5 text-xs text-gray-400 dark:text-gray-500">{{ action.description }}</p>
                    </div>
                  </div>
                } @else {
                  <a
                    [routerLink]="action.route!"
                    class="group block min-h-20 rounded-xl border border-gray-200 bg-white p-4 transition hover:border-brand-200 hover:bg-brand-50/60 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-500 dark:border-gray-800 dark:bg-white/[0.02] dark:hover:border-brand-500/30 dark:hover:bg-brand-500/10"
                  >
                    <span class="flex items-start justify-between gap-3">
                      <span class="flex items-start gap-3">
                        <span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-brand-50 text-brand-500 transition-colors duration-200 group-hover:bg-brand-100 group-hover:text-brand-600 dark:bg-brand-500/10 dark:text-brand-400 dark:group-hover:bg-brand-500/20">
                          <ng-icon [name]="getActionIcon(action.label)" size="18"></ng-icon>
                        </span>
                        <span>
                          <span class="block text-sm font-semibold text-gray-800 group-hover:text-brand-700 dark:text-white/90 dark:group-hover:text-brand-300">{{ action.label }}</span>
                          <span class="mt-0.5 block text-xs text-gray-500 dark:text-gray-400">{{ action.description }}</span>
                        </span>
                      </span>
                      <span aria-hidden="true" class="text-gray-400 transition-all duration-200 group-hover:translate-x-0.5 group-hover:text-brand-500">
                        <ng-icon name="heroArrowRight" size="14"></ng-icon>
                      </span>
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
          <div shell-actions class="inline-flex items-center gap-1.5 rounded-full bg-error-50 px-2.5 py-1 text-xs font-semibold text-error-700 dark:bg-error-500/15 dark:text-error-400 border border-error-100 dark:border-error-500/20">
            <span class="h-1.5 w-1.5 rounded-full bg-error-500"></span>
            <span>{{ formatGbp(outstandingTotalMinor) }} outstanding</span>
          </div>
          @if (sortedPayments.length === 0) {
            <app-empty-state title="No outstanding payments" message="All issued invoices are currently settled." />
          } @else {
            <table class="min-w-full text-left text-sm">
              <thead>
                <tr class="border-b border-gray-200 text-gray-500 dark:border-gray-800 dark:text-gray-400">
                  <th class="py-2.5 pr-3 font-semibold">Invoice</th>
                  <th class="py-2.5 pr-3 font-semibold">Child</th>
                  <th class="py-2.5 pr-3 font-semibold">Payer</th>
                  <th class="py-2.5 pr-3 font-semibold">Due</th>
                  <th class="py-2.5 pr-3 font-semibold">Status</th>
                  <th class="py-2.5 text-right font-semibold">Outstanding</th>
                </tr>
              </thead>
              <tbody>
                @for (item of sortedPayments; track item.id) {
                  <tr class="border-b border-gray-100 last:border-b-0 hover:bg-gray-50/60 dark:border-gray-800/60 dark:hover:bg-white/[0.02]">
                    <td class="py-3.5 pr-3 font-semibold text-gray-800 dark:text-white/90">{{ item.invoiceNumber }}</td>
                    <td class="py-3.5 pr-3 text-gray-800 dark:text-white/90 font-medium">
                      <div class="flex items-center gap-3">
                        <span class="shrink-0"><app-child-avatar [photoUrl]="item.photoUrl ?? null" [name]="item.childName" size="sm"></app-child-avatar></span>
                        <span>{{ item.childName }}</span>
                      </div>
                    </td>
                    <td class="py-3.5 pr-3 text-gray-600 dark:text-gray-300 font-medium">
                      <div class="flex items-center gap-1.5">
                        <ng-icon name="heroUserGroup" size="14" class="text-gray-400 shrink-0"></ng-icon>
                        <span>{{ item.payerName }}</span>
                      </div>
                    </td>
                    <td class="py-3.5 pr-3 text-gray-600 dark:text-gray-300">{{ item.dueDateLabel }}</td>
                    <td class="py-3.5 pr-3"><app-status-badge [status]="item.status" size="sm" /></td>
                    <td class="py-3.5 text-right font-bold tabular-nums text-gray-900 dark:text-white">{{ formatGbp(item.outstandingMinor) }}</td>
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
  readonly attendanceRoute = '/manager/attendance';

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

  getIconName(tone: AttendanceTileTone, label: string): string {
    if (label.includes('Checked in')) return 'heroClipboardDocumentCheck';
    if (label.includes('Not in yet')) return 'heroClock';
    if (label.includes('Enrollment')) return 'heroExclamationCircle';
    return 'heroExclamationTriangle';
  }

  getIconColorClass(tone: AttendanceTileTone, label: string): string {
    if (label.includes('Checked in')) return 'text-success-600 dark:text-success-400';
    if (label.includes('Not in yet')) return 'text-gray-500 dark:text-gray-400';
    if (label.includes('Enrollment')) return 'text-warning-600 dark:text-warning-400';
    return 'text-orange-600 dark:text-orange-400';
  }

  getIconBgClass(tone: AttendanceTileTone, label: string): string {
    if (label.includes('Checked in')) return 'bg-success-50/85 border border-success-100/50 dark:bg-success-500/10 dark:border-success-500/20';
    if (label.includes('Not in yet')) return 'bg-gray-50/85 border border-gray-100/50 dark:bg-white/5 dark:border-white/10';
    if (label.includes('Enrollment')) return 'bg-warning-50/85 border border-warning-100/50 dark:bg-warning-500/10 dark:border-warning-500/20';
    return 'bg-orange-50/85 border border-orange-100/50 dark:bg-orange-500/10 dark:border-orange-500/20';
  }

  getCardHoverClass(tone: AttendanceTileTone, label: string): string {
    if (label.includes('Checked in')) return 'hover:border-success-200 dark:hover:border-success-500/30';
    if (label.includes('Not in yet')) return 'hover:border-gray-300 dark:hover:border-white/20';
    if (label.includes('Enrollment')) return 'hover:border-warning-200 dark:hover:border-warning-500/30';
    return 'hover:border-orange-200 dark:hover:border-orange-500/30';
  }

  getActionIcon(label: string): string {
    switch (label) {
      case 'Open attendance':
        return 'heroClipboardDocumentCheck';
      case 'Attendance corrections':
        return 'heroClipboardDocumentList';
      case 'Manage children':
        return 'heroUserGroup';
      case 'Review payment follow-up':
        return 'heroDocumentText';
      default:
        return 'heroCheckCircle';
    }
  }

  get issuedPercentage(): number {
    const total = this.snapshot.invoiceRunStatus.eligibleChildren;
    if (total === 0) return 0;
    return (this.snapshot.invoiceRunStatus.issuedInvoices / total) * 100;
  }

  get draftsPercentage(): number {
    const total = this.snapshot.invoiceRunStatus.eligibleChildren;
    if (total === 0) return 0;
    return (this.snapshot.invoiceRunStatus.draftInvoices / total) * 100;
  }

  get blockedPercentage(): number {
    const total = this.snapshot.invoiceRunStatus.eligibleChildren;
    if (total === 0) return 0;
    return (this.snapshot.invoiceRunStatus.blockedChildren / total) * 100;
  }

  get invoiceProgressPercentage(): number {
    const run = this.snapshot.invoiceRunStatus;
    if (run.eligibleChildren === 0) return 0;
    return Math.round(((run.issuedInvoices + run.draftInvoices) / run.eligibleChildren) * 100);
  }
}
