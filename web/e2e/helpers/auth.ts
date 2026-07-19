import { Page, expect } from '@playwright/test';

export async function loginAsManager(page: Page): Promise<void> {
  const email = process.env.E2E_MANAGER_EMAIL || 'manager@example.com';
  const password = process.env.E2E_MANAGER_PASSWORD || 'Pass1234';

  await page.goto('/signin');
  await page.locator('input[name="email"]').fill(email);
  await page.locator('input[name="password"]').fill(password);
  await page.getByRole('button', { name: 'Sign in' }).click();
  await page.waitForURL(/\/manager/);
}

export async function loginAsOwner(page: Page): Promise<void> {
  const email = process.env.E2E_OWNER_EMAIL || 'owner@example.com';
  const password = process.env.E2E_OWNER_PASSWORD || 'Pass1234';

  await page.goto('/signin');
  await page.locator('input[name="email"]').fill(email);
  await page.locator('input[name="password"]').fill(password);
  await page.getByRole('button', { name: 'Sign in' }).click();
  await page.waitForURL(/\/owner/);
}

export async function navigateToNewRegistration(page: Page): Promise<void> {
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
