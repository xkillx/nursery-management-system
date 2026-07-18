import { AttendanceSummaryBarComponent } from './attendance-summary-bar.component';

describe('AttendanceSummaryBarComponent', () => {
  let component: AttendanceSummaryBarComponent;

  beforeEach(() => {
    component = new AttendanceSummaryBarComponent();
  });

  it('initializes with default zero counts', () => {
    expect(component.total).toBe(0);
    expect(component.present).toBe(0);
    expect(component.absent).toBe(0);
    expect(component.late).toBe(0);
  });

  it('accepts all four count inputs', () => {
    component.total = 25;
    component.present = 20;
    component.absent = 3;
    component.late = 2;
    expect(component.total).toBe(25);
    expect(component.present).toBe(20);
    expect(component.absent).toBe(3);
    expect(component.late).toBe(2);
  });

  it('handles zero counts correctly', () => {
    component.total = 0;
    component.present = 0;
    component.absent = 0;
    component.late = 0;
    expect(component.total).toBe(0);
    expect(component.present).toBe(0);
    expect(component.absent).toBe(0);
    expect(component.late).toBe(0);
  });
});
