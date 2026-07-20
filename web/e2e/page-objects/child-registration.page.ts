import { Page, Locator, expect } from '@playwright/test';

// Required consent model keys whose DOM toggle uses a different data-focus-target id.
const CONSENT_FOCUS_TARGET_OVERRIDES: Record<string, string> = {
  information_truthfulness_declaration: 'truthfulness-declaration',
  gdpr_data_processing_consent: 'gdpr-consent',
};

function consentFocusTarget(key: string): string {
  return CONSENT_FOCUS_TARGET_OVERRIDES[key] ?? key;
}

export class ChildRegistrationPage {
  readonly page: Page;

  constructor(page: Page) {
    this.page = page;
  }

  // Step navigation
  async goToStep(stepNumber: number): Promise<void> {
    const stepButtons = this.page.locator('nav[aria-label="Registration progress"] ol li button');
    await stepButtons.nth(stepNumber - 1).click();
  }

  async expectStepActive(stepNumber: number): Promise<void> {
    const stepButtons = this.page.locator('nav[aria-label="Registration progress"] ol li button');
    await expect(stepButtons.nth(stepNumber - 1)).toHaveAttribute('aria-current', 'step');
  }

  async clickContinue(): Promise<void> {
    await this.page.getByRole('button', { name: /Continue/ }).click();
  }

  async clickSaveChanges(): Promise<void> {
    await this.page.getByRole('button', { name: /Save changes/ }).click();
  }

  async clickBack(): Promise<void> {
    await this.page.getByRole('button', { name: /Back/ }).click();
  }

  // Step 1: Child Basics
  async fillChildBasics(data: {
    firstName: string;
    lastName?: string;
    dateOfBirth: string;
    startDate: string;
    homeAddress: string;
    language?: string;
    sex?: string;
  }): Promise<void> {
    await this.input('child-first-name').fill(data.firstName);
    if (data.lastName) {
      await this.input('child-last-name').fill(data.lastName);
    }
    await this.fillDatePicker('child-date-of-birth', data.dateOfBirth);
    await this.fillDatePicker('child-start-date', data.startDate);
    await this.input('child-home-address').fill(data.homeAddress);

    if (data.language) {
      await this.selectDropdown('child-first-language', data.language);
    }

    if (data.sex) {
      await this.clickRadio('child-sex', data.sex);
    }
  }

  async selectFirstAvailableRoom(): Promise<void> {
    // Wait for the room select to appear (loaded via API)
    const select = this.page.locator('select#child-primary-room, #child-primary-room select').first();
    try {
      await select.waitFor({ state: 'visible', timeout: 5000 });
      const options = select.locator('option:not([value=""])');
      const count = await options.count();
      if (count > 0) {
        const value = await options.first().getAttribute('value');
        if (value) await select.selectOption(value);
      }
    } catch {
      // Room select may not appear if API returns no rooms
      // Log for debugging but don't fail - the form validation will catch this
      console.log('Warning: No room select found - rooms may not be configured');
    }
  }

  async hasRoomSelect(): Promise<boolean> {
    const select = this.page.locator('select#child-primary-room, #child-primary-room select').first();
    return select.isVisible({ timeout: 2000 }).catch(() => false);
  }

  async selectDisabilityStatus(status: 'yes' | 'no'): Promise<void> {
    await this.clickRadio('child-disability', status);
  }

  // Step 2: Medical & Health
  async selectAllergyStatus(status: 'yes' | 'no'): Promise<void> {
    await this.clickRadio('allergy-status', status);
  }

  async fillAllergyDetails(details: string): Promise<void> {
    // id is duplicated on host <app-text-area> and inner <textarea>; target the textarea tag.
    await this.page.locator('textarea#allergy-details').fill(details);
  }

  async selectMedicationStatus(status: 'yes' | 'no'): Promise<void> {
    await this.clickRadio('medication-status', status);
  }

  async fillMedicationDetails(name: string, dosage: string): Promise<void> {
    await this.input('medication-name').fill(name);
    await this.input('medication-dosage').fill(dosage);
  }

  async selectDietaryStatus(status: 'none' | 'details'): Promise<void> {
    await this.clickRadio('dietary', status);
  }

  async selectMedicalHistoryStatus(status: 'none' | 'details'): Promise<void> {
    await this.clickRadio('med-history', status);
  }

  async selectSocialServicesStatus(status: 'yes' | 'no'): Promise<void> {
    await this.clickRadio('social-services', status);
  }

  async selectImmunisationStatus(status: 'up_to_date' | 'partial' | 'refused'): Promise<void> {
    const id = status === 'up_to_date' ? 'imm-up-to-date' : `imm-${status}`;
    await this.clickRadioByFor(id);
  }

  async selectDevelopmentalConcerns(value: 'yes' | 'no'): Promise<void> {
    await this.clickRadio('developmental-concerns', value);
  }

  // Step 3: Contacts & Collection
  async fillParentCarer(data: {
    fullName: string;
    relationship: string;
    telephone: string;
    email?: string;
    hasResponsibility?: boolean;
    addressStreet?: string;
    addressCity?: string;
    addressPostcode?: string;
  }): Promise<void> {
    await this.input('primary-full-name').fill(data.fullName);
    await this.selectDropdown('primary-relationship', data.relationship);
    await this.input('primary-telephone').fill(data.telephone);
    if (data.email) {
      await this.input('primary-email').fill(data.email);
    }
    if (data.addressStreet) {
      await this.input('parent1-address-street').fill(data.addressStreet);
    }
    if (data.addressCity) {
      await this.input('parent1-address-city').fill(data.addressCity);
    }
    if (data.addressPostcode) {
      await this.input('parent1-address-postcode').fill(data.addressPostcode);
    }
    if (data.hasResponsibility !== undefined) {
      await this.clickRadio('parent1-responsibility', data.hasResponsibility ? 'yes' : 'no');
    }
  }

  async fillEmergencyContact(data: {
    fullName: string;
    relationship: string;
    telephone: string;
  }): Promise<void> {
    const contactSection = this.page.locator('[data-focus-target="emergency-contacts-group"]');
    await contactSection.locator('input[type="text"]').first().fill(data.fullName);
    await contactSection.locator('select').first().selectOption({ label: data.relationship });
    await contactSection.locator('input[type="tel"]').fill(data.telephone);
  }

  async fillCollectionPassword(password: string): Promise<void> {
    // id duplicated on host <app-input-field> and inner <input>; target the input tag.
    await this.page.locator('input#collection-password').fill(password);
  }

  // Step 4: Consents
  async toggleConsent(key: string): Promise<void> {
    await this.page.locator(`[data-focus-target="${consentFocusTarget(key)}"]`).click();
  }

  async markAllRequiredConsents(): Promise<void> {
    const requiredKeys = [
      'gdpr_data_processing_consent',
      'information_truthfulness_declaration',
      'safeguarding_reporting_acknowledgement',
      'information_sharing_consent',
      'urgent_medical_treatment',
      'plasters',
    ];
    for (const key of requiredKeys) {
      await this.page.locator(`[data-focus-target="${consentFocusTarget(key)}"]`).click();
    }
  }

  async fillSignerName(name: string): Promise<void> {
    await this.input('signer-name').fill(name);
  }

  async fillSignedDate(dateIso: string): Promise<void> {
    // date-picker host + inner input share id; drive flatpickr directly.
    const ok = await this.page.evaluate(({ id, dateIso }) => {
      const inputs = document.querySelectorAll(`input#${id}`);
      for (const input of Array.from(inputs)) {
        const fp = (input as any)._flatpickr;
        if (fp) {
          fp.setDate(dateIso, true);
          return true;
        }
      }
      return false;
    }, { id: 'signed-date', dateIso });
    if (!ok) {
      await this.page.locator('input#signed-date').last().fill(dateIso);
    }
  }

  async clickSubmit(): Promise<void> {
    await this.page.getByRole('button', { name: /Submit registration|Mark Reviewed|Complete|Create child/i }).click();
  }

  // Validation helpers
  async expectFieldError(message: string): Promise<void> {
    // Scope to inline field-error <p> (id ends in -error). The toast surfaces the
    // same message via role=status, which would otherwise trigger strict-mode.
    const errorP = this.page.locator('p[id$="-error"]').filter({ hasText: message });
    await expect(errorP).toBeVisible();
  }

  async expectNoFieldError(message: string): Promise<void> {
    const errorP = this.page.locator('p[id$="-error"]').filter({ hasText: message });
    await expect(errorP).toHaveCount(0);
  }

  async expectStepLocked(stepNumber: number): Promise<void> {
    const stepButtons = this.page.locator('nav[aria-label="Registration progress"] ol li button');
    await expect(stepButtons.nth(stepNumber - 1)).toHaveAttribute('aria-disabled', 'true');
  }

  async expectToastVisible(): Promise<void> {
    await expect(this.page.locator('.toast, [role="status"]').first()).toBeVisible({ timeout: 5000 });
  }

  // Edit mode helpers
  async navigateToEditChild(childId: string): Promise<void> {
    await this.page.goto(`/manager/children/${childId}/edit`);
    const stepButtons = this.page.locator('nav[aria-label="Registration progress"] ol li button');
    await expect(stepButtons.first()).toBeVisible({ timeout: 10000 });
  }

  async expectFieldValue(fieldId: string, expectedValue: string): Promise<void> {
    const input = this.page.locator(`input#${fieldId}, select#${fieldId}`).first();
    await expect(input).toHaveValue(expectedValue);
  }

  async expectRadioSelected(name: string, value: string): Promise<void> {
    const radio = this.page.locator(`input[name="${name}"][value="${value}"]`);
    await expect(radio).toBeChecked();
  }

  async expectCheckboxChecked(fieldId: string): Promise<void> {
    const checkbox = this.page.locator(`input#${fieldId}`);
    await expect(checkbox).toBeChecked();
  }

  // --- Private helpers ---

  private input(id: string): Locator {
    return this.page.locator(`input#${id}`);
  }

  private async clickRadio(name: string, value: string): Promise<void> {
    await this.page.locator(`label[for="${name}-${value}"]`).click();
  }

  private async clickRadioByFor(forValue: string): Promise<void> {
    await this.page.locator(`label[for="${forValue}"]`).click();
  }

  private async fillDatePicker(id: string, dateStr: string): Promise<void> {
    // id is duplicated on host <app-date-picker> and inner <input>.
    // Query the actual <input> element which holds the flatpickr instance.
    const result = await this.page.evaluate(
      ({ id, dateStr }) => {
        const inputs = document.querySelectorAll(`input#${id}`);
        for (const input of Array.from(inputs)) {
          const fp = (input as any)._flatpickr;
          if (fp) {
            fp.setDate(dateStr, true);
            return true;
          }
        }
        return false;
      },
      { id, dateStr }
    );
    if (!result) {
      // Fallback: type into the input directly
      const input = this.page.locator(`input#${id}`).last();
      await input.click();
      await input.fill(dateStr);
      await input.press('Escape');
    }
  }

  private async selectDropdown(id: string, optionText: string): Promise<void> {
    const select = this.page.locator(`#${id} select`).first();
    if (await select.isVisible({ timeout: 2000 }).catch(() => false)) {
      await select.selectOption({ label: optionText });
    } else {
      await this.page.locator(`#${id}`).click();
      await this.page.getByRole('option', { name: optionText }).click();
    }
  }
}
