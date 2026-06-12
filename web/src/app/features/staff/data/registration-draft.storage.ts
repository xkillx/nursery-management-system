import { Injectable } from '@angular/core';

const STORAGE_KEY = 'nursery.registration_intake.draft';
const STEP_STORAGE_KEY = 'nursery.registration_intake.draft.step';
const SAVED_AT_STORAGE_KEY = 'nursery.registration_intake.draft.savedAt';

@Injectable({ providedIn: 'root' })
export class RegistrationDraftStorage {
  private readonly hasStorage = typeof globalThis !== 'undefined' && typeof globalThis.localStorage !== 'undefined';

  load(): string | null {
    if (!this.hasStorage) return null;
    return globalThis.localStorage.getItem(STORAGE_KEY);
  }

  loadStep(): string | null {
    if (!this.hasStorage) return null;
    return globalThis.localStorage.getItem(STEP_STORAGE_KEY);
  }

  loadSavedAt(): string | null {
    if (!this.hasStorage) return null;
    return globalThis.localStorage.getItem(SAVED_AT_STORAGE_KEY);
  }

  save(payload: unknown, step: string): void {
    if (!this.hasStorage) return;
    try {
      globalThis.localStorage.setItem(STORAGE_KEY, JSON.stringify(payload));
      globalThis.localStorage.setItem(STEP_STORAGE_KEY, step);
      globalThis.localStorage.setItem(SAVED_AT_STORAGE_KEY, new Date().toISOString());
    } catch {
      // ignore quota or serialization errors
    }
  }

  clear(): void {
    if (!this.hasStorage) return;
    globalThis.localStorage.removeItem(STORAGE_KEY);
    globalThis.localStorage.removeItem(STEP_STORAGE_KEY);
    globalThis.localStorage.removeItem(SAVED_AT_STORAGE_KEY);
  }
}
