import { DaySelectorComponent } from './day-selector.component';

describe('DaySelectorComponent', () => {
  let component: DaySelectorComponent;

  beforeEach(() => {
    component = new DaySelectorComponent();
  });

  it('renders all 7 day entries', () => {
    expect(component.days.length).toBe(7);
    expect(component.days[0].label).toBe('Mon');
    expect(component.days[6].label).toBe('Sun');
  });

  it('starts with no days selected', () => {
    expect(component.selectedDays).toEqual([]);
    expect(component.allSelected).toBe(false);
  });

  it('selects a day on toggle', () => {
    component.toggleDay(0, true);
    expect(component.selectedDays).toEqual([0]);
  });

  it('deselects a day on toggle', () => {
    component.selectedDays = [0, 2, 4];
    component.toggleDay(2, false);
    expect(component.selectedDays).toEqual([0, 4]);
  });

  it('keeps days sorted', () => {
    component.toggleDay(5, true);
    component.toggleDay(1, true);
    component.toggleDay(3, true);
    expect(component.selectedDays).toEqual([1, 3, 5]);
  });

  it('select All toggles all days on', () => {
    component.toggleAll(true);
    expect(component.selectedDays).toEqual([0, 1, 2, 3, 4, 5, 6]);
    expect(component.allSelected).toBe(true);
  });

  it('select All toggles all days off', () => {
    component.selectedDays = [0, 1, 2, 3, 4, 5, 6];
    component.toggleAll(false);
    expect(component.selectedDays).toEqual([]);
    expect(component.allSelected).toBe(false);
  });

  it('isSelected returns correct state', () => {
    component.selectedDays = [1, 3];
    expect(component.isSelected(0)).toBe(false);
    expect(component.isSelected(1)).toBe(true);
    expect(component.isSelected(3)).toBe(true);
  });

  it('writeValue sets selected days', () => {
    component.writeValue([2, 4, 6]);
    expect(component.selectedDays).toEqual([2, 4, 6]);
  });

  it('writeValue with null clears days', () => {
    component.selectedDays = [1, 2];
    component.writeValue(null);
    expect(component.selectedDays).toEqual([]);
  });

  it('registerOnChange fires on toggle', () => {
    let emitted: number[] = [];
    component.registerOnChange((val) => { emitted = val; });
    component.toggleDay(3, true);
    expect(emitted).toEqual([3]);
  });

  it('registerOnTouched fires on toggle', () => {
    let touched = false;
    component.registerOnTouched(() => { touched = true; });
    component.toggleDay(1, true);
    expect(touched).toBe(true);
  });

  it('does not duplicate a day if already selected', () => {
    component.selectedDays = [2];
    component.toggleDay(2, true);
    expect(component.selectedDays).toEqual([2]);
  });

  it('valueChange emits on toggle', () => {
    let emitted: number[] = [];
    component.valueChange.subscribe((val) => { emitted = val; });
    component.toggleDay(4, true);
    expect(emitted).toEqual([4]);
  });
});
