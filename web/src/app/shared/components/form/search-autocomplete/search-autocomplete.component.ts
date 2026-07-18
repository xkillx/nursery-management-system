import { CommonModule } from '@angular/common';
import { Component, ElementRef, EventEmitter, forwardRef, Input, Output, ViewChild, OnDestroy } from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';
import { Subject, Subscription, debounceTime } from 'rxjs';

@Component({
  selector: 'app-search-autocomplete',
  imports: [CommonModule],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => SearchAutocompleteComponent),
      multi: true,
    },
  ],
  template: `
    <div class="relative">
      <input
        #inputRef
        type="text"
        [placeholder]="placeholder"
        [value]="displayText"
        [disabled]="disabled"
        class="w-full h-10 rounded-lg border border-gray-200 bg-white px-3 text-sm text-gray-900 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-200 focus:border-brand-500 focus:ring-3 focus:ring-brand-500/10"
        [ngClass]="{ 'border-error-500': error }"
        (input)="onInput($event)"
        (focus)="onFocus()"
        (blur)="markTouched()"
        (keydown.escape)="closeDropdown()"
        role="combobox"
        [attr.aria-expanded]="showDropdown"
        aria-haspopup="listbox"
        aria-controls="autocomplete-listbox"
      />
      @if (showDropdown && filteredItems.length > 0) {
        <ul
          id="autocomplete-listbox"
          class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-xl border border-gray-200 bg-white shadow-theme-md dark:border-gray-700 dark:bg-gray-900"
          role="listbox"
        >
          @for (item of filteredItems; track $index; let i = $index) {
            <li
              class="cursor-pointer px-3 py-2 text-sm text-gray-900 hover:bg-brand-50 hover:text-brand-700 dark:text-gray-200 dark:hover:bg-brand-500/10"
              role="option"
              [attr.aria-selected]="i === 0"
              (mousedown)="selectItem(item)"
            >
              {{ labelFn(item) }}
            </li>
          }
        </ul>
      }
    </div>
  `,
})
export class SearchAutocompleteComponent<T> implements ControlValueAccessor, OnDestroy {
  @ViewChild('inputRef', { static: false }) inputRef!: ElementRef<HTMLInputElement>;

  @Input() items: T[] = [];
  @Input() labelFn: (item: T) => string = (item: T) => String(item);
  @Input() placeholder = 'Search...';
  @Input() error = false;

  @Output() valueChange = new EventEmitter<T | null>();
  @Output() blurred = new EventEmitter<void>();

  displayText = '';
  filteredItems: T[] = [];
  showDropdown = false;
  disabled = false;

  private selectedItem: T | null = null;
  private input$ = new Subject<string>();
  private subscription: Subscription;
  private onChange: (value: T | null) => void = () => { /* Set via registerOnChange */ };
  private onTouched: () => void = () => { /* Set via registerOnTouched */ };

  constructor() {
    this.subscription = this.input$.pipe(debounceTime(200)).subscribe((text) => {
      this.filterItems(text);
    });
  }

  ngOnDestroy(): void {
    this.subscription.unsubscribe();
  }

  onInput(event: Event): void {
    const text = (event.target as HTMLInputElement).value;
    this.displayText = text;
    this.selectedItem = null;
    this.onChange(null);
    this.valueChange.emit(null);
    this.input$.next(text);
  }

  onFocus(): void {
    if (this.displayText.length > 0) {
      this.filterItems(this.displayText);
    }
    this.showDropdown = true;
  }

  closeDropdown(): void {
    this.showDropdown = false;
  }

  selectItem(item: T): void {
    this.selectedItem = item;
    this.displayText = this.labelFn(item);
    this.showDropdown = false;
    this.onChange(item);
    this.valueChange.emit(item);
  }

  markTouched(): void {
    setTimeout(() => {
      this.showDropdown = false;
    }, 150);
    this.onTouched();
    this.blurred.emit();
  }

  writeValue(value: T | null): void {
    this.selectedItem = value;
    this.displayText = value ? this.labelFn(value) : '';
  }

  registerOnChange(fn: (value: T | null) => void): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: () => void): void {
    this.onTouched = fn;
  }

  setDisabledState(isDisabled: boolean): void {
    this.disabled = isDisabled;
  }

  private filterItems(text: string): void {
    if (!text) {
      this.filteredItems = this.items.slice(0, 20);
      return;
    }
    const lower = text.toLowerCase();
    this.filteredItems = this.items
      .filter((item) => this.labelFn(item).toLowerCase().startsWith(lower))
      .slice(0, 20);
  }
}
