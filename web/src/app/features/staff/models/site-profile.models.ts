export interface SiteProfile {
  nursery_name: string;
  description: string;
  phone: string;
  email: string;
  website: string;
  address_street: string;
  address_city: string;
  address_postcode: string;
}

export interface SiteProfileResponse {
  site_profile: SiteProfile | null;
}

export interface SiteProfileInput {
  nursery_name: string;
  description: string;
  phone: string;
  email: string;
  website: string;
  address_street: string;
  address_city: string;
  address_postcode: string;
}

export interface ApiFieldError {
  field: string;
  message: string;
}

export interface ApiValidationResponse {
  code: string;
  details?: {
    field_errors?: ApiFieldError[];
  };
}
