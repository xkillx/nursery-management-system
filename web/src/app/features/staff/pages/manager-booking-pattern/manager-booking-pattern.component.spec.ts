import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { ActivatedRoute } from '@angular/router';

import { AuthService } from '../../../../core/services/auth.service';
import { ManagerBookingPatternComponent } from './manager-booking-pattern.component';

describe('ManagerBookingPatternComponent', () => {
  it('mounts', async () => {
    await TestBed.configureTestingModule({
      imports: [ManagerBookingPatternComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: {
              paramMap: {
                get: (key: string) => (key === 'childId' ? 'kid-1' : null),
              },
            },
          },
        },
        { provide: AuthService, useValue: { activeMembership: () => null } },
      ],
    }).compileComponents();

    const fixture = TestBed.createComponent(ManagerBookingPatternComponent);
    expect(fixture.componentInstance).toBeTruthy();
  });
});
