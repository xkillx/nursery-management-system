import { BookingWizardDrawerComponent, BookingData } from './booking-wizard-drawer.component';

describe('BookingWizardDrawerComponent', () => {
  let component: BookingWizardDrawerComponent;

  beforeEach(() => {
    component = new BookingWizardDrawerComponent();
  });

  it('initializes with 3 steps', () => {
    expect(component.steps.length).toBe(3);
    expect(component.steps[0].key).toBe('booking-type');
    expect(component.steps[1].key).toBe('schedule');
    expect(component.steps[2].key).toBe('review');
  });

  it('starts at step 0', () => {
    expect(component.currentStepIndex).toBe(0);
    expect(component.isFirstStep).toBe(true);
    expect(component.isLastStep).toBe(false);
  });

  it('cannot proceed on step 0 without booking type', () => {
    component.bookingType = '';
    expect(component.canProceed).toBe(false);
  });

  it('can proceed on step 0 with booking type selected', () => {
    component.bookingType = 'recurring';
    expect(component.canProceed).toBe(true);
  });

  it('advances to next step', () => {
    component.bookingType = 'recurring';
    component.nextStep();
    expect(component.currentStepIndex).toBe(1);
    expect(component.isFirstStep).toBe(false);
  });

  it('goes back to previous step', () => {
    component.currentStepIndex = 1;
    component.prevStep();
    expect(component.currentStepIndex).toBe(0);
  });

  it('does not go before first step', () => {
    component.currentStepIndex = 0;
    component.prevStep();
    expect(component.currentStepIndex).toBe(0);
  });

  it('does not go past last step', () => {
    component.currentStepIndex = 2;
    component.nextStep();
    expect(component.currentStepIndex).toBe(2);
  });

  it('stepIsActive returns correct state', () => {
    component.currentStepIndex = 1;
    expect(component.stepIsActive('booking-type')).toBe(false);
    expect(component.stepIsActive('schedule')).toBe(true);
    expect(component.stepIsActive('review')).toBe(false);
  });

  it('stepIsComplete returns correct state', () => {
    component.currentStepIndex = 2;
    expect(component.stepIsComplete(0)).toBe(true);
    expect(component.stepIsComplete(1)).toBe(true);
    expect(component.stepIsComplete(2)).toBe(false);
  });

  describe('validation on step 1 (schedule)', () => {
    beforeEach(() => {
      component.currentStepIndex = 1;
    });

    it('recurring requires days and dates', () => {
      component.bookingType = 'recurring';
      component.daysOfWeek = [];
      component.startDate = '';
      component.endDate = '';
      expect(component.canProceed).toBe(false);

      component.daysOfWeek = [0, 2, 4];
      component.startDate = '2026-07-20';
      component.endDate = '2026-12-18';
      expect(component.canProceed).toBe(true);
    });

    it('adhoc requires specific date', () => {
      component.bookingType = 'adhoc';
      component.specificDate = '';
      expect(component.canProceed).toBe(false);

      component.specificDate = '2026-07-25';
      expect(component.canProceed).toBe(true);
    });

    it('funded requires start date and allocation', () => {
      component.bookingType = 'funded';
      component.startDate = '';
      component.fundingAllocation = '';
      expect(component.canProceed).toBe(false);

      component.startDate = '2026-07-20';
      component.fundingAllocation = '15hr';
      expect(component.canProceed).toBe(true);
    });
  });

  it('submit emits completed event with booking data', () => {
    let emitted: BookingData | null = null;
    component.completed.subscribe((data) => { emitted = data; });

    component.bookingType = 'recurring';
    component.daysOfWeek = [0, 2, 4];
    component.startDate = '2026-07-20';
    component.endDate = '2026-12-18';
    component.sessionTemplate = 'full_day';
    component.room = 'room_1';

    component.submit();

    expect(emitted).toBeTruthy();
    expect(emitted!.bookingType).toBe('recurring');
    expect(emitted!.daysOfWeek).toEqual([0, 2, 4]);
    expect(emitted!.startDate).toBe('2026-07-20');
    expect(emitted!.endDate).toBe('2026-12-18');
  });

  it('submit resets the form', () => {
    component.bookingType = 'recurring';
    component.daysOfWeek = [0, 2];
    component.submit();

    expect(component.currentStepIndex).toBe(0);
    expect(component.bookingType).toBe('');
    expect(component.daysOfWeek).toEqual([]);
  });

  it('onClose resets and emits closed', () => {
    let closed = false;
    component.closed.subscribe(() => { closed = true; });
    component.bookingType = 'adhoc';
    component.onClose();

    expect(closed).toBe(true);
    expect(component.bookingType).toBe('');
  });

  it('capacityStatusLevel returns green for low capacity', () => {
    component.currentCapacity = 5;
    component.maxCapacity = 10;
    expect(component.capacityStatusLevel).toBe('green');
  });

  it('capacityStatusLevel returns amber for 80-99%', () => {
    component.currentCapacity = 9;
    component.maxCapacity = 10;
    expect(component.capacityStatusLevel).toBe('amber');
  });

  it('capacityStatusLevel returns red for 100%+', () => {
    component.currentCapacity = 10;
    component.maxCapacity = 10;
    expect(component.capacityStatusLevel).toBe('red');
  });
});
