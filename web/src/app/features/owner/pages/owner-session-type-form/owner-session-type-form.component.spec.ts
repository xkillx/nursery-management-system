import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { ActivatedRoute, convertToParamMap, provideRouter } from '@angular/router';

import { AuthService } from '../../../../core/services/auth.service';
import { OwnerSessionTypeFormComponent } from './owner-session-type-form.component';

describe('OwnerSessionTypeFormComponent', () => {
  it('mounts in create mode', async () => {
    await TestBed.configureTestingModule({
      imports: [OwnerSessionTypeFormComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { paramMap: convertToParamMap({}) } },
        },
        { provide: AuthService, useValue: {} },
      ],
    }).compileComponents();

    const fixture = TestBed.createComponent(OwnerSessionTypeFormComponent);
    expect(fixture.componentInstance).toBeTruthy();
    expect(fixture.componentInstance.mode).toBe('create');
  });
});
