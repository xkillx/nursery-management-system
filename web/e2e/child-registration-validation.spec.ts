import { test, expect } from '@playwright/test';
import { loginAsManager, navigateToNewRegistration, ensureTestRoomExists } from './helpers/auth';
import { ChildRegistrationPage } from './page-objects/child-registration.page';

test.describe('Child Registration - Validation', () => {
  let regPage: ChildRegistrationPage;

  test.beforeEach(async ({ page }) => {
    await loginAsManager(page);
    await ensureTestRoomExists(page);
    await navigateToNewRegistration(page);
    regPage = new ChildRegistrationPage(page);
  });

  test.describe('Step 1 - Required fields', () => {
    test('blocks when first name is empty', async () => {
      await regPage.fillChildBasics({
        firstName: '',
        dateOfBirth: '2022-03-15',
        startDate: '2026-09-01',
        homeAddress: '123 High Street',
      });
      await regPage.selectDisabilityStatus('no');
      await regPage.clickContinue();
      await regPage.expectFieldError("Enter the child's first name");
      await regPage.expectStepActive(1);
    });

    test('blocks when date of birth is empty', async () => {
      await regPage.fillChildBasics({
        firstName: 'James',
        dateOfBirth: '',
        startDate: '2026-09-01',
        homeAddress: '123 High Street',
      });
      await regPage.selectDisabilityStatus('no');
      await regPage.clickContinue();
      await regPage.expectFieldError("Enter the child's date of birth");
    });

    test('blocks when date of birth is in the future', async () => {
      const futureDate = new Date(Date.now() + 86400000 * 365).toISOString().slice(0, 10);
      await regPage.fillChildBasics({
        firstName: 'James',
        dateOfBirth: futureDate,
        startDate: '2026-09-01',
        homeAddress: '123 High Street',
      });
      await regPage.selectDisabilityStatus('no');
      await regPage.clickContinue();
      await regPage.expectFieldError('Date of birth cannot be in the future');
    });

    test('blocks when start date is empty', async () => {
      await regPage.fillChildBasics({
        firstName: 'James',
        dateOfBirth: '2022-03-15',
        startDate: '',
        homeAddress: '123 High Street',
      });
      await regPage.selectDisabilityStatus('no');
      await regPage.clickContinue();
      await regPage.expectFieldError('Enter the proposed start date');
    });

    test('blocks when home address is empty', async () => {
      await regPage.fillChildBasics({
        firstName: 'James',
        dateOfBirth: '2022-03-15',
        startDate: '2026-09-01',
        homeAddress: '',
      });
      await regPage.selectDisabilityStatus('no');
      await regPage.clickContinue();
      await regPage.expectFieldError("Enter the child's home address");
    });

    test('blocks when disability status is not selected', async () => {
      await regPage.fillChildBasics({
        firstName: 'James',
        dateOfBirth: '2022-03-15',
        startDate: '2026-09-01',
        homeAddress: '123 High Street',
      });
      await regPage.clickContinue();
      await regPage.expectFieldError('Confirm whether the child has a disability');
    });

    test('allows proceeding when all required fields are filled', async () => {
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
      await regPage.expectStepActive(2);
    });
  });

  test.describe('Step 2 - Required fields', () => {
    test.beforeEach(async () => {
      await fillAndAdvanceToStep2(regPage);
    });

    test('blocks when allergy status is blank', async () => {
      await regPage.selectMedicationStatus('no');
      await regPage.selectImmunisationStatus('up_to_date');
      await regPage.selectDietaryStatus('none');
      await regPage.selectMedicalHistoryStatus('none');
      await regPage.selectSocialServicesStatus('no');
      await regPage.selectDevelopmentalConcerns('no');
      await regPage.clickContinue();
      await regPage.expectFieldError('Confirm whether the child has any known allergies');
    });

    test('blocks when medication status is blank', async () => {
      await regPage.selectAllergyStatus('no');
      await regPage.selectImmunisationStatus('up_to_date');
      await regPage.selectDietaryStatus('none');
      await regPage.selectMedicalHistoryStatus('none');
      await regPage.selectSocialServicesStatus('no');
      await regPage.selectDevelopmentalConcerns('no');
      await regPage.clickContinue();
      await regPage.expectFieldError('Confirm whether the child takes regular medication');
    });

    test('blocks when allergy is yes but details are empty', async () => {
      await regPage.selectAllergyStatus('yes');
      await regPage.selectMedicationStatus('no');
      await regPage.selectImmunisationStatus('up_to_date');
      await regPage.selectDietaryStatus('none');
      await regPage.selectMedicalHistoryStatus('none');
      await regPage.selectSocialServicesStatus('no');
      await regPage.selectDevelopmentalConcerns('no');
      await regPage.clickContinue();
      await regPage.expectFieldError('Provide allergy details');
    });

    test('allows proceeding with allergy yes and details filled', async () => {
      await regPage.selectAllergyStatus('yes');
      await regPage.fillAllergyDetails('Peanuts - severe anaphylaxis');
      await regPage.selectMedicationStatus('no');
      await regPage.selectImmunisationStatus('up_to_date');
      await regPage.selectDietaryStatus('none');
      await regPage.selectMedicalHistoryStatus('none');
      await regPage.selectSocialServicesStatus('no');
      await regPage.selectDevelopmentalConcerns('no');
      await regPage.clickContinue();
      await regPage.expectStepActive(3);
    });
  });

  test.describe('Step 3 - Required fields', () => {
    test.beforeEach(async () => {
      await fillAndAdvanceToStep3(regPage);
    });

    test('blocks when parent full name is empty', async () => {
      await regPage.fillParentCarer({
        fullName: '',
        relationship: 'Mother',
        telephone: '07700 900001',
        email: 'test@example.com',
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
      await regPage.expectFieldError('Enter the primary parent/carer full name');
    });

    test('blocks when parental responsibility is not answered', async () => {
      await regPage.fillParentCarer({
        fullName: 'Sarah Johnson',
        relationship: 'Mother',
        telephone: '07700 900001',
        email: 'test@example.com',
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
      await regPage.expectFieldError('Confirm whether the primary parent/carer holds parental responsibility');
    });
  });
});

async function fillAndAdvanceToStep2(regPage: ChildRegistrationPage): Promise<void> {
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
  await regPage.expectStepActive(2);
}

async function fillAndAdvanceToStep3(regPage: ChildRegistrationPage): Promise<void> {
  await fillAndAdvanceToStep2(regPage);
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
