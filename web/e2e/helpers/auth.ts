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

export async function navigateToNewRegistration(page: Page): Promise<void> {
  await page.goto('/manager/children/new');
  await expect(page.getByText('Child Profile')).toBeVisible();
}

export async function ensureTestRoomExists(page: Page): Promise<void> {
  const apiBase = process.env.E2E_API_BASE || 'http://localhost:8080';
  
  // Get site ID from the current URL or page state
  const siteId = await page.evaluate(async (apiBase) => {
    const token = localStorage.getItem('token');
    if (!token) return null;
    
    // Try to get the user's sites
    const res = await fetch(`${apiBase}/api/v1/sites`, {
      headers: { 'Authorization': `Bearer ${token}` },
    });
    if (!res.ok) return null;
    const data = await res.json();
    return data?.[0]?.id || null;
  }, apiBase);

  if (!siteId) return;

  // Check if rooms exist
  const hasRooms = await page.evaluate(async ({ apiBase, siteId }) => {
    const token = localStorage.getItem('token');
    const res = await fetch(`${apiBase}/api/v1/sites/${siteId}/rooms`, {
      headers: { 'Authorization': `Bearer ${token}` },
    });
    if (!res.ok) return false;
    const data = await res.json();
    return data?.items?.length > 0;
  }, { apiBase, siteId });

  if (!hasRooms) {
    // Create a test room
    await page.evaluate(async ({ apiBase, siteId }) => {
      const token = localStorage.getItem('token');
      await fetch(`${apiBase}/api/v1/sites/${siteId}/rooms`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: 'Test Room',
          age_group: '3-5',
          capacity: 20,
        }),
      });
    }, { apiBase, siteId });
  }
}
