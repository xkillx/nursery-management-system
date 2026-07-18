import { FundingBadgeComponent } from './funding-badge.component';

describe('FundingBadgeComponent', () => {
  let component: FundingBadgeComponent;

  beforeEach(() => {
    component = new FundingBadgeComponent();
  });

  it('maps private funding to light color', () => {
    component.fundingType = 'private';
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('light');
    expect(mapping.label).toBe('Private');
  });

  it('maps 15hr funding to success color', () => {
    component.fundingType = '15hr';
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('success');
    expect(mapping.label).toBe('15 Hours');
  });

  it('maps 30hr funding to primary color', () => {
    component.fundingType = '30hr';
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('primary');
    expect(mapping.label).toBe('30 Hours');
  });

  it('maps mixed funding to warning color', () => {
    component.fundingType = 'mixed';
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('warning');
    expect(mapping.label).toBe('Mixed');
  });

  it('falls back to unknown styling for unrecognized funding type', () => {
    component.fundingType = 'unknown_type';
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('light');
    expect(mapping.label).toBe('Unknown Type');
  });

  it('handles empty funding type gracefully', () => {
    component.fundingType = '';
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('light');
    expect(mapping.label).toBe('');
  });

  it('handles null funding type gracefully', () => {
    component.fundingType = null;
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('light');
    expect(mapping.label).toBe('');
  });

  it('handles undefined funding type gracefully', () => {
    component.fundingType = undefined;
    const mapping = component.resolvedMapping;
    expect(mapping.color).toBe('light');
    expect(mapping.label).toBe('');
  });

  it('uses custom label when provided', () => {
    component.fundingType = 'private';
    component.label = 'Self-funded';
    expect(component.resolvedMapping.label).toBe('Self-funded');
  });

  it('is case-insensitive', () => {
    component.fundingType = 'PRIVATE';
    expect(component.resolvedMapping.color).toBe('light');
    expect(component.resolvedMapping.label).toBe('Private');
  });

  it('handles 15hr with alternate casing', () => {
    component.fundingType = '15HR';
    expect(component.resolvedMapping.color).toBe('success');
  });
});
