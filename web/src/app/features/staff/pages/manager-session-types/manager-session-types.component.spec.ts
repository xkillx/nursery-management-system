import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { AuthService } from '../../../../core/services/auth.service';
import { ManagerSessionTypesComponent } from './manager-session-types.component';

describe('ManagerSessionTypesComponent', () => {
  it('mounts', async () => {
    await TestBed.configureTestingModule({
      imports: [ManagerSessionTypesComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: AuthService, useValue: {} },
      ],
    }).compileComponents();

    const fixture = TestBed.createComponent(ManagerSessionTypesComponent);
    expect(fixture.componentInstance).toBeTruthy();
  });
});
