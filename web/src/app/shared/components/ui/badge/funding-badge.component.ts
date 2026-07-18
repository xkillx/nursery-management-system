import { Component, Input } from '@angular/core';
import { BadgeComponent } from './badge.component';

type BadgeColor = 'primary' | 'success' | 'error' | 'warning' | 'info' | 'light' | 'dark';

interface FundingMapping {
  color: BadgeColor;
  label: string;
}

@Component({
  selector: 'app-funding-badge',
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
export class FundingBadgeComponent {
  @Input() fundingType: string | null | undefined = null;
  @Input() label?: string;
  @Input() size: 'sm' | 'md' = 'sm';
  @Input() variant: 'light' | 'solid' = 'light';

  private static readonly FUNDING_MAP: Record<string, FundingMapping> = {
    private: { color: 'light', label: 'Private' },
    '15hr': { color: 'success', label: '15 Hours' },
    '30hr': { color: 'primary', label: '30 Hours' },
    mixed: { color: 'warning', label: 'Mixed' },
  };

  get resolvedMapping(): FundingMapping {
    const key = (this.fundingType ?? '').toLowerCase().trim();
    const known = FundingBadgeComponent.FUNDING_MAP[key];
    if (known) {
      return {
        color: known.color,
        label: this.label ?? known.label,
      };
    }
    return {
      color: 'light',
      label: this.label ?? FundingBadgeComponent.titleCase(key),
    };
  }

  private static titleCase(value: string): string {
    return value
      .replace(/[-_]+/g, ' ')
      .replace(/\b\w/g, (c) => c.toUpperCase())
      .trim();
  }
}
