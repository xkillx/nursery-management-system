import { StatusFilter } from '../models/children.models';

export function formatHourlyRateGbp(minorUnits: number | null): string {
  if (minorUnits === null || minorUnits === undefined) {
    return 'Not set';
  }
  const pounds = minorUnits / 100;
  return `£${pounds.toFixed(2)}/hr`;
}

export function formatSiteRate(minorUnits: number | null | undefined): string {
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

export interface ChildNameParts {
	firstName?: string | null;
	middleName?: string | null;
	lastName?: string | null;
}

export function formatChildName(name: ChildNameParts): string {
	return [name.firstName, name.middleName, name.lastName]
		.map((part) => part?.trim() ?? '')
		.filter(Boolean)
		.join(' ');
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
	first_name: 'First name',
	date_of_birth: 'Date of birth',
  start_date: 'Start date',
  parent_carer_contact: 'Parent carer contact',
};

export function missingRequirementLabel(code: string): string {
  return REQUIREMENT_LABELS[code] ?? code;
}
