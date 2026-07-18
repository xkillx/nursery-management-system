import { StatusDropdownComponent } from './status-dropdown.component';

describe('StatusDropdownComponent', () => {
  let component: StatusDropdownComponent;

  beforeEach(() => {
    component = new StatusDropdownComponent();
  });

  it('initializes with no selected status', () => {
    expect(component.selectedValue).toBeNull();
    expect(component.selectedStatus).toBeUndefined();
    expect(component.isOpen).toBe(false);
  });

  it('has all 6 attendance statuses', () => {
    expect(component.statuses.length).toBe(6);
    const values = component.statuses.map((s) => s.value);
    expect(values).toContain('booked');
    expect(values).toContain('present');
    expect(values).toContain('absent');
    expect(values).toContain('late_arrival');
    expect(values).toContain('early_pickup');
    expect(values).toContain('no_show');
  });

  it('each status has distinct icon and color', () => {
    const icons = component.statuses.map((s) => s.icon);
    const uniqueIcons = new Set(icons);
    expect(uniqueIcons.size).toBe(6);

    const colors = component.statuses.map((s) => s.colorClass);
    const uniqueColors = new Set(colors);
    expect(uniqueColors.size).toBe(6);
  });

  it('toggle opens and closes', () => {
    component.toggle();
    expect(component.isOpen).toBe(true);
    component.toggle();
    expect(component.isOpen).toBe(false);
  });

  it('select sets value and closes', () => {
    component.select('present');
    expect(component.selectedValue).toBe('present');
    expect(component.isOpen).toBe(false);
  });

  it('select returns correct status mapping', () => {
    component.select('absent');
    expect(component.selectedStatus?.label).toBe('Absent');
    expect(component.selectedStatus?.icon).toBe('✗');
  });

  it('close sets isOpen to false', () => {
    component.isOpen = true;
    component.close();
    expect(component.isOpen).toBe(false);
  });

  it('writeValue sets the selected value', () => {
    component.writeValue('late_arrival');
    expect(component.selectedValue).toBe('late_arrival');
    expect(component.selectedStatus?.label).toBe('Late Arrival');
  });

  it('registerOnChange fires on select', () => {
    let emitted: unknown = null;
    component.registerOnChange((val) => { emitted = val; });
    component.select('present');
    expect(emitted).toBe('present');
  });

  it('registerOnTouched fires on close', () => {
    let touched = false;
    component.registerOnTouched(() => { touched = true; });
    component.close();
    expect(touched).toBe(true);
  });

  it('valueChange emits on select', () => {
    let emitted: unknown = null;
    component.valueChange.subscribe((val) => { emitted = val; });
    component.select('no_show');
    expect(emitted).toBe('no_show');
  });

  it('writeValue with null clears selection', () => {
    component.selectedValue = 'present';
    component.writeValue(null);
    expect(component.selectedValue).toBeNull();
    expect(component.selectedStatus).toBeUndefined();
  });
});
