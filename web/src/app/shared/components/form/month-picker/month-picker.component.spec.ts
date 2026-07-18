import { MonthPickerComponent } from './month-picker.component';

describe('MonthPickerComponent', () => {
  let component: MonthPickerComponent;

  beforeEach(() => {
    component = new MonthPickerComponent();
  });

  it('initializes with current month and year', () => {
    const now = new Date();
    expect(component.value.month).toBe(now.getMonth());
    expect(component.value.year).toBe(now.getFullYear());
  });

  it('returns correct month name', () => {
    component.value = { month: 0, year: 2026 };
    expect(component.monthName).toBe('January');

    component.value = { month: 11, year: 2026 };
    expect(component.monthName).toBe('December');
  });

  it('left arrow decrements month', () => {
    component.value = { month: 5, year: 2026 };
    component.prevMonth();
    expect(component.value.month).toBe(4);
    expect(component.value.year).toBe(2026);
  });

  it('left arrow wraps from Jan to Dec of previous year', () => {
    component.value = { month: 0, year: 2026 };
    component.prevMonth();
    expect(component.value.month).toBe(11);
    expect(component.value.year).toBe(2025);
  });

  it('right arrow increments month', () => {
    component.value = { month: 5, year: 2026 };
    component.nextMonth();
    expect(component.value.month).toBe(6);
    expect(component.value.year).toBe(2026);
  });

  it('right arrow wraps from Dec to Jan of next year', () => {
    component.value = { month: 11, year: 2026 };
    component.nextMonth();
    expect(component.value.month).toBe(0);
    expect(component.value.year).toBe(2027);
  });

  it('year dropdown changes year, preserves month', () => {
    component.value = { month: 3, year: 2026 };
    const event = { target: { value: '2028' } } as unknown as Event;
    component.onYearChange(event);
    expect(component.value.year).toBe(2028);
    expect(component.value.month).toBe(3);
  });

  it('writeValue sets the value', () => {
    component.writeValue({ month: 7, year: 2025 });
    expect(component.value.month).toBe(7);
    expect(component.value.year).toBe(2025);
  });

  it('writeValue with null keeps default', () => {
    const original = { ...component.value };
    component.writeValue(null);
    expect(component.value.month).toBe(original.month);
    expect(component.value.year).toBe(original.year);
  });

  it('registerOnChange fires on prevMonth', () => {
    let emitted: unknown = null;
    component.registerOnChange((val) => { emitted = val; });
    component.value = { month: 5, year: 2026 };
    component.prevMonth();
    expect(emitted).toEqual({ month: 4, year: 2026 });
  });

  it('registerOnChange fires on nextMonth', () => {
    let emitted: unknown = null;
    component.registerOnChange((val) => { emitted = val; });
    component.value = { month: 5, year: 2026 };
    component.nextMonth();
    expect(emitted).toEqual({ month: 6, year: 2026 });
  });

  it('valueChange emits on navigation', () => {
    let emitted: unknown = null;
    component.valueChange.subscribe((val) => { emitted = val; });
    component.value = { month: 5, year: 2026 };
    component.nextMonth();
    expect(emitted).toEqual({ month: 6, year: 2026 });
  });

  it('yearOptions includes range around current year', () => {
    const currentYear = new Date().getFullYear();
    expect(component.yearOptions).toContain(currentYear - 5);
    expect(component.yearOptions).toContain(currentYear + 5);
    expect(component.yearOptions).toContain(currentYear);
    expect(component.yearOptions.length).toBe(11);
  });
});
