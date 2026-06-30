import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroMagnifyingGlass,
  heroPlus,
  heroTrash,
  heroDocumentText,
  heroSparkles,
  heroExclamationTriangle,
  heroCheck,
  heroXMark,
} from '@ng-icons/heroicons/outline';
import { catchError, of } from 'rxjs';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { ManagerInvoiceCreateApiService } from '../../data/manager-invoice-create-api.service';
import { StaffApiService } from '../../data/staff-api.service';
import { AuthService } from '../../../../core/services/auth.service';
import { ToastService } from '../../../../shared/services/toast.service';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { formatGbp } from '../../../owner/utils/owner-formatters';
import { FormInvoiceLine } from '../../models/manager-invoice-create.models';
import { formatChildName } from '../../utils/manager-list-formatters';

@Component({
  selector: 'app-manager-invoice-create',
  imports: [
    CommonModule,
    FormsModule,
    NgIcon,
    AlertComponent,
    LoadingStateComponent,
  ],
  templateUrl: './manager-invoice-create.component.html',
  providers: [
    provideIcons({
      heroMagnifyingGlass,
      heroPlus,
      heroTrash,
      heroDocumentText,
      heroSparkles,
      heroExclamationTriangle,
      heroCheck,
      heroXMark,
    }),
  ],
})
export class ManagerInvoiceCreateComponent implements OnInit {
  private readonly api = inject(ManagerInvoiceCreateApiService);
  private readonly staffApi = inject(StaffApiService);
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly toast = inject(ToastService);

  readonly formatGbp = formatGbp;
  readonly formatChildName = formatChildName;
  readonly Math = Math;
  readonly Number = Number;

  readonly DEFAULT_PAYMENT_TERMS = 'Payments are due within 7 days. Late fees may apply as per the parent agreement.';

  mode: 'new' | 'edit' = 'new';
  editInvoiceId: string | null = null;

  childSearchTerm = '';
  searchResults: { childId: string; fullName: string }[] = [];
  isSearching = false;
  selectedChild: { childId: string; fullName: string } | null = null;

  billingMonth = '';
  isLoadingPrefill = false;
  prefillError: string | null = null;

  lines = signal<FormInvoiceLine[]>([]);
  entitlementLabel = '';
  hasFundingProfile = false;

  internalNotes = '';
  paymentTerms = this.DEFAULT_PAYMENT_TERMS;

  isSaving = false;
  isIssuing = false;
  submitError: string | null = null;

  readonly subtotalMinor = computed(() =>
    this.lines().reduce((sum, l) => sum + l.lineAmountMinor, 0)
  );

  readonly fundedDeductionMinor = computed(() =>
    this.lines()
      .filter((l) => l.isFundingOffset)
      .reduce((sum, l) => sum + l.lineAmountMinor, 0)
  );

  readonly totalDueMinor = computed(() =>
    Math.max(0, this.subtotalMinor() - this.fundedDeductionMinor())
  );

  readonly showWarningBanner = computed(() => {
    const s = this.subtotalMinor();
    const f = this.fundedDeductionMinor();
    return s > 0 && f > 0 && f > s / 4;
  });

  ngOnInit(): void {
    const invoiceId = this.route.snapshot.paramMap.get('invoiceId');
    if (invoiceId) {
      this.mode = 'edit';
      this.editInvoiceId = invoiceId;
    }
    this.setDefaultBillingMonth();
  }

  private setDefaultBillingMonth(): void {
    const now = new Date();
    const y = now.getFullYear();
    const m = String(now.getMonth()).padStart(2, '0');
    this.billingMonth = `${y}-${m}`;
  }

  onSearchInput(): void {
    const term = this.childSearchTerm.trim();
    if (term.length < 2) {
      this.searchResults = [];
      return;
    }

    this.isSearching = true;
    const membership = this.auth.activeMembership();
    const branchId = membership?.branch_id ?? '';

    this.staffApi
      .listChildren({ status: 'active', limit: 20, offset: 0 })
      .pipe(
        catchError(() => of({ items: [], total: 0 })),
      )
      .subscribe({
        next: (result) => {
          this.searchResults = result.items
            .filter((c) => {
              const name = c.fullName.toLowerCase();
              const q = term.toLowerCase();
              return name.includes(q);
            })
            .map((c) => ({
              childId: c.id,
              fullName: c.fullName,
            }));
          this.isSearching = false;
        },
        error: () => {
          this.isSearching = false;
        },
      });
  }

  selectChild(child: { childId: string; fullName: string }): void {
    this.selectedChild = child;
    this.childSearchTerm = child.fullName;
    this.searchResults = [];
    this.loadPrefill();
  }

  clearChild(): void {
    this.selectedChild = null;
    this.childSearchTerm = '';
    this.lines.set([]);
    this.entitlementLabel = '';
    this.hasFundingProfile = false;
  }

  loadPrefill(): void {
    if (!this.selectedChild || !this.billingMonth) return;

    this.isLoadingPrefill = true;
    this.prefillError = null;

    this.api.getPrefill(this.selectedChild.childId, this.billingMonth).subscribe({
      next: (prefill) => {
        this.lines.set(
          prefill.lines.map((l, i) => ({
            id: `line-${i}`,
            lineKind: l.lineKind,
            description: l.description,
            sortOrder: l.sortOrder,
            quantityMinutes: l.quantityMinutes,
            unitAmountMinor: l.unitAmountMinor,
            lineAmountMinor: l.lineAmountMinor,
            fundedAllowanceMinutes: l.fundedAllowanceMinutes,
            fundedDeductionMinutes: l.fundedDeductionMinutes,
            coreBillableMinutes: l.coreBillableMinutes,
            sessionCount: l.sessionCount,
            isFundingOffset: l.lineKind === 'funded_deduction',
          })),
        );
        this.entitlementLabel = prefill.entitlementStatus.statusLabel;
        this.hasFundingProfile = prefill.entitlementStatus.fundingProfileId !== null;
        this.isLoadingPrefill = false;
      },
      error: (err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.prefillError = formatPresentedApiError(presentApiError(mapped, 'invoice.prefill'));
        this.isLoadingPrefill = false;
      },
    });
  }

  addBlankLine(): void {
    this.lines.update((prev) => [
      ...prev,
      {
        id: `line-${Date.now()}`,
        lineKind: 'extra',
        description: '',
        sortOrder: prev.length + 1,
        quantityMinutes: 0,
        unitAmountMinor: 0,
        lineAmountMinor: 0,
        fundedAllowanceMinutes: 0,
        fundedDeductionMinutes: 0,
        coreBillableMinutes: 0,
        sessionCount: 0,
        isFundingOffset: false,
      },
    ]);
  }

  removeLine(lineId: string): void {
    this.lines.update((prev) => prev.filter((l) => l.id !== lineId));
  }

  updateLine(lineId: string, field: keyof FormInvoiceLine, value: number | string): void {
    this.lines.update((prev) =>
      prev.map((l) => {
        if (l.id !== lineId) return l;
        const updated = { ...l, [field]: value };
        if (field === 'quantityMinutes' || field === 'unitAmountMinor') {
          const q = typeof updated.quantityMinutes === 'number' ? updated.quantityMinutes : 0;
          const u = typeof updated.unitAmountMinor === 'number' ? updated.unitAmountMinor : 0;
          updated.lineAmountMinor = q * u;
        }
        return updated;
      }),
    );
  }

  saveDraft(): void {
    if (!this.canSubmit()) return;
    this.isSaving = true;
    this.submitError = null;

    this.api
      .createDraft({
        childId: this.selectedChild!.childId,
        billingMonth: this.billingMonth,
        lines: this.lines().map((l) => ({
          lineKind: l.lineKind,
          description: l.description,
          sortOrder: l.sortOrder,
          quantityMinutes: l.quantityMinutes,
          unitAmountMinor: l.unitAmountMinor,
          lineAmountMinor: l.lineAmountMinor,
        })),
        paymentTerms: this.paymentTerms,
        internalNotes: this.internalNotes,
      })
      .subscribe({
        next: () => {
          this.isSaving = false;
          this.toast.success('Draft invoice created.');
          this.router.navigate(['/manager/invoices']);
        },
        error: (err) => {
          this.isSaving = false;
          const mapped = this.errorMapper.mapAndHandle(err);
          this.submitError = formatPresentedApiError(presentApiError(mapped, 'invoice.createDraft'));
        },
      });
  }

  issueInvoice(): void {
    if (!this.canSubmit()) return;
    this.isIssuing = true;
    this.submitError = null;

    this.api
      .createAndIssue({
        childId: this.selectedChild!.childId,
        billingMonth: this.billingMonth,
        lines: this.lines().map((l) => ({
          lineKind: l.lineKind,
          description: l.description,
          sortOrder: l.sortOrder,
          quantityMinutes: l.quantityMinutes,
          unitAmountMinor: l.unitAmountMinor,
          lineAmountMinor: l.lineAmountMinor,
        })),
        paymentTerms: this.paymentTerms,
        internalNotes: this.internalNotes,
      })
      .subscribe({
        next: (result) => {
          this.isIssuing = false;
          this.toast.success(`Invoice ${result.invoiceNumber} issued.`);
          this.router.navigate(['/manager/invoices', result.invoiceId]);
        },
        error: (err) => {
          this.isIssuing = false;
          const mapped = this.errorMapper.mapAndHandle(err);
          this.submitError = formatPresentedApiError(presentApiError(mapped, 'invoice.issue'));
        },
      });
  }

  canSubmit(): boolean {
    if (!this.selectedChild || !this.billingMonth) return false;
    return this.lines().length > 0;
  }
}
