import { CommonModule } from '@angular/common';
import {
  Component,
  EventEmitter,
  HostListener,
  Input,
  Output, OnChanges,
} from '@angular/core';

@Component({
  selector: 'app-drawer',
  imports: [CommonModule],
  template: `
    @if (isOpen) {
      <div class="fixed inset-0 z-99999 flex">
        <div
          class="fixed inset-0 bg-gray-400/50 backdrop-blur-[32px]"
          tabindex="0"
          role="button"
          (click)="onBackdropClick()"
          (keydown.enter)="onBackdropClick()"
          (keydown.space)="onBackdropClick()"
        ></div>
        <div
          role="dialog"
          aria-modal="true"
          [attr.aria-label]="ariaLabel || null"
          class="fixed top-0 bottom-0 z-10 flex flex-col bg-white shadow-xl dark:bg-gray-900"
          [ngClass]="{
            'right-0': position === 'right',
            'left-0': position === 'left',
            'w-80': size === 'sm',
            'w-96': size === 'md',
            'w-[480px]': size === 'lg',
            'w-full max-w-md': size === 'xl',
          }"
        >
          <div class="flex items-center justify-between border-b border-gray-200 px-5 py-4 dark:border-gray-800">
            @if (title) {
              <h2 class="text-lg font-semibold text-gray-800 dark:text-white/90">{{ title }}</h2>
            }
            <button
              type="button"
              class="ml-auto flex h-8 w-8 items-center justify-center rounded-full text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-gray-800 dark:hover:text-gray-300"
              (click)="closed.emit()"
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                <path fillRule="evenodd" clipRule="evenodd" d="M6.04289 16.5413C5.65237 16.9318 5.65237 17.565 6.04289 17.9555C6.43342 18.346 7.06658 18.346 7.45711 17.9555L11.9987 13.4139L16.5408 17.956C16.9313 18.3466 17.5645 18.3466 17.955 17.956C18.3455 17.5655 18.3455 16.9323 17.955 16.5418L13.4129 11.9997L17.955 7.4576C18.3455 7.06707 18.3455 6.43391 17.955 6.04338C17.5645 5.65286 16.9313 5.65286 16.5408 6.04338L11.9987 10.5855L7.45711 6.0439C7.06658 5.65338 6.43342 5.65338 6.04289 6.0439C5.65237 6.43442 5.65237 7.06759 6.04289 7.45811L10.5845 11.9997L6.04289 16.5413Z" fill="currentColor"/>
              </svg>
            </button>
          </div>
          <div class="flex-1 overflow-y-auto p-5">
            <ng-content></ng-content>
          </div>
        </div>
      </div>
    }
  `,
})
export class DrawerComponent implements OnChanges {
  @Input() isOpen = false;
  @Input() title = '';
  @Input() position: 'right' | 'left' = 'right';
  @Input() size: 'sm' | 'md' | 'lg' | 'xl' = 'md';
  @Input() ariaLabel = '';
  @Input() closeOnBackdrop = true;
  @Input() closeOnEscape = true;

  @Output() closed = new EventEmitter<void>();

  private previousFocus: Element | null = null;

  ngOnChanges() {
    if (this.isOpen) {
      document.body.style.overflow = 'hidden';
      this.previousFocus = document.activeElement;
    } else {
      document.body.style.overflow = 'unset';
      if (this.previousFocus && typeof (this.previousFocus as HTMLElement).focus === 'function') {
        (this.previousFocus as HTMLElement).focus();
        this.previousFocus = null;
      }
    }
  }

  onBackdropClick() {
    if (this.closeOnBackdrop) {
      this.closed.emit();
    }
  }

  @HostListener('document:keydown.escape')
  onEscape() {
    if (this.isOpen && this.closeOnEscape) {
      this.closed.emit();
    }
  }
}
