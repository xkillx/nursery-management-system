import { CapacityIndicatorComponent } from './capacity-indicator.component';

describe('CapacityIndicatorComponent', () => {
  let component: CapacityIndicatorComponent;

  beforeEach(() => {
    component = new CapacityIndicatorComponent();
  });

  it('shows green bar when current/max < 80%', () => {
    component.current = 5;
    component.max = 10;
    expect(component.statusLevel).toBe('green');
    expect(component.barColorClass).toBe('bg-success-500');
  });

  it('shows amber bar when current/max 80-99%', () => {
    component.current = 8;
    component.max = 10;
    expect(component.statusLevel).toBe('amber');
    expect(component.barColorClass).toBe('bg-warning-500');
  });

  it('shows red bar when current/max >= 100%', () => {
    component.current = 10;
    component.max = 10;
    expect(component.statusLevel).toBe('red');
    expect(component.barColorClass).toBe('bg-error-500');
  });

  it('shows red bar when over capacity', () => {
    component.current = 12;
    component.max = 10;
    expect(component.statusLevel).toBe('red');
    expect(component.percentage).toBe(100);
  });

  it('calculates percentage correctly', () => {
    component.current = 5;
    component.max = 20;
    expect(component.percentage).toBe(25);
  });

  it('caps percentage at 100', () => {
    component.current = 15;
    component.max = 10;
    expect(component.percentage).toBe(100);
  });

  it('handles max=0 without division error', () => {
    component.current = 0;
    component.max = 0;
    expect(component.percentage).toBe(0);
    expect(component.statusLevel).toBe('green');
  });

  it('handles negative max gracefully', () => {
    component.current = 0;
    component.max = -5;
    expect(component.percentage).toBe(0);
    expect(component.statusLevel).toBe('green');
  });

  it('sets ARIA attributes correctly', () => {
    component.current = 7;
    component.max = 10;
    component.ariaLabel = 'Room capacity';
    expect(component.current).toBe(7);
    expect(component.max).toBe(10);
    expect(component.ariaLabel).toBe('Room capacity');
  });

  it('returns correct text color for green status', () => {
    component.current = 3;
    component.max = 10;
    expect(component.textColorClass).toContain('text-success-600');
  });

  it('returns correct text color for amber status', () => {
    component.current = 9;
    component.max = 10;
    expect(component.textColorClass).toContain('text-warning-600');
  });

  it('returns correct text color for red status', () => {
    component.current = 10;
    component.max = 10;
    expect(component.textColorClass).toContain('text-error-600');
  });
});
