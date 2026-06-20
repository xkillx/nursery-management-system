import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { AuthService } from '../../../../core/services/auth.service';
import { ManagerSessionTemplatesComponent } from './manager-session-templates.component';

describe('ManagerSessionTemplatesComponent', () => {
  it('mounts', async () => {
    await TestBed.configureTestingModule({
      imports: [ManagerSessionTemplatesComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: AuthService, useValue: {} },
      ],
    }).compileComponents();

    const fixture = TestBed.createComponent(ManagerSessionTemplatesComponent);
    expect(fixture.componentInstance).toBeTruthy();
  });
});
