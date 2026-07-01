import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';

import { ManagerSiteProfileComponent } from './manager-site-profile.component';
import { SiteProfileResponse, SiteProfile } from '../../models/site-profile.models';

describe('ManagerSiteProfileComponent', () => {
  let component: ManagerSiteProfileComponent;
  let fixture: ComponentFixture<ManagerSiteProfileComponent>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [HttpClientTestingModule, ManagerSiteProfileComponent],
      providers: [provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(ManagerSiteProfileComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  function flushGet(response: SiteProfileResponse): void {
    const req = httpMock.expectOne('/api/v1/site-profile');
    expect(req.request.method).toBe('GET');
    req.flush(response);
  }

  it('loads empty profile with all fields bound to empty strings', () => {
    fixture.detectChanges();
    flushGet({ site_profile: null });

    expect(component.loading).toBeFalse();
    expect(component.model.nursery_name).toBe('');
    expect(component.model.phone).toBe('');
    expect(component.model.email).toBe('');
    expect(component.model.website).toBe('');
    expect(component.model.address_street).toBe('');
    expect(component.model.address_city).toBe('');
    expect(component.model.address_postcode).toBe('');
  });

  it('loads saved profile with all fields populated', () => {
    fixture.detectChanges();
    flushGet({
      site_profile: {
        nursery_name: 'Little Stars Nursery',
        description: 'A warm nursery',
        phone: '+44 161 555 0100',
        email: 'hello@littlestars.example',
        website: 'https://littlestars.example',
        address_street: '12 Acacia Ave',
        address_city: 'Manchester',
        address_postcode: 'M1 4BT',
      },
    });

    expect(component.loading).toBeFalse();
    expect(component.model.nursery_name).toBe('Little Stars Nursery');
    expect(component.model.phone).toBe('+44 161 555 0100');
    expect(component.model.address_postcode).toBe('M1 4BT');
  });

  it('submit with all-empty fields: no service call, each field shows error', () => {
    fixture.detectChanges();
    flushGet({ site_profile: null });

    component.submit({ control: { markAllAsTouched: () => {} } } as any);

    expect(Object.keys(component.fieldErrors).length).toBeGreaterThanOrEqual(7);
    expect(component.fieldErrors.nursery_name).toBe('is required');
    expect(component.fieldErrors.phone).toBe('is required');
    expect(component.fieldErrors.email).toBe('is required');
    expect(component.fieldErrors.website).toBe('is required');
    expect(component.fieldErrors.address_street).toBe('is required');
    expect(component.fieldErrors.address_city).toBe('is required');
    expect(component.fieldErrors.address_postcode).toBe('is required');
  });

  it('submit with whitespace-only name: shows required error', () => {
    fixture.detectChanges();
    flushGet({ site_profile: null });

    component.model.nursery_name = '   ';
    component.model.phone = '+44 161 555 0100';
    component.model.email = 'test@example.com';
    component.model.website = 'https://example.com';
    component.model.address_street = '12 Acacia Ave';
    component.model.address_city = 'Manchester';
    component.model.address_postcode = 'M1 4BT';

    component.submit({ control: { markAllAsTouched: () => {} } } as any);

    expect(component.fieldErrors.nursery_name).toBe('is required');
  });

  it('submit valid form: calls service once, navigates on success', () => {
    const navigateSpy = spyOn(component['router'], 'navigate');
    fixture.detectChanges();
    flushGet({ site_profile: null });

    component.model.nursery_name = 'Little Stars Nursery';
    component.model.phone = '+44 161 555 0100';
    component.model.email = 'hello@littlestars.example';
    component.model.website = 'https://littlestars.example';
    component.model.address_street = '12 Acacia Ave';
    component.model.address_city = 'Manchester';
    component.model.address_postcode = 'M1 4BT';

    component.submit({ control: { markAllAsTouched: () => {} } } as any);

    const putReq = httpMock.expectOne('/api/v1/site-profile');
    expect(putReq.request.method).toBe('PUT');
    putReq.flush({});

    expect(navigateSpy).toHaveBeenCalledWith(['/manager/site-settings']);
  });

  it('service returns 400 with multi-field error: maps to field errors', () => {
    fixture.detectChanges();
    flushGet({ site_profile: null });

    component.model.nursery_name = 'Little Stars Nursery';
    component.model.phone = '+44 161 555 0100';
    component.model.email = 'hello@littlestars.example';
    component.model.website = 'https://littlestars.example';
    component.model.address_street = '12 Acacia Ave';
    component.model.address_city = 'Manchester';
    component.model.address_postcode = 'M1 4BT';

    component.submit({ control: { markAllAsTouched: () => {} } } as any);

    const putReq = httpMock.expectOne('/api/v1/site-profile');
    putReq.flush(
      {
        code: 'validation_error',
        details: {
          field_errors: [
            { field: 'nursery_name', message: 'is required' },
            { field: 'phone', message: 'must be 32 characters or fewer' },
          ],
        },
      },
      { status: 400, statusText: 'Bad Request' },
    );

    expect(component.fieldErrors.nursery_name).toBe('is required');
    expect(component.fieldErrors.phone).toBe('must be 32 characters or fewer');
  });

  it('service returns 500: shows generic error, no navigation', () => {
    const navigateSpy = spyOn(component['router'], 'navigate');
    fixture.detectChanges();
    flushGet({ site_profile: null });

    component.model.nursery_name = 'Little Stars Nursery';
    component.model.phone = '+44 161 555 0100';
    component.model.email = 'hello@littlestars.example';
    component.model.website = 'https://littlestars.example';
    component.model.address_street = '12 Acacia Ave';
    component.model.address_city = 'Manchester';
    component.model.address_postcode = 'M1 4BT';

    component.submit({ control: { markAllAsTouched: () => {} } } as any);

    const putReq = httpMock.expectOne('/api/v1/site-profile');
    putReq.flush({ message: 'Server error' }, { status: 500, statusText: 'Server Error' });

    expect(navigateSpy).not.toHaveBeenCalled();
  });

  it('cancel button navigates back without service call', () => {
    const navigateSpy = spyOn(component['router'], 'navigate');
    fixture.detectChanges();
    flushGet({ site_profile: null });

    component.onCancel();

    expect(navigateSpy).toHaveBeenCalledWith(['/manager/site-settings']);
  });
});
