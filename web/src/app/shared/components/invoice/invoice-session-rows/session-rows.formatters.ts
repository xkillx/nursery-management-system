import { InvoiceSession } from '../../../models/invoice-session.models';

const WEEKDAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

export function formatSessionDate(dateStr: string): string {
  const d = new Date(`${dateStr}T00:00:00`);
  if (isNaN(d.getTime())) return dateStr;
  return `${WEEKDAYS[d.getDay()]} ${d.getDate()} ${MONTHS[d.getMonth()]}`;
}

export function formatSessionTimes(session: Pick<InvoiceSession, 'startMinutes' | 'endMinutes'>): string | null {
  if (
    (session.startMinutes == null || session.startMinutes <= 0) &&
    (session.endMinutes == null || session.endMinutes <= 0)
  ) {
    return null;
  }
  const start = session.startMinutes != null ? formatMinutesOfDay(session.startMinutes) : '';
  const end = session.endMinutes != null ? formatMinutesOfDay(session.endMinutes) : '';
  return `${start}–${end}`;
}

export function formatMinutesOfDay(minutes: number): string {
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`;
}

export function formatSessionBooking(session: InvoiceSession): string {
  const times = formatSessionTimes(session);
  if (!session.sessionTypeName) return times ? `(${times})` : '';
  return times ? `${session.sessionTypeName} (${times})` : session.sessionTypeName;
}

export function formatSessionDescription(session: InvoiceSession): string {
  const date = formatSessionDate(session.occurrenceDate);
  return session.sessionTypeName ? `${date} · ${formatSessionBooking(session)}` : date;
}

export function formatSessionQuantity(minutes: number | null): string {
  if (minutes == null || minutes <= 0) return '—';
  if (minutes % 60 === 0) {
    return `${minutes / 60} hrs`;
  }
  const hours = minutes / 60;
  const trimmed = Number(hours.toFixed(1));
  return `${trimmed} hrs`;
}

export function formatUnitAmount(minor: number | null): string {
  if (minor == null) return '—';
  const pounds = minor / 100;
  return `£${pounds.toLocaleString('en-GB', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}
