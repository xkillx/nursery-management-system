import { Component, EventEmitter, HostListener, Input, OnChanges, Output, SimpleChanges, ViewChild } from '@angular/core';
import { FullCalendarComponent, FullCalendarModule } from '@fullcalendar/angular';
import { CalendarOptions, DateSelectArg, EventClickArg, EventContentArg } from '@fullcalendar/core';
import dayGridPlugin from '@fullcalendar/daygrid';
import timeGridPlugin from '@fullcalendar/timegrid';
import interactionPlugin from '@fullcalendar/interaction';

export interface BookingCalendarEvent {
  id: string;
  title: string;
  start: string;
  end?: string;
  type: 'regular' | 'funded' | 'adhoc' | 'cancelled' | 'wraparound';
  extendedProps?: Record<string, unknown>;
}

export interface CalendarBackgroundEvent {
  start: string;
  end?: string;
  display: 'background';
  backgroundColor: string;
  classNames?: string[];
}

const BOOKING_TYPE_COLORS: Record<string, string> = {
  regular: 'fc-bg-regular',
  funded: 'fc-bg-funded',
  adhoc: 'fc-bg-adhoc',
  cancelled: 'fc-bg-cancelled',
  wraparound: 'fc-bg-wraparound',
};

export const CLOSURE_BACKGROUND_COLOR = 'rgba(239, 68, 68, 0.15)';
export const HOLIDAY_BACKGROUND_COLOR = 'rgba(245, 158, 11, 0.15)';

@Component({
  selector: 'app-booking-calendar',
  imports: [FullCalendarModule],
  templateUrl: './booking-calendar.component.html',
})
export class BookingCalendarComponent implements OnChanges {
  @ViewChild('calendar') calendarComponent!: FullCalendarComponent;

  @Input() bookings: BookingCalendarEvent[] = [];
  @Input() backgroundEvents: CalendarBackgroundEvent[] = [];
  @Input() view: 'week' | 'month' = 'month';

  @Output() dateSelect = new EventEmitter<{ start: string; end: string }>();
  @Output() bookingClick = new EventEmitter<BookingCalendarEvent>();
  @Output() datesSet = new EventEmitter<{ start: string; end: string }>();

  calendarOptions!: CalendarOptions;

  private currentWidth = typeof window !== 'undefined' ? window.innerWidth : 1280;

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['bookings'] || changes['view']) {
      this.updateCalendarOptions();
    }
  }

  @HostListener('window:resize')
  onResize(): void {
    const newWidth = window.innerWidth;
    if (this.shouldUpdateView(this.currentWidth, newWidth)) {
      this.currentWidth = newWidth;
      this.updateCalendarOptions();
    }
  }

  private shouldUpdateView(oldWidth: number, newWidth: number): boolean {
    const breakpoints = [768, 1280];
    return breakpoints.some(
      (bp) => (oldWidth < bp && newWidth >= bp) || (oldWidth >= bp && newWidth < bp)
    );
  }

  private updateCalendarOptions(): void {
    const initialView = this.getResponsiveView();

    this.calendarOptions = {
      plugins: [dayGridPlugin, timeGridPlugin, interactionPlugin],
      initialView,
      headerToolbar: {
        left: 'prev,next today',
        center: 'title',
        right: 'dayGridMonth,timeGridWeek,timeGridDay',
      },
      selectable: true,
      events: [
        ...this.bookings.map((b) => ({
          id: b.id,
          title: b.title,
          start: b.start,
          end: b.end,
          extendedProps: { type: b.type, ...b.extendedProps },
        })),
        ...this.backgroundEvents.map((bg) => ({
          start: bg.start,
          end: bg.end,
          display: 'background' as const,
          backgroundColor: bg.backgroundColor,
          classNames: bg.classNames,
        })),
      ],
      select: (info: DateSelectArg) => {
        this.dateSelect.emit({ start: info.startStr, end: info.endStr });
      },
      eventClick: (info: EventClickArg) => {
        const booking = this.bookings.find((b) => b.id === info.event.id);
        if (booking) {
          this.bookingClick.emit(booking);
        }
      },
      datesSet: (info: { start: Date; end: Date }) => {
        this.datesSet.emit({
          start: info.start.toISOString().split('T')[0],
          end: info.end.toISOString().split('T')[0],
        });
      },
      eventContent: (arg: EventContentArg) => this.renderEventContent(arg),
    };
  }

  private getResponsiveView(): string {
    if (this.view === 'week') {
      if (this.currentWidth < 768) return 'timeGridDay';
      if (this.currentWidth < 1280) return 'timeGridWeek';
      return 'timeGridWeek';
    }
    return 'dayGridMonth';
  }

  renderEventContent(eventInfo: EventContentArg): { html: string } {
    const type = eventInfo.event.extendedProps['type'] as string;
    const colorClass = BOOKING_TYPE_COLORS[type] ?? 'fc-bg-regular';
    return {
      html: `
        <div class="event-fc-color flex fc-event-main ${colorClass} p-1 rounded-sm">
          <div class="fc-daygrid-event-dot"></div>
          <div class="fc-event-time">${eventInfo.timeText || ''}</div>
          <div class="fc-event-title">${eventInfo.event.title}</div>
        </div>
      `,
    };
  }
}
