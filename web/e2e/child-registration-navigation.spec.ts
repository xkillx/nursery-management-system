import { test, expect } from '@playwright/test';
import { loginAsManager, navigateToNewRegistration, ensureTestRoomsExist } from './helpers/auth';
import { ChildRegistrationPage } from './page-objects/child-registration.page';

test.describe('Child Registration - Step Navigation', () => {
  let regPage: ChildRegistrationPage;

  test.beforeEach(async ({ page }) => {
    await loginAsManager(page);
    await ensureTestRoomsExist(page);
    await navigateToNewRegistration(page);
    regPage = new ChildRegistrationPage(page);
  });

  test('starts on step 1', async () => {
    await regPage.expectStepActive(1);
  });

  test('step 2 is locked when step 1 is incomplete', async () => {
    await regPage.expectStepLocked(2);
  });

  test('step 3 is locked when step 1 is incomplete', async () => {
    await regPage.expectStepLocked(3);
  });

  test('step 4 is locked when step 1 is incomplete', async () => {
    await regPage.expectStepLocked(4);
  });

  test('can navigate back to step 1 from step 2', async () => {
    await fillStep1AndAdvance(regPage);
    await regPage.expectStepActive(2);
    await regPage.clickBack();
    await regPage.expectStepActive(1);
  });

  test('can navigate back to step 2 from step 3', async () => {
    await fillStep1AndAdvance(regPage);
    await fillStep2AndAdvance(regPage);
    await regPage.expectStepActive(3);
    await regPage.clickBack();
    await regPage.expectStepActive(2);
  });

  test('back navigation preserves form data', async () => {
    await fillStep1AndAdvance(regPage);
    await regPage.clickBack();
    await regPage.expectStepActive(1);

    const firstName = regPage.page.locator('#child-first-name');
    await expect(firstName).toHaveValue('James');
  });

  test('step 2 unlocks after step 1 is completed', async () => {
    await fillStep1AndAdvance(regPage);
    await regPage.expectStepActive(2);
  });

  test('step 3 unlocks after step 2 is completed', async () => {
    await fillStep1AndAdvance(regPage);
    await fillStep2AndAdvance(regPage);
    await regPage.expectStepActive(3);
  });

  test('step 4 unlocks after step 3 is completed', async () => {
    await fillStep1AndAdvance(regPage);
    await fillStep2AndAdvance(regPage);
    await fillStep3AndAdvance(regPage);
    await regPage.expectStepActive(4);
  });
});

async function fillStep1AndAdvance(regPage: ChildRegistrationPage): Promise<void> {
  await regPage.fillChildBasics({
    firstName: 'James',
    dateOfBirth: '2022-03-15',
    startDate: '2026-09-01',
    homeAddress: '123 High Street',
    language: 'English',
  });
  await regPage.selectFirstAvailableRoom();
  await regPage.selectDisabilityStatus('no');
  await regPage.clickContinue();
}

async function fillStep2AndAdvance(regPage: ChildRegistrationPage): Promise<void> {
  await regPage.selectAllergyStatus('no');
  await regPage.selectMedicationStatus('no');
  await regPage.selectImmunisationStatus('up_to_date');
  await regPage.selectDietaryStatus('none');
  await regPage.selectMedicalHistoryStatus('none');
  await regPage.selectSocialServicesStatus('no');
  await regPage.selectDevelopmentalConcerns('no');
  await regPage.clickContinue();
}

async function fillStep3AndAdvance(regPage: ChildRegistrationPage): Promise<void> {
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
  await regPage.clickContinue();
}
