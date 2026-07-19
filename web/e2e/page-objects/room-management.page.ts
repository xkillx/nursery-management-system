import { Page, Locator, expect } from '@playwright/test';

export class RoomManagementPage {
  readonly page: Page;
  readonly role: 'manager' | 'owner';

  constructor(page: Page, role: 'manager' | 'owner') {
    this.page = page;
    this.role = role;
  }

  // Navigation

  async navigateToRoomsList(): Promise<void> {
    const path = this.role === 'manager' ? '/manager/site-settings/rooms' : '/owner/rooms';
    await this.page.goto(path);
    await this.page.waitForLoadState('networkidle');
  }

  async navigateToCreateRoom(): Promise<void> {
    await this.page.locator('[data-testid="room-add-link"]').click();
    await this.page.waitForLoadState('networkidle');
  }

  async navigateToEditRoom(roomName: string): Promise<void> {
    await this.page.locator(`a[aria-label="Edit room ${roomName}"]`).click();
    await this.page.waitForLoadState('networkidle');
  }

  // Room form actions

  async createRoom(data: { name: string; ageGroup: string; capacity: number; description?: string }): Promise<void> {
    await this.navigateToCreateRoom();
    await this.fillForm(data);
    await this.page.getByRole('button', { name: /Create Room/ }).click();
    await this.page.waitForLoadState('networkidle');
  }

  async updateRoom(roomName: string, data: { name?: string; ageGroup?: string; capacity?: number; description?: string }): Promise<void> {
    await this.navigateToEditRoom(roomName);
    if (data.name) await this.input('room-name').fill(data.name);
    if (data.ageGroup) await this.selectOption('room-age-group', data.ageGroup);
    if (data.capacity !== undefined) await this.input('room-capacity').fill(String(data.capacity));
    if (data.description !== undefined) await this.textarea('room-description').fill(data.description);
    await this.page.getByRole('button', { name: /Save Room/ }).click();
    await this.page.waitForLoadState('networkidle');
  }

  // Archive / reactivate

  async archiveRoom(roomName: string): Promise<void> {
    this.page.once('dialog', (dialog) => dialog.accept());
    await this.page.locator(`button[aria-label="Archive room ${roomName}"]`).click();
    await this.page.waitForLoadState('networkidle');
  }

  async reactivateRoom(roomName: string): Promise<void> {
    await this.page.locator(`button[aria-label="Reactivate room ${roomName}"]`).click();
    await this.page.waitForLoadState('networkidle');
  }

  // Search / filter

  async searchRooms(query: string): Promise<void> {
    await this.page.locator('#rooms-search').fill(query);
  }

  async filterByStatus(status: 'all' | 'active' | 'archived'): Promise<void> {
    const select = this.page.locator('#room-status-filter select');
    await select.selectOption(status);
  }

  // Owner site selector

  async selectSite(siteName: string): Promise<void> {
    const select = this.page.locator('#room-site-select select');
    await select.selectOption({ label: siteName });
    await this.page.waitForLoadState('networkidle');
  }

  // Assertions

  async expectRoomInList(name: string): Promise<void> {
    await expect(this.page.getByText(name, { exact: false }).first()).toBeVisible();
  }

  async expectRoomNotInList(name: string): Promise<void> {
    await expect(this.page.getByText(name, { exact: false })).not.toBeVisible();
  }

  async expectValidationError(message: string): Promise<void> {
    await expect(this.page.getByText(message)).toBeVisible();
  }

  async expectFormTitle(title: string): Promise<void> {
    await expect(this.page.getByText(title, { exact: false }).first()).toBeVisible();
  }

  // Private helpers

  private input(id: string): Locator {
    return this.page.locator(`input#${id}`);
  }

  private textarea(id: string): Locator {
    return this.page.locator(`textarea#${id}`);
  }

  private async selectOption(id: string, value: string): Promise<void> {
    const select = this.page.locator(`#${id} select`);
    await select.selectOption(value);
  }

  private async fillForm(data: { name: string; ageGroup: string; capacity: number; description?: string }): Promise<void> {
    await this.input('room-name').fill(data.name);
    await this.selectOption('room-age-group', data.ageGroup);
    await this.input('room-capacity').fill(String(data.capacity));
    if (data.description) {
      await this.textarea('room-description').fill(data.description);
    }
  }
}
