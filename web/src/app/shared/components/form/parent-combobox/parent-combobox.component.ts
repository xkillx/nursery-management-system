import { CommonModule } from '@angular/common';
import { Component, ElementRef, EventEmitter, inject, Input, OnDestroy, Output, ViewChild } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Subject, Subscription, debounceTime } from 'rxjs';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroMagnifyingGlass, heroPlus, heroXMark, heroCheck } from '@ng-icons/heroicons/outline';
import { ParentsApiService } from '../../../../features/staff/data/parents-api.service';
import { ParentRecord } from '../../../../features/staff/models/parents.models';
import { BadgeComponent } from '../../ui/badge/badge.component';

@Component({
  selector: 'app-parent-combobox',
  imports: [CommonModule, FormsModule, NgIcon, BadgeComponent],
  providers: [
    provideIcons({ heroMagnifyingGlass, heroPlus, heroXMark, heroCheck }),
  ],
  template: `
    <div class="relative">
      <div class="relative">
        <span class="pointer-events-none absolute left-4 inset-y-0 z-10 flex items-center text-gray-400">
          <ng-icon name="heroMagnifyingGlass" size="18" aria-hidden="true" />
        </span>
        <input
          #searchInput
          type="text"
          [value]="searchTerm"
          (input)="onSearchInput($event)"
          (focus)="onFocus()"
          (keydown.escape)="closeDropdown()"
          (keydown.arrowdown)="onArrowDown($event)"
          (keydown.arrowup)="onArrowUp($event)"
          (keydown.enter)="onEnter($event)"
          [placeholder]="placeholder"
          [disabled]="disabled"
          class="h-12 w-full rounded-xl border border-gray-200 bg-white pl-11 pr-4 py-3 text-sm text-gray-900 shadow-theme-xs dark:border-gray-700 dark:bg-gray-900 dark:text-gray-200 focus:border-brand-300 focus:outline-hidden focus:ring-3 focus:ring-brand-500/10 dark:focus:border-brand-800"
          role="combobox"
          [attr.aria-expanded]="showDropdown"
          aria-haspopup="listbox"
          aria-controls="parent-combobox-listbox"
          autocomplete="off"
        />
        @if (searchTerm) {
          <button
            type="button"
            (click)="clearSearch()"
            class="absolute right-3 inset-y-0 z-10 flex items-center text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
          >
            <ng-icon name="heroXMark" size="18" />
          </button>
        }
      </div>

      @if (showDropdown) {
        <div
          id="parent-combobox-listbox"
          class="absolute z-50 mt-1 w-full overflow-hidden rounded-xl border border-gray-200 bg-white shadow-theme-md dark:border-gray-700 dark:bg-gray-900"
          role="listbox"
          tabindex="-1"
          (keydown)="onDropdownKeydown($event)"
        >
          @if (isSearching) {
            <div class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">
              Searching...
            </div>
          } @else if (searchTerm && filteredResults.length === 0 && !showCreateOption) {
            <div class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">
              No parents found.
            </div>
          } @else {
            @if (filteredResults.length > 0) {
              <div class="max-h-48 overflow-auto">
                @for (parent of filteredResults; track parent.id; let i = $index) {
                  <div
                    class="flex items-center gap-3 px-4 py-3 cursor-pointer transition-colors"
                    [class.bg-brand-50]="i === activeIndex"
                    [class.dark:bg-brand-500/10]="i === activeIndex"
                    [class.hover:bg-gray-50]="i !== activeIndex"
                    [class.dark:hover:bg-gray-800]="i !== activeIndex"
                    role="option"
                    [attr.aria-selected]="i === activeIndex"
                    (mousedown)="selectParent(parent)"
                  >
                    <div class="w-8 h-8 rounded-lg bg-brand-50 text-brand-600 font-bold text-xs flex items-center justify-center border border-brand-100 shrink-0 dark:bg-brand-500/10 dark:text-brand-400 dark:border-brand-500/20">
                      {{ parent.first_name[0] }}{{ parent.last_name ? parent.last_name[0] : '' }}
                    </div>
                    <div class="min-w-0 flex-1">
                      <div class="text-sm font-semibold text-gray-900 dark:text-white truncate">
                        {{ parent.first_name }} {{ parent.last_name }}
                      </div>
                      <div class="text-xs text-gray-500 dark:text-gray-400 truncate">
                        {{ parent.email || parent.phone || 'No contact info' }}
                      </div>
                    </div>
                    @if (!parent.is_active) {
                      <app-badge color="warning" size="sm">Inactive</app-badge>
                    }
                  </div>
                }
              </div>
            }

            @if (showCreateOption) {
              <div
                class="flex items-center gap-3 px-4 py-3 cursor-pointer border-t border-gray-100 dark:border-gray-800 transition-colors"
                [class.bg-brand-50]="activeIndex === filteredResults.length"
                [class.dark:bg-brand-500/10]="activeIndex === filteredResults.length"
                [class.hover:bg-gray-50]="activeIndex !== filteredResults.length"
                [class.dark:hover:bg-gray-800]="activeIndex !== filteredResults.length"
                role="option"
                [attr.aria-selected]="activeIndex === filteredResults.length"
                (mousedown)="createNewParent()"
              >
                <div class="w-8 h-8 rounded-lg bg-success-50 text-success-600 flex items-center justify-center border border-success-100 shrink-0 dark:bg-success-500/10 dark:text-success-400 dark:border-success-500/20">
                  <ng-icon name="heroPlus" size="16" />
                </div>
                <div class="text-sm font-medium text-gray-900 dark:text-white">
                  Create new: <span class="text-brand-600 dark:text-brand-400">{{ searchTerm }}</span>
                </div>
              </div>
            }
          }
        </div>
      }
    </div>
  `,
})
export class ParentComboboxComponent implements OnDestroy {
  @ViewChild('searchInput', { static: false }) searchInput!: ElementRef<HTMLInputElement>;

  @Input() excludeIds: string[] = [];
  @Input() branchId = '';
  @Input() placeholder = 'Search for a parent...';
  @Input() disabled = false;

  @Output() parentSelected = new EventEmitter<ParentRecord>();
  @Output() parentCreateRequested = new EventEmitter<{ name: string }>();

  private readonly parentsApi = inject(ParentsApiService);
  private readonly input$ = new Subject<string>();
  private subscription: Subscription;

  searchTerm = '';
  searchResults: ParentRecord[] = [];
  isSearching = false;
  showDropdown = false;
  activeIndex = -1;

  constructor() {
    this.subscription = this.input$.pipe(debounceTime(200)).subscribe((text) => {
      this.performSearch(text);
    });
  }

  ngOnDestroy(): void {
    this.subscription.unsubscribe();
  }

  get filteredResults(): ParentRecord[] {
    return this.searchResults.filter(p => !this.excludeIds.includes(p.id));
  }

  get showCreateOption(): boolean {
    return !!this.searchTerm.trim() && !this.exactMatch;
  }

  get exactMatch(): boolean {
    const term = this.searchTerm.trim().toLowerCase();
    return this.filteredResults.some(p =>
      `${p.first_name} ${p.last_name || ''}`.trim().toLowerCase() === term
    );
  }

  onSearchInput(event: Event): void {
    const value = (event.target as HTMLInputElement).value;
    this.searchTerm = value;
    this.activeIndex = -1;
    this.input$.next(value);
  }

  onFocus(): void {
    this.showDropdown = true;
    if (!this.searchTerm && this.searchResults.length === 0) {
      this.performSearch('');
    }
  }

  clearSearch(): void {
    this.searchTerm = '';
    this.searchResults = [];
    this.activeIndex = -1;
    this.searchInput?.nativeElement.focus();
  }

  closeDropdown(): void {
    this.showDropdown = false;
    this.activeIndex = -1;
  }

  onArrowDown(event: Event): void {
    event.preventDefault();
    const max = this.filteredResults.length + (this.showCreateOption ? 0 : -1);
    if (this.activeIndex < max) {
      this.activeIndex++;
    }
  }

  onArrowUp(event: Event): void {
    event.preventDefault();
    if (this.activeIndex > 0) {
      this.activeIndex--;
    }
  }

  onEnter(event: Event): void {
    event.preventDefault();
    if (this.activeIndex >= 0 && this.activeIndex < this.filteredResults.length) {
      this.selectParent(this.filteredResults[this.activeIndex]);
    } else if (this.showCreateOption && this.activeIndex === this.filteredResults.length) {
      this.createNewParent();
    } else if (this.showCreateOption) {
      this.createNewParent();
    }
  }

  onDropdownKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      this.closeDropdown();
      this.searchInput?.nativeElement.focus();
    }
  }

  selectParent(parent: ParentRecord): void {
    this.parentSelected.emit(parent);
    this.searchTerm = '';
    this.showDropdown = false;
    this.activeIndex = -1;
  }

  createNewParent(): void {
    const name = this.searchTerm.trim();
    if (name) {
      this.parentCreateRequested.emit({ name });
      this.searchTerm = '';
      this.showDropdown = false;
      this.activeIndex = -1;
    }
  }

  private performSearch(text: string): void {
    if (!text.trim()) {
      this.parentsApi.list(1, 20, 'active').subscribe({
        next: (res) => {
          this.searchResults = res.parents;
          this.isSearching = false;
        },
        error: () => {
          this.searchResults = [];
          this.isSearching = false;
        },
      });
      return;
    }

    this.isSearching = true;
    this.parentsApi.list(1, 20, 'active', text.trim()).subscribe({
      next: (res) => {
        this.searchResults = res.parents;
        this.isSearching = false;
      },
      error: () => {
        this.searchResults = [];
        this.isSearching = false;
      },
    });
  }
}
