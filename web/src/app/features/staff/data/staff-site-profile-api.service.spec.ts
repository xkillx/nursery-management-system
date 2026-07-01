import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';

import { StaffSiteProfileApiService } from './staff-site-profile-api.service';
import { SiteProfile, SiteProfileResponse, SiteProfileInput, ApiValidationResponse } from '../models/site-profile.models';

describe('StaffSiteProfileApiService', () => {
  let service: StaffSiteProfileApiService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
    });
    service = TestBed.inject(StaffSiteProfileApiService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('getSiteProfile deserializes null', () => {
    service.getSiteProfile().subscribe((resp) => {
      expect(resp.site_profile).toBeNull();
    });

    const req = httpMock.expectOne('/api/v1/site-profile');
    expect(req.request.method).toBe('GET');
    req.flush({ site_profile: null });
  });

  it('getSiteProfile deserializes full DTO', () => {
    const mockResponse: SiteProfileResponse = {
      site_profile: {
        nursery_name: 'Little Stars Nursery',
        description: 'A warm nursery',
        phone: '+44 161 555 0100',
        email: 'hello@example.com',
        website: 'https://example.com',
        address_street: '12 Acacia Ave',
        address_city: 'Manchester',
        address_postcode: 'M1 4BT',
      },
    };

    service.getSiteProfile().subscribe((resp) => {
      const sp = resp.site_profile!;
      expect(sp.nursery_name).toBe('Little Stars Nursery');
      expect(sp.phone).toBe('+44 161 555 0100');
      expect(sp.address_postcode).toBe('M1 4BT');
    });

    const req = httpMock.expectOne('/api/v1/site-profile');
    req.flush(mockResponse);
  });

  it('updateSiteProfile posts snake_case payload', () => {
    const input: SiteProfileInput = {
      nursery_name: 'Little Stars Nursery',
      description: '',
      phone: '+44 161 555 0100',
      email: 'hello@example.com',
      website: 'https://example.com',
      address_street: '12 Acacia Ave',
      address_city: 'Manchester',
      address_postcode: 'M1 4BT',
    };

    service.updateSiteProfile(input).subscribe();

    const req = httpMock.expectOne('/api/v1/site-profile');
    expect(req.request.method).toBe('PUT');
    expect(req.request.body).toEqual(input);
    req.flush(input as SiteProfile);
  });

  it('updateSiteProfile surfaces validation error', () => {
    const input: SiteProfileInput = {
      nursery_name: '',
      description: '',
      phone: '',
      email: '',
      website: '',
      address_street: '',
      address_city: '',
      address_postcode: '',
    };

    service.updateSiteProfile(input).subscribe({
      error: (err) => {
        const body = err.error as ApiValidationResponse;
        expect(body.code).toBe('validation_error');
        expect(body.details?.field_errors?.length).toBeGreaterThan(0);
      },
    });

    const req = httpMock.expectOne('/api/v1/site-profile');
    req.flush(
      {
        code: 'validation_error',
        details: {
          field_errors: [
            { field: 'nursery_name', message: 'is required' },
            { field: 'phone', message: 'is required' },
          ],
        },
      },
      { status: 400, statusText: 'Bad Request' },
    );
  });
});
