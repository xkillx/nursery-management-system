import { Component, Input } from '@angular/core';
import { BadgeComponent } from './badge.component';

type StatusColor = 'primary' | 'success' | 'error' | 'warning' | 'info' | 'light' | 'dark';

interface StatusMapping {
  color: StatusColor;
  label: string;
}

@Component({
  selector: 'app-status-badge',
  imports: [BadgeComponent],
  template: `
    <app-badge
      [variant]="variant"
      [size]="size"
      [color]="resolvedMapping.color"
    >
      {{ resolvedMapping.label }}
    </app-badge>
  `,
})
export class StatusBadgeComponent {
  @Input() status: string | null | undefined = null;
  @Input() label?: string;
  @Input() size: 'sm' | 'md' = 'sm';
  @Input() variant: 'light' | 'solid' = 'light';

  private static readonly STATUS_MAP: Record<string, StatusMapping> = {
    active: { color: 'success', label: 'Active' },
    inactive: { color: 'light', label: 'Inactive' },
    complete: { color: 'success', label: 'Complete' },
    completed: { color: 'success', label: 'Complete' },
    incomplete: { color: 'warning', label: 'Incomplete' },
    checked_in: { color: 'success', label: 'Checked in' },
    not_checked_in: { color: 'light', label: 'Not in' },
    absent: { color: 'warning', label: 'Absent' },
    draft: { color: 'info', label: 'Draft' },
    issued: { color: 'primary', label: 'Issued' },
    paid: { color: 'success', label: 'Paid' },
    overdue: { color: 'warning', label: 'Overdue' },
    payment_failed: { color: 'error', label: 'Payment failed' },
    not_due: { color: 'light', label: 'Not due' },
    due: { color: 'warning', label: 'Due' },
    payable: { color: 'success', label: 'Payable' },
    not_payable: { color: 'light', label: 'Not payable' },
    pending: { color: 'warning', label: 'Pending' },
    accepted: { color: 'success', label: 'Accepted' },
    revoked: { color: 'error', label: 'Revoked' },
    expired: { color: 'light', label: 'Expired' },
    unpaid: { color: 'warning', label: 'Unpaid' },
    awaiting_provider_update: { color: 'info', label: 'Awaiting provider update' },
    not_issued: { color: 'light', label: 'Not issued' },
    no_payment_due: { color: 'light', label: 'No payment due' },
    room_assigned: { color: 'info', label: 'Room assigned' },
  };

  get resolvedMapping(): StatusMapping {
    const key = (this.status ?? '').toLowerCase().trim();
    const known = StatusBadgeComponent.STATUS_MAP[key];
    if (known) {
      return {
        color: known.color,
        label: this.label ?? known.label,
      };
    }
    return {
      color: 'light',
      label: this.label ?? StatusBadgeComponent.titleCase(key),
    };
  }

  private static titleCase(value: string): string {
    return value
      .replace(/[-_]+/g, ' ')
      .replace(/\b\w/g, (c) => c.toUpperCase())
      .trim();
  }
}
