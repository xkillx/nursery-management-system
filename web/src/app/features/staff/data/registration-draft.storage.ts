import { Injectable } from '@angular/core';

interface DraftEnvelope {
  step: number;
  payload: unknown;
  savedAt: string;
}

@Injectable({ providedIn: 'root' })
export class RegistrationDraftStorage {
  private readonly key = 'nursery.registration_intake.draft';

  save(payload: unknown, step: number | string): void {
    try {
      const envelope: DraftEnvelope = { step: Number(step), payload, savedAt: new Date().toISOString() };
      localStorage.setItem(this.key, JSON.stringify(envelope));
    } catch {
      // ignore
    }
  }

  load(): string | null {
    try {
      const raw = localStorage.getItem(this.key);
      if (!raw) return null;
      // Support both the legacy plain-object format and the new envelope.
      try {
        const parsed = JSON.parse(raw) as DraftEnvelope;
        if (parsed && typeof parsed === 'object' && 'payload' in parsed) {
          return JSON.stringify(parsed.payload ?? {});
        }
      } catch {
        // not an envelope
      }
      return raw;
    } catch {
      return null;
    }
  }

  loadSavedAt(): string | null {
    try {
      const raw = localStorage.getItem(this.key);
      if (!raw) return null;
      const parsed = JSON.parse(raw) as DraftEnvelope;
      if (parsed && typeof parsed === 'object' && 'savedAt' in parsed) {
        return parsed.savedAt ?? null;
      }
      return null;
    } catch {
      return null;
    }
  }

  clear(): void {
    try {
      localStorage.removeItem(this.key);
    } catch {
      // ignore
    }
  }
}
