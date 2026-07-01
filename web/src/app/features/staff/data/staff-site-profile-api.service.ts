import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { apiUrl } from '../../../core/config/api.config';
import { SiteProfile, SiteProfileResponse, SiteProfileInput } from '../models/site-profile.models';

@Injectable({ providedIn: 'root' })
export class StaffSiteProfileApiService {
  private readonly http = inject(HttpClient);

  getSiteProfile(): Observable<SiteProfileResponse> {
    return this.http.get<SiteProfileResponse>(apiUrl('/site-profile'));
  }

  updateSiteProfile(input: SiteProfileInput): Observable<SiteProfile> {
    return this.http.put<SiteProfile>(apiUrl('/site-profile'), input);
  }
}
