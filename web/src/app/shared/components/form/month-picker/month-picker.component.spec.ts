import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MonthPickerComponent } from './month-picker.component';

describe('MonthPickerComponent', () => {
  let component: MonthPickerComponent;
  let fixture: ComponentFixture<MonthPickerComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MonthPickerComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(MonthPickerComponent);
    component = fixture.componentInstance;
  });

  it('initializes with null value', () => {
    expect(component.value).toBeNull();
  });

  it('returns empty month name when value is null', () => {
    expect(component.monthName).toBe('');
  });

  it('returns correct month name after writeValue', () => {
    component.writeValue({ month: 0, year: 2026 });
    expect(component.monthName).toBe('January');

    component.writeValue({ month: 11, year: 2026 });
    expect(component.monthName).toBe('December');
  });

  it('left arrow decrements month', () => {
    component.value = { month: 5, year: 2026 };
    component.prevMonth();
    expect(component.value!.month).toBe(4);
    expect(component.value!.year).toBe(2026);
  });

  it('left arrow wraps from Jan to Dec of previous year', () => {
    component.value = { month: 0, year: 2026 };
    component.prevMonth();
    expect(component.value!.month).toBe(11);
    expect(component.value!.year).toBe(2025);
  });

  it('right arrow increments month', () => {
    component.value = { month: 5, year: 2026 };
    component.nextMonth();
    expect(component.value!.month).toBe(6);
    expect(component.value!.year).toBe(2026);
  });

  it('right arrow wraps from Dec to Jan of next year', () => {
    component.value = { month: 11, year: 2026 };
    component.nextMonth();
    expect(component.value!.month).toBe(0);
    expect(component.value!.year).toBe(2027);
  });

  it('year dropdown changes year, preserves month', () => {
    component.value = { month: 3, year: 2026 };
    const event = { target: { value: '2028' } } as unknown as Event;
    component.onYearChange(event);
    expect(component.value!.year).toBe(2028);
    expect(component.value!.month).toBe(3);
  });

  it('writeValue sets the value with MonthYear object', () => {
    component.writeValue({ month: 7, year: 2025 });
    expect(component.value!.month).toBe(7);
    expect(component.value!.year).toBe(2025);
  });

  it('writeValue parses YYYY-MM string', () => {
    component.writeValue('2025-03');
    expect(component.value!.month).toBe(2);
    expect(component.value!.year).toBe(2025);
  });

  it('writeValue with null keeps existing value', () => {
    component.value = { month: 5, year: 2026 };
    component.writeValue(null);
    expect(component.value!.month).toBe(5);
    expect(component.value!.year).toBe(2026);
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

  describe('dropdown mode', () => {
    beforeEach(() => {
      component.mode = 'dropdown';
      fixture.detectChanges();
    });

    it('selectMonth sets value and closes dropdown', () => {
      component.isOpen = true;
      component.dropdownYear = 2026;
      component.selectMonth(5);
      expect(component.value).toEqual({ month: 5, year: 2026 });
      expect(component.isOpen).toBeFalse();
    });

    it('prevDropdownYear decrements year', () => {
      component.dropdownYear = 2026;
      component.prevDropdownYear();
      expect(component.dropdownYear).toBe(2025);
    });

    it('nextDropdownYear increments year', () => {
      component.dropdownYear = 2026;
      component.nextDropdownYear();
      expect(component.dropdownYear).toBe(2027);
    });

    it('displayLabel returns formatted string', () => {
      component.value = { month: 6, year: 2026 };
      expect(component.displayLabel).toBe('July 2026');
    });

    it('displayLabel returns empty when value is null', () => {
      expect(component.displayLabel).toBe('');
    });

    it('toggleDropdown opens and sets dropdownYear from value', () => {
      component.value = { month: 3, year: 2025 };
      component.toggleDropdown();
      expect(component.isOpen).toBeTrue();
      expect(component.dropdownYear).toBe(2025);
    });

    it('toggleDropdown does nothing when disabled', () => {
      component.disabled = true;
      component.toggleDropdown();
      expect(component.isOpen).toBeFalse();
    });
  });
});
