import { CommonModule } from '@angular/common';
import {
  Component,
  ElementRef,
  EventEmitter,
  HostListener,
  Input,
  Output, OnInit, OnDestroy, OnChanges, inject,
} from '@angular/core';

@Component({
  selector: 'app-modal',
  imports: [
    CommonModule,
  ],
  templateUrl: './modal.component.html',
  styles: ``,
})
export class ModalComponent implements OnInit, OnDestroy, OnChanges {

  @Input() isOpen = false;
  @Output() closed = new EventEmitter<void>();
  @Input() className = '';
  @Input() showCloseButton = true;
  @Input() isFullscreen = false;
  @Input() ariaLabel = '';
  @Input() ariaLabelledBy = '';
  @Input() closeOnBackdrop = true;
  @Input() closeOnEscape = true;
  @Input() initialFocusSelector = '';

  private static openCount = 0;
  private previousFocus: Element | null = null;

  private readonly el = inject(ElementRef);

  ngOnInit() {
    if (this.isOpen) {
      this.onOpen();
    }
  }

  ngOnDestroy() {
    this.onCloseCleanup();
  }

  ngOnChanges() {
    if (this.isOpen) {
      this.onOpen();
    } else {
      this.onCloseCleanup();
    }
  }

  onBackdropClick() {
    if (!this.isFullscreen && this.closeOnBackdrop) {
      this.closed.emit();
    }
  }

  onContentClick(event: Event) {
    event.stopPropagation();
  }

  @HostListener('document:keydown.escape')
  onEscape() {
    if (this.isOpen && this.closeOnEscape) {
      this.closed.emit();
    }
  }

  private onOpen() {
    ModalComponent.openCount++;
    document.body.style.overflow = 'hidden';
    this.previousFocus = document.activeElement;

    if (this.initialFocusSelector) {
      setTimeout(() => {
        const target = this.el.nativeElement.querySelector(this.initialFocusSelector);
        if (target) (target as HTMLElement).focus();
      });
    }
  }

  private onCloseCleanup() {
    if (ModalComponent.openCount > 0) {
      ModalComponent.openCount--;
    }
    if (ModalComponent.openCount === 0) {
      document.body.style.overflow = 'unset';
    }
    if (this.previousFocus && typeof (this.previousFocus as HTMLElement).focus === 'function') {
      (this.previousFocus as HTMLElement).focus();
      this.previousFocus = null;
    }
  }
}
