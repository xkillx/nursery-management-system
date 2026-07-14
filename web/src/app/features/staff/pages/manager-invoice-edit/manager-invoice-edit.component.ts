import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroPlus,
  heroTrash,
  heroArrowPath,
  heroCheck,
  heroPaperAirplane,
  heroExclamationTriangle,
  heroChevronRight,
} from '@ng-icons/heroicons/outline';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { ManagerInvoicesApiService } from '../../data/manager-invoices-api.service';
import { ManagerInvoiceCreateApiService } from '../../data/manager-invoice-create-api.service';
import { ToastService } from '../../../../shared/services/toast.service';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { formatGbp, formatMinutes, formatBillingMonthLabel } from '../../utils/invoice-run-formatters';
import { FormInvoiceLine } from '../../models/manager-invoice-create.models';
import {
  ManagerInvoiceDetail,
  ManagerInvoiceLine,
  AddInvoiceLineInput,
  UpdateInvoiceLineInput,
} from '../../models/manager-invoices.models';

const SYSTEM_LINE_KINDS = new Set(['core_childcare', 'funded_deduction', 'hourly']);

function isEditableLine(line: FormInvoiceLine): boolean {
  return line.lineKind === 'extra' || line.lineKind === 'ad_hoc';
}

function mapApiLineToForm(line: ManagerInvoiceLine): FormInvoiceLine {
  return {
    id: line.lineId,
    lineKind: line.lineKind,
    description: line.description,
    sortOrder: line.sortOrder,
    quantityMinutes: line.quantityMinutes ?? 0,
    unitAmountMinor: line.unitAmountMinor ?? 0,
    lineAmountMinor: line.lineAmountMinor,
    fundedAllowanceMinutes: line.fundedAllowanceMinutes ?? 0,
    fundedDeductionMinutes: line.fundedDeductionMinutes ?? 0,
    coreBillableMinutes: line.coreBillableMinutes ?? 0,
    sessionCount: line.sessionCount ?? 0,
    isFundingOffset: line.lineKind === 'funded_deduction',
  };
}

@Component({
  selector: 'app-manager-invoice-edit',
  imports: [
    CommonModule,
    FormsModule,
    RouterModule,
    NgIcon,
    AlertComponent,
    LoadingStateComponent,
    PageHeaderComponent,
  ],
  templateUrl: './manager-invoice-edit.component.html',
  providers: [
    provideIcons({
      heroPlus,
      heroTrash,
      heroArrowPath,
      heroCheck,
      heroPaperAirplane,
      heroExclamationTriangle,
      heroChevronRight,
    }),
  ],
})
export class ManagerInvoiceEditComponent implements OnInit {
  private readonly api = inject(ManagerInvoicesApiService);
  private readonly createApi = inject(ManagerInvoiceCreateApiService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly toast = inject(ToastService);

  readonly formatGbp = formatGbp;
  readonly formatMinutes = formatMinutes;
  readonly formatBillingMonthLabel = formatBillingMonthLabel;
  readonly isEditableLine = isEditableLine;
  readonly Math = Math;

  invoice: ManagerInvoiceDetail | null = null;
  isLoading = false;
  errorMessage: string | null = null;

  lines = signal<FormInvoiceLine[]>([]);
  originalLines = signal<FormInvoiceLine[]>([]);

  isSaving = false;
  isIssuing = false;
  isRegenerating = false;
  submitError: string | null = null;

  readonly subtotalMinor = computed(() =>
    this.lines().reduce((sum, l) => sum + l.lineAmountMinor, 0),
  );

  readonly fundedDeductionMinor = computed(() =>
    this.lines()
      .filter((l) => l.lineKind === 'funded_deduction')
      .reduce((sum, l) => sum + l.lineAmountMinor, 0),
  );

  readonly totalDueMinor = computed(() =>
    Math.max(0, this.subtotalMinor() - this.fundedDeductionMinor()),
  );

  get hasUnsavedChanges(): boolean {
    return JSON.stringify(this.lines()) !== JSON.stringify(this.originalLines());
  }

  get childName(): string {
    return this.invoice?.childName ?? '';
  }

  get billingMonth(): string {
    return this.invoice?.billingMonth ?? '';
  }

  get parentContactName(): string {
    return this.invoice?.parentContact?.fullName ?? '';
  }

  ngOnInit(): void {
    const invoiceId = this.route.snapshot.paramMap.get('invoiceId');
    if (!invoiceId) {
      this.errorMessage = 'Invoice ID is missing.';
      return;
    }
    this.loadInvoice(invoiceId);
  }

  private loadInvoice(invoiceId: string): void {
    this.isLoading = true;
    this.errorMessage = null;

    this.api.getInvoice(invoiceId).subscribe({
      next: (detail) => {
        if (detail.status !== 'draft') {
          this.router.navigate(['/manager/invoices', invoiceId]);
          this.toast.warning('Only draft invoices can be edited.');
          return;
        }

        this.invoice = detail;
        const formLines = detail.lines.map(mapApiLineToForm);
        this.lines.set(formLines);
        this.originalLines.set(formLines.map((l) => ({ ...l })));
        this.isLoading = false;
      },
      error: (err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'invoice.managerDetail'));
        this.isLoading = false;
      },
    });
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

  regenerate(): void {
    if (!this.invoice || this.isRegenerating) return;
    this.isRegenerating = true;

    this.createApi.getPrefill(this.invoice.childId, this.invoice.billingMonth).subscribe({
      next: (prefill) => {
        const manualLines = this.lines().filter(
          (l) => l.lineKind === 'extra' || l.lineKind === 'ad_hoc',
        );
        const systemLines = prefill.lines
          .filter((l) => SYSTEM_LINE_KINDS.has(l.lineKind))
          .map(
            (l, i): FormInvoiceLine => ({
              id: `regen-${Date.now()}-${i}`,
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
            }),
          );
        this.lines.set([...systemLines, ...manualLines]);
        this.isRegenerating = false;
        this.toast.success('System lines regenerated.');
      },
      error: (err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.submitError = formatPresentedApiError(presentApiError(mapped, 'invoice.prefill'));
        this.isRegenerating = false;
      },
    });
  }

  saveChanges(): void {
    if (!this.invoice || this.isSaving) return;
    this.isSaving = true;
    this.submitError = null;

    const invoiceId = this.invoice.invoiceId;
    const current = this.lines();
    const original = this.originalLines();

    const originalMap = new Map(original.map((l) => [l.id, l]));
    const currentMap = new Map(current.map((l) => [l.id, l]));

    const toDelete = original.filter((l) => !currentMap.has(l.id) && !l.id.startsWith('line-'));
    const toUpdate = current.filter((l) => {
      const orig = originalMap.get(l.id);
      return (
        orig &&
        !l.id.startsWith('line-') &&
        (orig.description !== l.description ||
          orig.quantityMinutes !== l.quantityMinutes ||
          orig.unitAmountMinor !== l.unitAmountMinor ||
          orig.lineAmountMinor !== l.lineAmountMinor)
      );
    });
    const toAdd = current.filter((l) => l.id.startsWith('line-'));

    const operations: Promise<unknown>[] = [];

    for (const line of toDelete) {
      operations.push(this.api.deleteLine(invoiceId, line.id).toPromise());
    }

    for (const line of toUpdate) {
      const input: UpdateInvoiceLineInput = {
        description: line.description,
        quantityMinutes: line.quantityMinutes,
        unitAmountMinor: line.unitAmountMinor,
        lineAmountMinor: line.lineAmountMinor,
      };
      operations.push(this.api.updateLine(invoiceId, line.id, input).toPromise());
    }

    for (const line of toAdd) {
      const input: AddInvoiceLineInput = {
        lineKind: line.lineKind,
        description: line.description,
        quantityMinutes: line.quantityMinutes,
        unitAmountMinor: line.unitAmountMinor,
        lineAmountMinor: line.lineAmountMinor,
      };
      operations.push(this.api.addLine(invoiceId, input).toPromise());
    }

    if (operations.length === 0) {
      this.isSaving = false;
      this.toast.info('No changes to save.');
      return;
    }

    Promise.all(operations)
      .then(() => this.api.getInvoice(invoiceId).toPromise())
      .then((detail) => {
        if (!detail) return;
        this.invoice = detail;
        const formLines = detail.lines.map(mapApiLineToForm);
        this.lines.set(formLines);
        this.originalLines.set(formLines.map((l) => ({ ...l })));
        this.isSaving = false;
        this.toast.success('Invoice saved.');
      })
      .catch((err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.submitError = formatPresentedApiError(presentApiError(mapped, 'invoice.createDraft'));
        this.isSaving = false;
      });
  }

  issueInvoice(): void {
    if (!this.invoice || this.isIssuing) return;
    this.isIssuing = true;
    this.submitError = null;

    const invoiceId = this.invoice.invoiceId;

    const current = this.lines();
    const original = this.originalLines();
    const originalMap = new Map(original.map((l) => [l.id, l]));
    const currentMap = new Map(current.map((l) => [l.id, l]));

    const toDelete = original.filter((l) => !currentMap.has(l.id) && !l.id.startsWith('line-'));
    const toUpdate = current.filter((l) => {
      const orig = originalMap.get(l.id);
      return (
        orig &&
        !l.id.startsWith('line-') &&
        (orig.description !== l.description ||
          orig.quantityMinutes !== l.quantityMinutes ||
          orig.unitAmountMinor !== l.unitAmountMinor ||
          orig.lineAmountMinor !== l.lineAmountMinor)
      );
    });
    const toAdd = current.filter((l) => l.id.startsWith('line-'));

    const operations: Promise<unknown>[] = [];
    for (const line of toDelete) {
      operations.push(this.api.deleteLine(invoiceId, line.id).toPromise());
    }
    for (const line of toUpdate) {
      const input: UpdateInvoiceLineInput = {
        description: line.description,
        quantityMinutes: line.quantityMinutes,
        unitAmountMinor: line.unitAmountMinor,
        lineAmountMinor: line.lineAmountMinor,
      };
      operations.push(this.api.updateLine(invoiceId, line.id, input).toPromise());
    }
    for (const line of toAdd) {
      const input: AddInvoiceLineInput = {
        lineKind: line.lineKind,
        description: line.description,
        quantityMinutes: line.quantityMinutes,
        unitAmountMinor: line.unitAmountMinor,
        lineAmountMinor: line.lineAmountMinor,
      };
      operations.push(this.api.addLine(invoiceId, input).toPromise());
    }

    Promise.all(operations)
      .then(() => this.api.issueInvoice(invoiceId).toPromise())
      .then(() => {
        this.isIssuing = false;
        this.toast.success('Invoice issued.');
        this.router.navigate(['/manager/invoices', invoiceId]);
      })
      .catch((err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.submitError = formatPresentedApiError(presentApiError(mapped, 'invoice.issue'));
        this.isIssuing = false;
      });
  }

  trackByLineId(_index: number, line: FormInvoiceLine): string {
    return line.id;
  }
}
