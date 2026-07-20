import { Page, expect } from '@playwright/test';
import { ChildRegistrationPage } from '../page-objects/child-registration.page';

export async function loginAsManager(page: Page): Promise<void> {
  const email = process.env.E2E_MANAGER_EMAIL || 'manager@example.com';
  const password = process.env.E2E_MANAGER_PASSWORD || 'Pass1234';

  await page.goto('/signin');
  await clearAuthState(page);
  await page.locator('input[name="email"]').fill(email);
  await page.locator('input[name="password"]').fill(password);
  await page.getByRole('button', { name: 'Sign in' }).click();
  await page.waitForURL(/\/manager/, { timeout: 20000 });
}

export async function loginAsOwner(page: Page): Promise<void> {
  const email = process.env.E2E_OWNER_EMAIL || 'owner@example.com';
  const password = process.env.E2E_OWNER_PASSWORD || 'Pass1234';

  await page.goto('/signin');
  await clearAuthState(page);
  await page.locator('input[name="email"]').fill(email);
  await page.locator('input[name="password"]').fill(password);
  await page.getByRole('button', { name: 'Sign in' }).click();
  await page.waitForURL(/\/owner/, { timeout: 20000 });
}

// Clear any stale tokens / registration draft left over from a prior test so the
// app doesn't try a 10s refresh against an expired session before login.
async function clearAuthState(page: Page): Promise<void> {
  await page.evaluate(() => {
    localStorage.clear();
    sessionStorage.clear();
  });
}

export async function navigateToNewRegistration(page: Page): Promise<void> {
  // Clear any in-progress registration draft so the wizard always opens on step 1
  // (otherwise a restored draft from a prior test lands on a later step and the
  // "Child Profile" heading is not present).
  await page.evaluate(() => localStorage.removeItem('nursery.registration_intake.draft'));
  await page.goto('/manager/children/new');
  await expect(page.getByText('Child Profile')).toBeVisible();
}

const REQUIRED_ROOMS = [
  { name: 'Baby Room', age_group: 'baby', capacity: 12 },
  { name: 'Toddler Room', age_group: 'toddler', capacity: 16 },
  { name: 'Preschool Room', age_group: 'preschool', capacity: 20 },
] as const;

async function getSiteId(page: Page, apiBase: string): Promise<string | null> {
  return page.evaluate(async (apiBase) => {
    const token = localStorage.getItem('token');
    if (!token) return null;
    const res = await fetch(`${apiBase}/api/v1/sites`, {
      headers: { 'Authorization': `Bearer ${token}` },
    });
    if (!res.ok) return null;
    const data = await res.json();
    return data?.[0]?.id || null;
  }, apiBase);
}

async function getExistingAgeGroups(page: Page, apiBase: string, siteId: string): Promise<string[]> {
  return page.evaluate(async ({ apiBase, siteId }) => {
    const token = localStorage.getItem('token');
    const res = await fetch(`${apiBase}/api/v1/sites/${siteId}/rooms?include_archived=true`, {
      headers: { 'Authorization': `Bearer ${token}` },
    });
    if (!res.ok) return [];
    const data = await res.json();
    return (data?.items ?? []).map((r: { age_group: string }) => r.age_group);
  }, { apiBase, siteId });
}

async function createRoom(page: Page, apiBase: string, siteId: string, room: { name: string; age_group: string; capacity: number }): Promise<void> {
  await page.evaluate(async ({ apiBase, siteId, room }) => {
    const token = localStorage.getItem('token');
    await fetch(`${apiBase}/api/v1/sites/${siteId}/rooms`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(room),
    });
  }, { apiBase, siteId, room });
}

export async function ensureTestRoomsExist(page: Page): Promise<void> {
  const apiBase = process.env.E2E_API_BASE || 'http://localhost:8080';
  const siteId = await getSiteId(page, apiBase);
  if (!siteId) return;

  const existing = await getExistingAgeGroups(page, apiBase, siteId);
  for (const room of REQUIRED_ROOMS) {
    if (!existing.includes(room.age_group)) {
      await createRoom(page, apiBase, siteId, room);
    }
  }
}

export async function ensureTestRoomExists(page: Page): Promise<void> {
  const apiBase = process.env.E2E_API_BASE || 'http://localhost:8080';
  const siteId = await getSiteId(page, apiBase);
  if (!siteId) return;

  const existing = await getExistingAgeGroups(page, apiBase, siteId);
  if (existing.length === 0) {
    await createRoom(page, apiBase, siteId, { name: 'Test Room', age_group: 'preschool', capacity: 20 });
  }
}

export async function registerChildAndGetId(page: Page): Promise<string> {
  await navigateToNewRegistration(page);
  const regPage = new ChildRegistrationPage(page);

  const hasRoom = await regPage.hasRoomSelect();
  if (!hasRoom) {
    throw new Error('No rooms configured — cannot register child');
  }

  // Step 1: Child basics
  await regPage.fillChildBasics({
    firstName: 'Emma',
    lastName: 'Wilson',
    dateOfBirth: '2022-06-10',
    startDate: '2026-09-01',
    homeAddress: '456 Oak Avenue, Manchester',
    language: 'English',
    sex: 'female',
  });
  await regPage.selectFirstAvailableRoom();
  await regPage.selectDisabilityStatus('no');
  await regPage.clickContinue();
  await regPage.expectStepActive(2);

  // Step 2: Medical & health
  await regPage.selectAllergyStatus('no');
  await regPage.selectMedicationStatus('no');
  await regPage.selectImmunisationStatus('up_to_date');
  await regPage.selectDietaryStatus('none');
  await regPage.selectMedicalHistoryStatus('none');
  await regPage.selectSocialServicesStatus('no');
  await regPage.selectDevelopmentalConcerns('no');
  await regPage.clickContinue();
  await regPage.expectStepActive(3);

  // Step 3: Contacts & collection
  await regPage.fillParentCarer({
    fullName: 'Laura Wilson',
    relationship: 'Mother',
    telephone: '07700 900100',
    email: 'laura.wilson@example.com',
    hasResponsibility: true,
    addressStreet: '456 Oak Avenue',
    addressCity: 'Manchester',
    addressPostcode: 'M1 2AB',
  });
  await regPage.fillEmergencyContact({
    fullName: 'David Wilson',
    relationship: 'Father',
    telephone: '07700 900101',
  });
  await regPage.fillCollectionPassword('oak-nursery-1');
  await regPage.clickContinue();
  await regPage.expectStepActive(4);

  // Step 4: Consents and submit
  await regPage.markAllRequiredConsents();
  await regPage.fillSignerName('Laura Wilson');
  await regPage.fillSignedDate('2026-07-20');

  await regPage.clickSubmit();

  await expect(page).not.toHaveURL(/\/manager\/children\/new/, { timeout: 10000 });

  const url = page.url();
  const match = url.match(/\/manager\/children\/([^/]+)/);
  if (!match) {
    throw new Error(`Could not extract child ID from URL: ${url}`);
  }
  return match[1];
}
