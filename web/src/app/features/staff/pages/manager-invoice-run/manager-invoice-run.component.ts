import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';

import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { ConfirmationDialogComponent } from '../../../../shared/components/ui/modal/confirmation-dialog.component';
import { ToastService } from '../../../../shared/services/toast.service';
import { InvoiceRunMockService } from '../../data/invoice-run-mock.service';
import {
  InvoiceRunStep,
  InvoiceRunPreflight,
  InvoiceDraftReviewItem,
  DraftGenerationResult,
  IssueResultSummary,
  InvoiceRunException,
  InvoiceRunBlockerCode,
} from '../../models/invoice-run.models';
import {
  formatGbp,
  formatMinutes,
  formatBillingMonthLabel,
  defaultCompletedBillingMonth,
  blockerNextAction,
  blockerLabel,
  BlockerNextAction,
} from '../../utils/invoice-run-formatters';

@Component({
  selector: 'app-manager-invoice-run',
  imports: [
    CommonModule,
    FormsModule,
    RouterLink,
    PageHeaderComponent,
    AlertComponent,
    EmptyStateComponent,
    LoadingStateComponent,
    StatusBadgeComponent,
    ConfirmationDialogComponent,
  ],
  templateUrl: './manager-invoice-run.component.html',
})
export class ManagerInvoiceRunComponent implements OnInit {
  private readonly mockService = inject(InvoiceRunMockService);
  private readonly toast = inject(ToastService);

  selectedBillingMonth = defaultCompletedBillingMonth();
  currentStep: InvoiceRunStep = 'preflight';

  preflight: InvoiceRunPreflight | null = null;
  generationResult: DraftGenerationResult | null = null;
  drafts: InvoiceDraftReviewItem[] = [];
  issueResult: IssueResultSummary | null = null;

  selectedInvoiceIds = new Set<string>();
  expandedInvoiceIds = new Set<string>();

  isLoadingPreflight = false;
  isGenerating = false;
  isIssuing = false;

  showBulkConfirm = false;
  singleIssueDraft: InvoiceDraftReviewItem | null = null;

  errorMessage: string | null = null;

  readonly formatGbp = formatGbp;
  readonly formatMinutes = formatMinutes;
  readonly formatBillingMonthLabel = formatBillingMonthLabel;
  readonly blockerLabel = blockerLabel;

  ngOnInit(): void {
    this.loadPreflight();
  }

  onMonthChange(month: string): void {
    this.selectedBillingMonth = month;
    this.mockService.resetMonth(month);
    this.generationResult = null;
    this.drafts = [];
    this.issueResult = null;
    this.selectedInvoiceIds.clear();
    this.expandedInvoiceIds.clear();
    this.showBulkConfirm = false;
    this.singleIssueDraft = null;
    this.currentStep = 'preflight';
    this.loadPreflight();
  }

  get billingMonthLabel(): string {
    return formatBillingMonthLabel(this.selectedBillingMonth);
  }

  get readyDrafts(): InvoiceDraftReviewItem[] {
    return this.drafts.filter(d => d.status === 'draft');
  }

  get selectedDrafts(): InvoiceDraftReviewItem[] {
    return this.readyDrafts.filter(d => this.selectedInvoiceIds.has(d.invoiceId));
  }

  get selectedTotalMinor(): number {
    return this.selectedDrafts.reduce((sum, d) => sum + d.netDueMinor, 0);
  }

  get allReadySelected(): boolean {
    return this.readyDrafts.length > 0 && this.readyDrafts.every(d => this.selectedInvoiceIds.has(d.invoiceId));
  }

  blockerAction(blockerCode: InvoiceRunBlockerCode, exception: InvoiceRunException): BlockerNextAction {
    return blockerNextAction(blockerCode, exception.childId, this.selectedBillingMonth);
  }

  generateDrafts(): void {
    this.isGenerating = true;
    this.errorMessage = null;

    this.mockService.generateDrafts(this.selectedBillingMonth).subscribe({
      next: (result) => {
        this.generationResult = result;
        this.isGenerating = false;
        this.currentStep = 'review';

        this.mockService.listDrafts(this.selectedBillingMonth).subscribe((drafts) => {
          this.drafts = drafts;
          for (const d of this.readyDrafts) {
            this.selectedInvoiceIds.add(d.invoiceId);
          }
        });

        this.toast.success(`Generated ${result.generatedCount} draft invoices.`);
      },
      error: () => {
        this.isGenerating = false;
        this.errorMessage = 'Failed to generate draft invoices. Try again.';
      },
    });
  }

  toggleDraftSelection(invoiceId: string): void {
    if (this.selectedInvoiceIds.has(invoiceId)) {
      this.selectedInvoiceIds.delete(invoiceId);
    } else {
      this.selectedInvoiceIds.add(invoiceId);
    }
  }

  toggleAllReady(): void {
    if (this.allReadySelected) {
      this.selectedInvoiceIds.clear();
    } else {
      for (const d of this.readyDrafts) {
        this.selectedInvoiceIds.add(d.invoiceId);
      }
    }
  }

  toggleExpand(invoiceId: string): void {
    if (this.expandedInvoiceIds.has(invoiceId)) {
      this.expandedInvoiceIds.delete(invoiceId);
    } else {
      this.expandedInvoiceIds.add(invoiceId);
    }
  }

  isExpanded(invoiceId: string): boolean {
    return this.expandedInvoiceIds.has(invoiceId);
  }

  openBulkConfirm(): void {
    this.showBulkConfirm = true;
  }

  closeBulkConfirm(): void {
    this.showBulkConfirm = false;
  }

  confirmBulkIssue(): void {
    this.isIssuing = true;
    const ids = Array.from(this.selectedInvoiceIds);

    this.mockService.bulkIssue(this.selectedBillingMonth, ids).subscribe({
      next: (result) => {
        this.issueResult = result;
        this.isIssuing = false;
        this.showBulkConfirm = false;
        this.currentStep = 'result';

        this.mockService.listDrafts(this.selectedBillingMonth).subscribe((drafts) => {
          this.drafts = drafts;
          this.selectedInvoiceIds.clear();
          for (const d of this.readyDrafts) {
            this.selectedInvoiceIds.add(d.invoiceId);
          }
        });

        this.toast.success(`Issued ${result.issuedCount} invoices.`);
      },
      error: () => {
        this.isIssuing = false;
        this.showBulkConfirm = false;
        this.errorMessage = 'Failed to issue invoices. Try again.';
      },
    });
  }

  openSingleConfirm(draft: InvoiceDraftReviewItem): void {
    this.singleIssueDraft = draft;
  }

  closeSingleConfirm(): void {
    this.singleIssueDraft = null;
  }

  confirmSingleIssue(): void {
    if (!this.singleIssueDraft) return;
    this.isIssuing = true;

    this.mockService.issueOne(this.singleIssueDraft.invoiceId).subscribe({
      next: (result) => {
        this.isIssuing = false;
        this.singleIssueDraft = null;

        if (this.issueResult) {
          this.issueResult = {
            ...this.issueResult,
            issuedCount: this.issueResult.issuedCount + result.issuedCount,
            totalIssuedMinor: this.issueResult.totalIssuedMinor + result.totalIssuedMinor,
            issued: [...this.issueResult.issued, ...result.issued],
            skipped: [...this.issueResult.skipped, ...result.skipped],
          };
        } else {
          this.issueResult = result;
        }

        this.currentStep = 'result';

        this.mockService.listDrafts(this.selectedBillingMonth).subscribe((drafts) => {
          this.drafts = drafts;
          this.selectedInvoiceIds.clear();
          for (const d of this.readyDrafts) {
            this.selectedInvoiceIds.add(d.invoiceId);
          }
        });

        this.toast.success(`Issued ${result.issuedCount} invoice.`);
      },
      error: () => {
        this.isIssuing = false;
        this.singleIssueDraft = null;
        this.errorMessage = 'Failed to issue invoice. Try again.';
      },
    });
  }

  private loadPreflight(): void {
    this.isLoadingPreflight = true;
    this.errorMessage = null;

    this.mockService.loadPreflight(this.selectedBillingMonth).subscribe({
      next: (preflight) => {
        this.preflight = preflight;
        this.isLoadingPreflight = false;
      },
      error: () => {
        this.isLoadingPreflight = false;
        this.errorMessage = 'Failed to load preflight data. Try again.';
      },
    });
  }
}
