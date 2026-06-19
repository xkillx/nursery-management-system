import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { AuthService } from '../../../../core/services/auth.service';
import { OwnerSessionTypesComponent } from './owner-session-types.component';

describe('OwnerSessionTypesComponent', () => {
  it('mounts', async () => {
    await TestBed.configureTestingModule({
      imports: [OwnerSessionTypesComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: AuthService, useValue: {} },
      ],
    }).compileComponents();

    const fixture = TestBed.createComponent(OwnerSessionTypesComponent);
    expect(fixture.componentInstance).toBeTruthy();
  });
});
