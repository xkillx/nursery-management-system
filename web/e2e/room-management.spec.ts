import { test, expect } from '@playwright/test';
import { loginAsManager, loginAsOwner, ensureTestRoomsExist } from './helpers/auth';
import { RoomManagementPage } from './page-objects/room-management.page';

const VIEWPORT = { width: 1280, height: 720 };

function uniqueName(prefix: string): string {
  return `${prefix} ${Date.now()}-${Math.random().toString(36).slice(2, 6)}`;
}

test.describe('Manager Room Management', () => {
  let roomsPage: RoomManagementPage;

  test.beforeEach(async ({ page }) => {
    await page.setViewportSize(VIEWPORT);
    await loginAsManager(page);
    await ensureTestRoomsExist(page);
    roomsPage = new RoomManagementPage(page, 'manager');
    await roomsPage.navigateToRoomsList();
  });

  test('create room with valid data', async ({ page }) => {
    const name = uniqueName('Test Room');
    await roomsPage.createRoom({ name, ageGroup: 'toddler', capacity: 15 });
    await page.waitForURL(/\/manager\/site-settings\/rooms$/);
    await roomsPage.expectRoomInList(name);
  });

  test('create room with description', async ({ page }) => {
    const name = uniqueName('Desc Room');
    await roomsPage.createRoom({ name, ageGroup: 'baby', capacity: 10, description: 'A cozy room for babies' });
    await page.waitForURL(/\/manager\/site-settings\/rooms$/);
    await roomsPage.expectRoomInList(name);
  });

  test('show validation error when name is missing', async () => {
    await roomsPage.navigateToCreateRoom();
    await roomsPage.page.locator('#room-age-group select').selectOption('toddler');
    await roomsPage.page.locator('#room-capacity').fill('10');
    await roomsPage.page.getByRole('button', { name: /Create Room/ }).click();
    await roomsPage.page.locator('#room-name').focus();
    await roomsPage.expectValidationError('required');
  });

  test('show validation error when age group is missing', async () => {
    await roomsPage.navigateToCreateRoom();
    await roomsPage.page.locator('#room-name').fill('Test Room');
    await roomsPage.page.locator('#room-capacity').fill('10');
    await roomsPage.page.getByRole('button', { name: /Create Room/ }).click();
    await roomsPage.expectValidationError('required');
  });

  test('show validation error when capacity is zero', async () => {
    await roomsPage.navigateToCreateRoom();
    await roomsPage.page.locator('#room-name').fill('Test Room');
    await roomsPage.page.locator('#room-age-group select').selectOption('toddler');
    await roomsPage.page.locator('#room-capacity').fill('0');
    await roomsPage.page.getByRole('button', { name: /Create Room/ }).click();
    await roomsPage.expectValidationError('at least');
  });

  test('edit room name and capacity', async ({ page }) => {
    const createName = uniqueName('Edit Me');
    await roomsPage.createRoom({ name: createName, ageGroup: 'preschool', capacity: 20 });

    const updatedName = uniqueName('Updated');
    await roomsPage.updateRoom(createName, { name: updatedName, capacity: 25 });
    await page.waitForURL(/\/manager\/site-settings\/rooms$/);
    await roomsPage.expectRoomInList(updatedName);
  });

  test('archive and reactivate room', async () => {
    const name = uniqueName('Archive Me');
    await roomsPage.createRoom({ name, ageGroup: 'mixed', capacity: 12 });
    await roomsPage.archiveRoom(name);
    await roomsPage.filterByStatus('archived');
    await roomsPage.expectRoomInList(name);

    await roomsPage.reactivateRoom(name);
    await roomsPage.filterByStatus('active');
    await roomsPage.expectRoomInList(name);
  });

  test('search rooms by name', async () => {
    const name = uniqueName('Searchable');
    await roomsPage.createRoom({ name, ageGroup: 'toddler', capacity: 10 });
    await roomsPage.searchRooms(name);
    await roomsPage.expectRoomInList(name);
  });

  test('filter by status', async () => {
    await roomsPage.filterByStatus('active');
    await roomsPage.page.waitForTimeout(500);
    await roomsPage.filterByStatus('archived');
    await roomsPage.page.waitForTimeout(500);
    await roomsPage.filterByStatus('all');
  });
});

test.describe('Owner Room Management', () => {
  let roomsPage: RoomManagementPage;

  test.beforeEach(async ({ page }) => {
    await page.setViewportSize(VIEWPORT);
    await loginAsOwner(page);
    await ensureTestRoomsExist(page);
    roomsPage = new RoomManagementPage(page, 'owner');
    await roomsPage.navigateToRoomsList();
  });

  test('site selector is visible on owner room list', async () => {
    await expect(roomsPage.page.locator('#room-site-select')).toBeVisible();
  });

  test('create room from owner view', async ({ page }) => {
    const name = uniqueName('Owner Room');
    await roomsPage.navigateToCreateRoom();
    await expect(page.locator('#form-site-select')).toBeVisible();
    await roomsPage.createRoom({ name, ageGroup: 'preschool', capacity: 18 });
    await page.waitForURL(/\/owner\/rooms$/);
    await roomsPage.expectRoomInList(name);
  });

  test('archive room from owner view', async () => {
    const name = uniqueName('Owner Archive');
    await roomsPage.createRoom({ name, ageGroup: 'baby', capacity: 8 });
    await roomsPage.archiveRoom(name);
    await roomsPage.filterByStatus('archived');
    await roomsPage.expectRoomInList(name);
  });
});
