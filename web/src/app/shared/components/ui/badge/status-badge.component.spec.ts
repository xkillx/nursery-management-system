import { StatusBadgeComponent } from './status-badge.component';

describe('StatusBadgeComponent', () => {
  let component: StatusBadgeComponent;

  beforeEach(() => {
    component = new StatusBadgeComponent();
  });

  it('maps active status to success color', () => {
    component.status = 'active';
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('success');
    expect(mapping.label).toBe('Active');
  });

  it('maps inactive status to light color', () => {
    component.status = 'inactive';
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('light');
    expect(mapping.label).toBe('Inactive');
  });

  it('maps checked_in status to success', () => {
    component.status = 'checked_in';
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('success');
    expect(mapping.label).toBe('Checked in');
  });

  it('maps not_checked_in status to light', () => {
    component.status = 'not_checked_in';
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('light');
    expect(mapping.label).toBe('Not in');
  });

  it('maps overdue status to warning', () => {
    component.status = 'overdue';
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('warning');
    expect(mapping.label).toBe('Overdue');
  });

  it('maps payment_failed to error', () => {
    component.status = 'payment_failed';
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('error');
    expect(mapping.label).toBe('Payment failed');
  });

  it('maps draft to info', () => {
    component.status = 'draft';
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('info');
    expect(mapping.label).toBe('Draft');
  });

  it('title-cases unknown statuses', () => {
    component.status = 'some_unknown_status';
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('light');
    expect(mapping.label).toBe('Some Unknown Status');
  });

  it('handles null status as neutral', () => {
    component.status = null;
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('light');
    expect(mapping.label).toBe('');
  });

  it('handles undefined status as neutral', () => {
    component.status = undefined;
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('light');
    expect(mapping.label).toBe('');
  });

  it('uses custom label when provided', () => {
    component.status = 'active';
    component.label = 'Custom Label';
    expect(component.resolvedMapping.label).toBe('Custom Label');
  });

  it('is case-insensitive', () => {
    component.status = 'ACTIVE';
    expect(component.resolvedMapping.color).toBe('success');
  });

  it('maps pending status to warning', () => {
    component.status = 'pending';
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('warning');
    expect(mapping.label).toBe('Pending');
  });

  it('maps accepted status to success', () => {
    component.status = 'accepted';
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('success');
    expect(mapping.label).toBe('Accepted');
  });

  it('maps revoked status to error', () => {
    component.status = 'revoked';
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('error');
    expect(mapping.label).toBe('Revoked');
  });

  it('maps expired status to light', () => {
    component.status = 'expired';
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('light');
    expect(mapping.label).toBe('Expired');
  });

  it('maps not_due status to light color with Not due label', () => {
    component.status = 'not_due';
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('light');
    expect(mapping.label).toBe('Not due');
  });

  it('maps due status to warning color with Due label', () => {
    component.status = 'due';
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('warning');
    expect(mapping.label).toBe('Due');
  });
});
