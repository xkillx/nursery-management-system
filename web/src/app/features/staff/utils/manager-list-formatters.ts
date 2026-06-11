import { StatusFilter } from '../models/children.models';

export function formatHourlyRateGbp(minorUnits: number | null): string {
  if (minorUnits === null || minorUnits === undefined) {
    return 'Not set';
  }
  const pounds = minorUnits / 100;
  return `£${pounds.toFixed(2)}/hr`;
}

export function minorToPounds(minorUnits: number | null): number | null {
  if (minorUnits === null || minorUnits === undefined) {
    return null;
  }
  return minorUnits / 100;
}

export function poundsToMinor(pounds: number | string): number {
  return Math.round(Number(pounds) * 100);
}

export function statusFilterLabel(status: StatusFilter): string {
  const labels: Record<StatusFilter, string> = {
    active: 'Active',
    inactive: 'Inactive',
    all: 'All',
  };
  return labels[status];
}

const REQUIREMENT_LABELS: Record<string, string> = {
  full_name: 'Full name',
  date_of_birth: 'Date of birth',
  start_date: 'Start date',
  billing_rate: 'Billing rate',
  guardian_link: 'Linked guardian',
};

export function missingRequirementLabel(code: string): string {
  return REQUIREMENT_LABELS[code] ?? code;
}
