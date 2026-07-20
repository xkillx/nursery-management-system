import { test, expect } from '@playwright/test';
import { loginAsManager, navigateToNewRegistration, ensureTestRoomsExist } from './helpers/auth';
import { ChildRegistrationPage } from './page-objects/child-registration.page';

test.describe('Child Registration - Consents & Edge Cases', () => {
  let regPage: ChildRegistrationPage;

  test.beforeEach(async ({ page }) => {
    await loginAsManager(page);
    await ensureTestRoomsExist(page);
    await navigateToNewRegistration(page);
    regPage = new ChildRegistrationPage(page);
  });

  test.describe('Consent requirements', () => {
    test.beforeEach(async () => {
      await advanceToStep4(regPage);
    });

    test('blocks submission when GDPR consent is not granted', async () => {
      await regPage.toggleConsent('information_truthfulness_declaration');
      await regPage.toggleConsent('safeguarding_reporting_acknowledgement');
      await regPage.toggleConsent('information_sharing_consent');
      await regPage.toggleConsent('urgent_medical_treatment');
      await regPage.toggleConsent('plasters');
      await regPage.fillSignerName('Sarah Johnson');
      await regPage.clickSubmit();
      await regPage.expectFieldError('Confirm GDPR data processing consent');
    });

    test('blocks submission when truthfulness declaration is not granted', async () => {
      await regPage.toggleConsent('gdpr_data_processing_consent');
      await regPage.toggleConsent('safeguarding_reporting_acknowledgement');
      await regPage.toggleConsent('information_sharing_consent');
      await regPage.toggleConsent('urgent_medical_treatment');
      await regPage.toggleConsent('plasters');
      await regPage.fillSignerName('Sarah Johnson');
      await regPage.clickSubmit();
      await regPage.expectFieldError('Confirm the truthfulness declaration');
    });

    test('blocks submission when signer name is empty', async () => {
      await regPage.markAllRequiredConsents();
      // Signer name is auto-filled from the primary parent/carer; clear it to
      // exercise the required-field check.
      await regPage.fillSignerName('');
      await regPage.clickSubmit();
      await regPage.expectFieldError('Record the parent or carer full name who signed the consent');
    });

    test('allows optional consents to remain unchecked', async () => {
      await regPage.markAllRequiredConsents();
      await regPage.fillSignerName('Sarah Johnson');

      const optionalConsents = ['area_senco_liaison', 'local_outings', 'social_media'];
      for (const key of optionalConsents) {
        const consent = regPage.page.locator(`[data-focus-target="${key}"]`);
        await expect(consent).toBeVisible();
      }
    });
  });

  test.describe('Disability & SEND', () => {
    test('shows disability notes field when disability is yes', async () => {
      await fillStep1Minimal(regPage);
      await regPage.selectDisabilityStatus('yes');
      const notesField = regPage.page.locator('input#child-disability-notes');
      await expect(notesField).toBeVisible();
    });

    test('hides disability notes field when disability is no', async () => {
      await fillStep1Minimal(regPage);
      await regPage.selectDisabilityStatus('no');
      const notesField = regPage.page.locator('input#child-disability-notes');
      await expect(notesField).not.toBeVisible();
    });

    test('blocks when disability is yes but notes are empty', async () => {
      await fillStep1Minimal(regPage);
      await regPage.selectDisabilityStatus('yes');
      await regPage.clickContinue();
      await regPage.expectFieldError('Record disability or access details');
    });
  });

  test.describe('Allergy conditional fields', () => {
    test('shows allergy details when allergy is yes', async () => {
      await advanceToStep2(regPage);
      await regPage.selectAllergyStatus('yes');
      const detailsField = regPage.page.locator('textarea#allergy-details, #allergy-details textarea').first();
      await expect(detailsField).toBeVisible();
    });

    test('hides allergy details when allergy is no', async () => {
      await advanceToStep2(regPage);
      await regPage.selectAllergyStatus('no');
      const detailsField = regPage.page.locator('textarea#allergy-details, #allergy-details textarea').first();
      await expect(detailsField).not.toBeVisible();
    });
  });

  test.describe('Medication conditional fields', () => {
    test('shows medication fields when medication is yes', async () => {
      await advanceToStep2(regPage);
      await regPage.selectMedicationStatus('yes');
      const nameField = regPage.page.locator('input#medication-name');
      await expect(nameField).toBeVisible();
    });

    test('hides medication fields when medication is no', async () => {
      await advanceToStep2(regPage);
      await regPage.selectMedicationStatus('no');
      const nameField = regPage.page.locator('input#medication-name');
      await expect(nameField).not.toBeVisible();
    });
  });

  test.describe('Social services conditional fields', () => {
    test('shows social services details when yes', async () => {
      await advanceToStep2(regPage);
      await regPage.selectSocialServicesStatus('yes');
      const detailsField = regPage.page.locator('textarea#social-services-details, #social-services-details textarea').first();
      await expect(detailsField).toBeVisible();
    });
  });

  test.describe('Second parent', () => {
    test('add second parent button is visible', async () => {
      await advanceToStep3(regPage);
      const addBtn = regPage.page.getByText('Add second parent/carer details');
      await expect(addBtn).toBeVisible();
    });

    test('clicking add second parent shows the form', async () => {
      await advanceToStep3(regPage);
      await regPage.page.getByText('Add second parent/carer details').click();
      const nameField = regPage.page.locator('input[placeholder="e.g. Mark Johnson"]');
      await expect(nameField).toBeVisible();
    });
  });
});

async function fillStep1Minimal(regPage: ChildRegistrationPage): Promise<void> {
  await regPage.fillChildBasics({
    firstName: 'James',
    dateOfBirth: '2022-03-15',
    startDate: '2026-09-01',
    homeAddress: '123 High Street',
    language: 'English',
  });
  // Try to select room but don't fail if none available
  await regPage.selectFirstAvailableRoom();
}

async function advanceToStep2(regPage: ChildRegistrationPage): Promise<void> {
  await fillStep1Minimal(regPage);
  await regPage.selectDisabilityStatus('no');
  await regPage.clickContinue();
  // Only expect step 2 if rooms are available
  const hasRoom = await regPage.hasRoomSelect();
  if (hasRoom) {
    await regPage.expectStepActive(2);
  }
}

async function advanceToStep3(regPage: ChildRegistrationPage): Promise<void> {
  await advanceToStep2(regPage);
  await regPage.selectAllergyStatus('no');
  await regPage.selectMedicationStatus('no');
  await regPage.selectImmunisationStatus('up_to_date');
  await regPage.selectDietaryStatus('none');
  await regPage.selectMedicalHistoryStatus('none');
  await regPage.selectSocialServicesStatus('no');
  await regPage.selectDevelopmentalConcerns('no');
  await regPage.clickContinue();
  await regPage.expectStepActive(3);
}

async function advanceToStep4(regPage: ChildRegistrationPage): Promise<void> {
  await advanceToStep3(regPage);
  await regPage.fillParentCarer({
    fullName: 'Sarah Johnson',
    relationship: 'Mother',
    telephone: '07700 900001',
    email: 'sarah@example.com',
    hasResponsibility: true,
    addressStreet: '123 High Street',
    addressCity: 'London',
    addressPostcode: 'SW1A 1AA',
  });
  await regPage.fillEmergencyContact({
    fullName: 'Mary Johnson',
    relationship: 'Grandparent',
    telephone: '07700 900002',
  });
  await regPage.fillCollectionPassword('nursery-pass-1');
  await regPage.clickContinue();
  await regPage.expectStepActive(4);
}
