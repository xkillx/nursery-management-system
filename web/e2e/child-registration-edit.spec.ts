import { test, expect } from '@playwright/test';
import { loginAsManager, ensureTestRoomsExist, registerChildAndGetId } from './helpers/auth';
import { ChildRegistrationPage } from './page-objects/child-registration.page';

const REGISTRATION_DATA = {
  child: {
    firstName: 'Emma',
    lastName: 'Wilson',
    dateOfBirth: '2022-06-10',
    startDate: '2026-09-01',
    homeAddress: '456 Oak Avenue, Manchester',
    language: 'English',
    sex: 'female',
  },
  parent: {
    fullName: 'Laura Wilson',
    relationship: 'Mother',
    telephone: '07700 900100',
    email: 'laura.wilson@example.com',
    hasResponsibility: true,
    addressStreet: '456 Oak Avenue',
    addressCity: 'Manchester',
    addressPostcode: 'M1 2AB',
  },
  emergencyContact: {
    fullName: 'David Wilson',
    relationship: 'Father',
    telephone: '07700 900101',
  },
  collectionPassword: 'oak-nursery-1',
  signerName: 'Laura Wilson',
};

test.describe('Child Registration - Edit Field Verification', () => {
  let regPage: ChildRegistrationPage;
  let childId: string;

  test.beforeEach(async ({ page }) => {
    await loginAsManager(page);
    await ensureTestRoomsExist(page);
    childId = await registerChildAndGetId(page);
    regPage = new ChildRegistrationPage(page);
    await regPage.navigateToEditChild(childId);
  });

  test('Step 1 fields match registration input', async () => {
    await test.step('Verify child basics', async () => {
      await regPage.expectFieldValue('child-first-name', REGISTRATION_DATA.child.firstName);
      await regPage.expectFieldValue('child-last-name', REGISTRATION_DATA.child.lastName);
      await regPage.expectFieldValue('child-home-address', REGISTRATION_DATA.child.homeAddress);
      await regPage.expectRadioSelected('child-sex', REGISTRATION_DATA.child.sex);
    });
  });

  test('Step 2 fields match registration input', async () => {
    await regPage.goToStep(2);
    await regPage.expectStepActive(2);

    await test.step('Verify allergy status', async () => {
      await regPage.expectRadioSelected('allergy_status', 'no');
    });

    await test.step('Verify medication status', async () => {
      await regPage.expectRadioSelected('medication_status', 'no');
    });

    await test.step('Verify immunisation status', async () => {
      await expect(regPage.page.locator('input#imm-up-to-date')).toBeChecked();
    });

    await test.step('Verify dietary status', async () => {
      await regPage.expectRadioSelected('dietary', 'none');
    });

    await test.step('Verify social services status', async () => {
      await regPage.expectRadioSelected('social_services_status', 'no');
    });
  });

  test('Step 3 fields match registration input', async () => {
    await regPage.goToStep(3);
    await regPage.expectStepActive(3);

    await test.step('Verify parent/carer details', async () => {
      await regPage.expectFieldValue('primary-full-name', REGISTRATION_DATA.parent.fullName);
      await regPage.expectFieldValue('primary-telephone', REGISTRATION_DATA.parent.telephone);
      await regPage.expectFieldValue('primary-email', REGISTRATION_DATA.parent.email);
    });

    await test.step('Verify collection password', async () => {
      await regPage.expectFieldValue('collection-password', REGISTRATION_DATA.collectionPassword);
    });
  });

  test('Step 4 consents and signer name match registration input', async () => {
    await regPage.goToStep(4);
    await regPage.expectStepActive(4);

    await test.step('Verify required consents are checked', async () => {
      await regPage.expectCheckboxChecked('gdpr_data_processing_consent');
      await regPage.expectCheckboxChecked('information_truthfulness_declaration');
      await regPage.expectCheckboxChecked('safeguarding_reporting_acknowledgement');
      await regPage.expectCheckboxChecked('information_sharing_consent');
      await regPage.expectCheckboxChecked('urgent_medical_treatment');
      await regPage.expectCheckboxChecked('plasters');
    });

    await test.step('Verify signer name', async () => {
      await regPage.expectFieldValue('signer-name', REGISTRATION_DATA.signerName);
    });
  });
});

test.describe('Child Registration - Edit and Save', () => {
  let regPage: ChildRegistrationPage;
  let childId: string;

  test.beforeEach(async ({ page }) => {
    await loginAsManager(page);
    await ensureTestRoomsExist(page);
    childId = await registerChildAndGetId(page);
    regPage = new ChildRegistrationPage(page);
    await regPage.navigateToEditChild(childId);
  });

  test('modified first name persists after save + reload', async ({ page }) => {
    const updatedFirstName = 'Emily';

    await test.step('Modify first name and save', async () => {
      await page.locator('input#child-first-name').clear();
      await page.locator('input#child-first-name').fill(updatedFirstName);
      await regPage.clickSaveChanges();
      await regPage.expectToastVisible();
    });

    await test.step('Reload and verify', async () => {
      await regPage.navigateToEditChild(childId);
      await regPage.expectFieldValue('child-first-name', updatedFirstName);
    });
  });

  test('modified parent phone persists after save + reload', async ({ page }) => {
    const updatedPhone = '07700 999999';

    await test.step('Navigate to step 3, modify phone, and save', async () => {
      await regPage.goToStep(3);
      await regPage.expectStepActive(3);
      await page.locator('input#primary-telephone').clear();
      await page.locator('input#primary-telephone').fill(updatedPhone);
      await regPage.clickSaveChanges();
      await regPage.expectToastVisible();
    });

    await test.step('Reload and verify', async () => {
      await regPage.navigateToEditChild(childId);
      await regPage.goToStep(3);
      await regPage.expectFieldValue('primary-telephone', updatedPhone);
    });
  });

  test('allergy details persist after save + reload', async ({ page }) => {
    const allergyDetails = 'Peanut allergy - severe';

    await test.step('Navigate to step 2, set allergy to yes, add details, and save', async () => {
      await regPage.goToStep(2);
      await regPage.expectStepActive(2);
      await regPage.selectAllergyStatus('yes');
      await regPage.fillAllergyDetails(allergyDetails);
      await regPage.clickSaveChanges();
      await regPage.expectToastVisible();
    });

    await test.step('Reload and verify', async () => {
      await regPage.navigateToEditChild(childId);
      await regPage.goToStep(2);
      await regPage.expectRadioSelected('allergy_status', 'yes');
      await expect(page.locator('textarea#allergy-details')).toHaveValue(allergyDetails);
    });
  });
});
