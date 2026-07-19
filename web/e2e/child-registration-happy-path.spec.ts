import { test, expect } from '@playwright/test';
import { loginAsManager, navigateToNewRegistration, ensureTestRoomsExist } from './helpers/auth';
import { ChildRegistrationPage } from './page-objects/child-registration.page';

test.describe('Child Registration - Happy Path', () => {
  let regPage: ChildRegistrationPage;

  test.beforeEach(async ({ page }) => {
    await loginAsManager(page);
    await ensureTestRoomsExist(page);
    await navigateToNewRegistration(page);
    regPage = new ChildRegistrationPage(page);
  });

  test('complete registration through all 4 steps', async ({ page }) => {
    // Skip if no rooms available
    const hasRoom = await regPage.hasRoomSelect();
    if (!hasRoom) {
      test.skip(true, 'No rooms configured - skipping full registration test');
      return;
    }

    await test.step('Step 1: Fill child basics', async () => {
      await regPage.fillChildBasics({
        firstName: 'James',
        lastName: 'Smith',
        dateOfBirth: '2022-03-15',
        startDate: '2026-09-01',
        homeAddress: '123 High Street, London',
        language: 'English',
        sex: 'male',
      });
      await regPage.selectFirstAvailableRoom();
      await regPage.selectDisabilityStatus('no');
      await regPage.clickContinue();
      await regPage.expectStepActive(2);
    });

    await test.step('Step 2: Fill medical & health', async () => {
      await regPage.selectAllergyStatus('no');
      await regPage.selectMedicationStatus('no');
      await regPage.selectImmunisationStatus('up_to_date');
      await regPage.selectDietaryStatus('none');
      await regPage.selectMedicalHistoryStatus('none');
      await regPage.selectSocialServicesStatus('no');
      await regPage.selectDevelopmentalConcerns('no');
      await regPage.clickContinue();
      await regPage.expectStepActive(3);
    });

    await test.step('Step 3: Fill contacts & collection', async () => {
      await regPage.fillParentCarer({
        fullName: 'Sarah Johnson',
        relationship: 'Mother',
        telephone: '07700 900001',
        email: 'sarah.johnson@example.com',
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
      await regPage.expectStepActive(4);
    });

    await test.step('Step 4: Fill consents and submit', async () => {
      await regPage.markAllRequiredConsents();
      await regPage.fillSignerName('Sarah Johnson');

      await regPage.clickSubmit();

      await expect(page).not.toHaveURL(/\/manager\/children\/new/);
    });
  });

  test('child name appears in avatar initials', async () => {
    await regPage.fillChildBasics({
      firstName: 'James',
      lastName: 'Smith',
      dateOfBirth: '2022-03-15',
      startDate: '2026-09-01',
      homeAddress: '123 High Street',
    });

    const avatar = regPage.page.locator('.rounded-2xl.bg-gradient-to-tr');
    await expect(avatar).toContainText('JS');
  });

  test('registration date defaults to today', async () => {
    const today = new Date().toISOString().slice(0, 10);
    const datePicker = regPage.page.locator('input#child-registration-date');
    await expect(datePicker).toHaveValue(today);
  });
});
