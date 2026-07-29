import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal, computed, ElementRef, HostListener } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
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
  heroChevronRight,
  heroArrowRight,
  heroCalendar,
  heroUser,
} from '@ng-icons/heroicons/outline';
import { catchError, of } from 'rxjs';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { ManagerInvoiceCreateApiService } from '../../data/manager-invoice-create-api.service';
import { StaffApiService } from '../../data/staff-api.service';
import { StaffRoomsApiService } from '../../data/staff-rooms-api.service';
import { AuthService } from '../../../../core/services/auth.service';
import { ToastService } from '../../../../shared/services/toast.service';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { ChildAvatarComponent } from '../../../../shared/components/ui/avatar/child-avatar/child-avatar.component';
import { MonthPickerComponent, MonthYear } from '../../../../shared/components/form/month-picker/month-picker.component';
import { formatGbp } from '../../../owner/utils/owner-formatters';
import { FormInvoiceLine } from '../../models/manager-invoice-create.models';
import { formatChildName } from '../../utils/manager-list-formatters';
import { ChildRecord } from '../../models/children.models';

@Component({
  selector: 'app-manager-invoice-create',
  imports: [
    CommonModule,
    FormsModule,
    RouterModule,
    NgIcon,
    AlertComponent,
    LoadingStateComponent,
    ChildAvatarComponent,
    MonthPickerComponent,
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
      heroChevronRight,
      heroArrowRight,
      heroCalendar,
      heroUser,
    }),
  ],
})
export class ManagerInvoiceCreateComponent implements OnInit {
  private readonly api = inject(ManagerInvoiceCreateApiService);
  private readonly staffApi = inject(StaffApiService);
  private readonly roomsApi = inject(StaffRoomsApiService);
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly elRef = inject<ElementRef<HTMLElement>>(ElementRef);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly toast = inject(ToastService);

  readonly formatGbp = formatGbp;
  readonly formatChildName = formatChildName;
  readonly Math = Math;
  readonly Number = Number;

  readonly DEFAULT_PAYMENT_TERMS = 'Payments are due within 7 days. Late fees may apply as per the parent agreement.';

  readonly mockInvoiceNumber = 'INV-2026-042';
  readonly issueDate = new Date().toISOString().split('T')[0];
  readonly dueDate = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];

  mode: 'new' | 'edit' = 'new';
  editInvoiceId: string | null = null;

  childSearchTerm = '';
  searchResults: ChildRecord[] = [];
  isSearching = false;
  isDropdownOpen = false;
  highlightedIndex = -1;
  selectedChild: ChildRecord | null = null;

  @HostListener('document:click', ['$event'])
  onDocumentClick(event: MouseEvent): void {
    if (!this.isDropdownOpen) return;
    const target = event.target as Node;
    if (!target || !document.body.contains(target)) return;
    const searchContainer = this.elRef.nativeElement.querySelector('.relative');
    if (searchContainer && !searchContainer.contains(target)) {
      this.isDropdownOpen = false;
    }
  }

  parentCarerName = signal<string>('');
  roomName = signal<string>('');
  ageGroup = signal<string>('');

  billingMonth = signal<string>('');
  isLoadingPrefill = false;
  prefillError: string | null = null;

  lines = signal<FormInvoiceLine[]>([]);
  entitlementLabel = '';
  hasFundingProfile = false;

  internalNotes = '';
  paymentTerms = this.DEFAULT_PAYMENT_TERMS;
  parentFacingNote = '';
  useCustomDates = false;
  overrideStartDate = '';
  overrideEndDate = '';

  isSaving = false;
  isIssuing = false;
  submitError: string | null = null;

  readonly billingPeriodStart = computed(() => {
    if (this.useCustomDates && this.overrideStartDate) {
      return this.overrideStartDate;
    }
    const month = this.billingMonth();
    if (!month) return '';
    return `${month}-01`;
  });

  readonly billingPeriodEnd = computed(() => {
    if (this.useCustomDates && this.overrideEndDate) {
      return this.overrideEndDate;
    }
    const month = this.billingMonth();
    if (!month) return '';
    const [year, m] = month.split('-').map(Number);
    const lastDay = new Date(year, m, 0).getDate();
    const mm = String(m).padStart(2, '0');
    return `${year}-${mm}-${lastDay}`;
  });

  readonly subtotalMinor = computed(() =>
    this.lines()
      .filter((l) => !l.isFundingOffset)
      .reduce((sum, l) => sum + l.lineAmountMinor, 0)
  );

  readonly fundedDeductionMinor = computed(() =>
    this.lines()
      .filter((l) => l.isFundingOffset)
      .reduce((sum, l) => sum + Math.abs(l.lineAmountMinor), 0)
  );

  readonly totalDueMinor = computed(() =>
    Math.max(0, this.subtotalMinor() - this.fundedDeductionMinor())
  );

  readonly showWarningBanner = computed(() => {
    const s = this.subtotalMinor();
    const f = this.fundedDeductionMinor();
    return s > 0 && f > 0 && f > s / 4;
  });

  readonly autoGeneratedLines = computed(() =>
    this.lines().filter((l) => l.lineKind !== 'extra' && l.lineKind !== 'ad_hoc')
  );

  readonly extraLines = computed(() =>
    this.lines().filter((l) => l.lineKind === 'extra' || l.lineKind === 'ad_hoc')
  );

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
    this.billingMonth.set(`${y}-${m}`);
  }

  onSearchInput(): void {
    const term = this.childSearchTerm.trim();
    this.highlightedIndex = -1;
    if (term.length < 2) {
      this.searchResults = [];
      this.isDropdownOpen = false;
      return;
    }

    this.isSearching = true;
    this.isDropdownOpen = true;

    this.staffApi
      .listChildren({ status: 'active', limit: 200, offset: 0 })
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
            });
          this.isSearching = false;
          this.isDropdownOpen = true;
        },
        error: () => {
          this.isSearching = false;
        },
      });
  }

  onSearchFocus(): void {
    if (this.childSearchTerm.trim().length >= 2) {
      this.isDropdownOpen = true;
    }
  }

  onSearchKeydown(event: KeyboardEvent): void {
    if (!this.isDropdownOpen && (event.key === 'ArrowDown' || event.key === 'ArrowUp')) {
      if (this.childSearchTerm.trim().length >= 2) {
        this.isDropdownOpen = true;
      }
    }

    if (event.key === 'ArrowDown') {
      if (this.searchResults.length > 0) {
        this.highlightedIndex = (this.highlightedIndex + 1) % this.searchResults.length;
      }
      event.preventDefault();
    } else if (event.key === 'ArrowUp') {
      if (this.searchResults.length > 0) {
        this.highlightedIndex =
          (this.highlightedIndex - 1 + this.searchResults.length) % this.searchResults.length;
      }
      event.preventDefault();
    } else if (event.key === 'Enter') {
      if (this.isDropdownOpen && this.highlightedIndex >= 0 && this.highlightedIndex < this.searchResults.length) {
        this.selectChild(this.searchResults[this.highlightedIndex]);
        event.preventDefault();
      }
    } else if (event.key === 'Escape') {
      this.isDropdownOpen = false;
      this.highlightedIndex = -1;
      event.preventDefault();
    }
  }

  clearSearch(): void {
    this.childSearchTerm = '';
    this.searchResults = [];
    this.isDropdownOpen = false;
    this.highlightedIndex = -1;
    if (this.selectedChild) {
      this.clearChild();
    }
  }

  getHighlightedParts(text: string, query: string): { text: string; match: boolean }[] {
    if (!query || !text) return [{ text, match: false }];
    const q = query.trim().toLowerCase();
    if (!q) return [{ text, match: false }];

    const index = text.toLowerCase().indexOf(q);
    if (index === -1) return [{ text, match: false }];

    return [
      { text: text.slice(0, index), match: false },
      { text: text.slice(index, index + q.length), match: true },
      { text: text.slice(index + q.length), match: false },
    ].filter((p) => p.text.length > 0);
  }

  selectChild(child: ChildRecord): void {
    this.selectedChild = child;
    this.childSearchTerm = child.fullName;
    this.searchResults = [];
    this.isDropdownOpen = false;
    this.highlightedIndex = -1;

    // Load parent/carer contacts
    this.parentCarerName.set('Loading...');
    this.staffApi.getChildContacts(child.id).subscribe({
      next: (contacts) => {
        if (contacts.parentCarers && contacts.parentCarers.length > 0) {
          const parent = contacts.parentCarers[0];
          this.parentCarerName.set(parent.fullName);
        } else {
          this.parentCarerName.set('Not assigned');
        }
      },
      error: () => {
        this.parentCarerName.set('Not assigned');
      }
    });

    // Compute age group dynamically
    const ageGroupStr = this.calculateAgeGroup(child.dateOfBirth);
    this.ageGroup.set(ageGroupStr);

    // Load actual room assignment
    this.roomName.set('Loading...');
    const membership = this.auth.activeMembership();
    const siteId = membership?.branch_id;
    if (!siteId) {
      this.roomName.set('Not assigned');
      this.loadPrefill();
      return;
    }

    this.staffApi.listChildRoomAssignments(child.id).subscribe({
      next: (assignments) => {
        const current = assignments.find((a) => a.is_current);
        if (!current) {
          this.roomName.set('Not assigned');
          this.loadPrefill();
          return;
        }
        this.roomsApi.listRooms(siteId).subscribe({
          next: (rooms) => {
            const room = rooms.find((r) => r.id === current.room_id);
            this.roomName.set(room?.name ?? 'Not assigned');
            this.loadPrefill();
          },
          error: () => {
            this.roomName.set('Not assigned');
            this.loadPrefill();
          },
        });
      },
      error: () => {
        this.roomName.set('Not assigned');
        this.loadPrefill();
      },
    });
  }

  calculateAgeGroup(dobString: string): string {
    if (!dobString) return 'Unknown';
    const dob = new Date(dobString);
    const today = new Date();
    let age = today.getFullYear() - dob.getFullYear();
    const m = today.getMonth() - dob.getMonth();
    if (m < 0 || (m === 0 && today.getDate() < dob.getDate())) {
      age--;
    }
    if (age < 1) return 'Under 1 Year';
    if (age < 2) return '1-2 Years';
    if (age < 3) return '2-3 Years';
    return '3-5 Years';
  }

  onBillingMonthChange(val: MonthYear): void {
    const mm = String(val.month + 1).padStart(2, '0');
    this.billingMonth.set(`${val.year}-${mm}`);
    if (this.selectedChild) {
      this.loadPrefill();
    }
  }

  toggleCustomDates(): void {
    this.useCustomDates = !this.useCustomDates;
    if (this.useCustomDates) {
      this.overrideStartDate = this.billingPeriodStart();
      this.overrideEndDate = this.billingPeriodEnd();
    } else {
      this.overrideStartDate = '';
      this.overrideEndDate = '';
    }
  }

  clearChild(): void {
    this.selectedChild = null;
    this.childSearchTerm = '';
    this.searchResults = [];
    this.isDropdownOpen = false;
    this.highlightedIndex = -1;
    this.lines.set([]);
    this.entitlementLabel = '';
    this.hasFundingProfile = false;
    this.parentCarerName.set('');
    this.roomName.set('');
    this.ageGroup.set('');
  }

  loadPrefill(): void {
    if (!this.selectedChild || !this.billingMonth()) return;

    this.isLoadingPrefill = true;
    this.prefillError = null;

    this.api.getPrefill(this.selectedChild.id, this.billingMonth()).subscribe({
      next: (prefill) => {
        this.lines.set(
          prefill.lines.map((l, i) => ({
            id: `line-${i}`,
            lineKind: l.lineKind,
            description: l.description,
            sortOrder: l.sortOrder,
            quantityHours: l.quantityHours,
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
        quantityHours: 0,
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

  addPresetLine(description: string, unitPriceMinor: number, quantity: number): void {
    this.lines.update((prev) => [
      ...prev,
      {
        id: `line-${Date.now()}`,
        lineKind: 'extra',
        description,
        sortOrder: prev.length + 1,
        quantityHours: quantity,
        unitAmountMinor: unitPriceMinor,
        lineAmountMinor: quantity * unitPriceMinor,
        fundedAllowanceMinutes: 0,
        fundedDeductionMinutes: 0,
        coreBillableMinutes: 0,
        sessionCount: 0,
        isFundingOffset: false,
      },
    ]);
    this.toast.success(`Preset "${description}" added.`);
  }

  removeLine(lineId: string): void {
    this.lines.update((prev) => prev.filter((l) => l.id !== lineId));
  }

  updateLine(lineId: string, field: keyof FormInvoiceLine, value: number | string): void {
    this.lines.update((prev) =>
      prev.map((l) => {
        if (l.id !== lineId) return l;
        const updated = { ...l, [field]: value };
        if (field === 'quantityHours' || field === 'unitAmountMinor') {
          const q = typeof updated.quantityHours === 'number' ? updated.quantityHours : 0;
          const u = typeof updated.unitAmountMinor === 'number' ? updated.unitAmountMinor : 0;
          updated.lineAmountMinor = q * u;
        }
        return updated;
      }),
    );
  }

  saveDraft(): void {
    if (!this.canSaveDraft()) return;
    this.isSaving = true;
    this.submitError = null;

    this.api
      .createDraft({
        childId: this.selectedChild!.id,
        billingMonth: this.billingMonth(),
        lines: this.lines().map((l) => ({
          lineKind: l.lineKind,
          description: l.description,
          sortOrder: l.sortOrder,
          quantityHours: l.quantityHours,
          unitAmountMinor: l.unitAmountMinor,
          lineAmountMinor: l.lineAmountMinor,
          fundedAllowanceMinutes: l.fundedAllowanceMinutes,
          fundedDeductionMinutes: l.fundedDeductionMinutes,
          coreBillableMinutes: l.coreBillableMinutes,
          sessionCount: l.sessionCount,
        })),
        paymentTerms: this.paymentTerms,
        internalNotes: this.internalNotes,
        parentNote: this.parentFacingNote,
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
    if (!this.canIssue()) return;
    this.isIssuing = true;
    this.submitError = null;

    this.api
      .createAndIssue({
        childId: this.selectedChild!.id,
        billingMonth: this.billingMonth(),
        lines: this.lines().map((l) => ({
          lineKind: l.lineKind,
          description: l.description,
          sortOrder: l.sortOrder,
          quantityHours: l.quantityHours,
          unitAmountMinor: l.unitAmountMinor,
          lineAmountMinor: l.lineAmountMinor,
          fundedAllowanceMinutes: l.fundedAllowanceMinutes,
          fundedDeductionMinutes: l.fundedDeductionMinutes,
          coreBillableMinutes: l.coreBillableMinutes,
          sessionCount: l.sessionCount,
        })),
        paymentTerms: this.paymentTerms,
        internalNotes: this.internalNotes,
        parentNote: this.parentFacingNote,
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

  canSaveDraft(): boolean {
    return !!this.selectedChild && !!this.billingMonth() && this.lines().length > 0;
  }

  canIssue(): boolean {
    return this.canSaveDraft();
  }

  canSubmit(): boolean {
    return this.canSaveDraft();
  }
}
