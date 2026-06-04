import { Injectable } from '@angular/core';

export type ToastVariant = 'success' | 'error' | 'warning' | 'info';

export interface Toast {
  id: string;
  variant: ToastVariant;
  message: string;
  title?: string;
  durationMs?: number;
  createdAt: number;
}

export interface ToastOptions {
  title?: string;
  durationMs?: number;
}

@Injectable({ providedIn: 'root' })
export class ToastService {
  private static MAX_VISIBLE = 5;
  private static DEFAULT_DURATION = 5000;
  private static counter = 0;

  private toasts: Toast[] = [];
  private timers = new Map<string, ReturnType<typeof setTimeout>>();

  readonly state = {
    toasts: () => this.toasts,
  };

  success(message: string, options?: ToastOptions): string {
    return this.add('success', message, options);
  }

  error(message: string, options?: ToastOptions): string {
    return this.add('error', message, options);
  }

  warning(message: string, options?: ToastOptions): string {
    return this.add('warning', message, options);
  }

  info(message: string, options?: ToastOptions): string {
    return this.add('info', message, options);
  }

  dismiss(id: string): void {
    const timer = this.timers.get(id);
    if (timer) {
      clearTimeout(timer);
      this.timers.delete(id);
    }
    this.toasts = this.toasts.filter((t) => t.id !== id);
  }

  clear(): void {
    this.timers.forEach((timer) => clearTimeout(timer));
    this.timers.clear();
    this.toasts = [];
  }

  private add(variant: ToastVariant, message: string, options?: ToastOptions): string {
    const id = `toast-${++ToastService.counter}`;
    const durationMs = options?.durationMs ?? ToastService.DEFAULT_DURATION;

    const toast: Toast = {
      id,
      variant,
      message,
      title: options?.title,
      durationMs,
      createdAt: Date.now(),
    };

    this.toasts = [...this.toasts, toast];

    if (this.toasts.length > ToastService.MAX_VISIBLE) {
      const removed = this.toasts.shift();
      if (removed) {
        const timer = this.timers.get(removed.id);
        if (timer) {
          clearTimeout(timer);
          this.timers.delete(removed.id);
        }
      }
      this.toasts = [...this.toasts];
    }

    this.scheduleDismiss(id, durationMs);

    return id;
  }

  private scheduleDismiss(id: string, durationMs: number): void {
    const timer = setTimeout(() => {
      this.dismiss(id);
    }, durationMs);
    this.timers.set(id, timer);
  }
}
