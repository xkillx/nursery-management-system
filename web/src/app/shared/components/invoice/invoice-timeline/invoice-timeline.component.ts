import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroCheckBadge,
  heroClock,
  heroExclamationCircle,
  heroExclamationTriangle,
  heroLockClosed,
  heroReceiptPercent,
  heroShieldCheck,
} from '@ng-icons/heroicons/outline';

export type TimelineTone = 'success' | 'primary' | 'warning' | 'error' | 'neutral';

export interface TimelineEntry {
  key: string;
  icon: string;
  tone: TimelineTone;
  title: string;
  description: string;
  timestamp: string | null;
  isPending?: boolean;
}

const TONE_CLASSES: Record<TimelineTone, string> = {
  success: 'bg-success-500 text-white',
  primary: 'bg-brand-500 text-white',
  warning: 'bg-warning-500 text-white',
  error: 'bg-error-500 text-white',
  neutral: 'bg-gray-200 text-gray-700 dark:bg-gray-700 dark:text-gray-200',
};

const PENDING_TONE_CLASSES: Record<TimelineTone, string> = {
  success: 'border-success-500 text-success-500',
  primary: 'border-brand-500 text-brand-500',
  warning: 'border-warning-500 text-warning-500',
  error: 'border-error-500 text-error-500',
  neutral: 'border-gray-300 text-gray-400 dark:border-gray-600 dark:text-gray-500',
};

const CONNECTOR_CLASSES: Record<TimelineTone, string> = {
  success: 'bg-success-500',
  primary: 'bg-brand-500',
  warning: 'bg-warning-500',
  error: 'bg-error-500',
  neutral: 'bg-gray-200 dark:bg-gray-700',
};

@Component({
  selector: 'app-invoice-timeline',
  imports: [CommonModule, NgIcon],
  providers: [
    provideIcons({
      heroCheckBadge,
      heroClock,
      heroExclamationCircle,
      heroExclamationTriangle,
      heroLockClosed,
      heroReceiptPercent,
      heroShieldCheck,
    }),
  ],
  templateUrl: './invoice-timeline.component.html',
})
export class InvoiceTimelineComponent {
  entries = input.required<TimelineEntry[]>();

  nodeClasses(entry: TimelineEntry): string {
    if (entry.isPending) {
      return `border-2 border-dashed ${PENDING_TONE_CLASSES[entry.tone]}`;
    }
    return TONE_CLASSES[entry.tone];
  }

  connectorClasses(entry: TimelineEntry): string {
    return CONNECTOR_CLASSES[entry.tone];
  }

  formatTimestamp(timestamp: string | null): string {
    if (!timestamp) return '';
    const d = new Date(timestamp);
    return new Intl.DateTimeFormat('en-GB', {
      timeZone: 'Europe/London',
      dateStyle: 'medium',
      timeStyle: 'short',
    }).format(d);
  }
}
