import { BookingCalendarComponent, BookingCalendarEvent } from './booking-calendar.component';

describe('BookingCalendarComponent', () => {
  let component: BookingCalendarComponent;

  beforeEach(() => {
    component = new BookingCalendarComponent();
  });

  it('initializes with default inputs', () => {
    expect(component.bookings).toEqual([]);
    expect(component.view).toBe('month');
  });

  it('emits dateSelect with start and end', () => {
    let emitted: unknown = null;
    component.dateSelect.subscribe((val) => { emitted = val; });

    component.dateSelect.emit({ start: '2026-07-18', end: '2026-07-19' });

    expect(emitted).toEqual({ start: '2026-07-18', end: '2026-07-19' });
  });

  it('emits bookingClick with the booking', () => {
    const booking: BookingCalendarEvent = {
      id: '1',
      title: 'Test Booking',
      start: '2026-07-18',
      type: 'regular',
    };
    let emitted: unknown = null;
    component.bookingClick.subscribe((val) => { emitted = val; });

    component.bookings = [booking];
    component.bookingClick.emit(booking);

    expect(emitted).toEqual(booking);
  });

  it('determines responsive view correctly for month', () => {
    component.view = 'month';
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (component as any).currentWidth = 1400;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    expect((component as any).getResponsiveView()).toBe('dayGridMonth');

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (component as any).currentWidth = 600;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    expect((component as any).getResponsiveView()).toBe('dayGridMonth');
  });

  it('determines responsive view correctly for week', () => {
    component.view = 'week';
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (component as any).currentWidth = 1400;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    expect((component as any).getResponsiveView()).toBe('timeGridWeek');

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (component as any).currentWidth = 900;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    expect((component as any).getResponsiveView()).toBe('timeGridWeek');

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (component as any).currentWidth = 500;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    expect((component as any).getResponsiveView()).toBe('timeGridDay');
  });

  it('shouldUpdateView detects breakpoint crossing', () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    expect((component as any).shouldUpdateView(600, 800)).toBe(true);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    expect((component as any).shouldUpdateView(800, 900)).toBe(false);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    expect((component as any).shouldUpdateView(1200, 1300)).toBe(true);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    expect((component as any).shouldUpdateView(1300, 1400)).toBe(false);
  });

  it('renderEventContent returns HTML with correct color class', () => {
    const result = component.renderEventContent({
      event: {
        title: 'Test',
        extendedProps: { type: 'funded' },
      },
      timeText: '10:00',
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } as any);
    expect(result.html).toContain('fc-bg-funded');
    expect(result.html).toContain('Test');
  });

  it('renderEventContent defaults to regular color for unknown type', () => {
    const result = component.renderEventContent({
      event: {
        title: 'Unknown',
        extendedProps: { type: 'unknown' },
      },
      timeText: '',
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } as any);
    expect(result.html).toContain('fc-bg-regular');
  });
});
