import { ComponentFixture, TestBed } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';

import { SignUpComponent } from './sign-up.component';

describe('SignUpComponent', () => {
  let fixture: ComponentFixture<SignUpComponent>;
  let native: HTMLElement;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [SignUpComponent, RouterTestingModule],
    }).compileComponents();

    fixture = TestBed.createComponent(SignUpComponent);
    fixture.detectChanges();
    native = fixture.nativeElement;
  });

  it('renders invitation-only heading', () => {
    expect(native.textContent).toContain('Access is invitation-only');
  });

  it('explains manager invitation policy', () => {
    const text = native.textContent!;
    expect(text).toContain('invitation');
    expect(text).toContain('manager');
  });

  it('links to /signin as the only action', () => {
    const link = native.querySelector('a[ng-reflect-router-link="/signin"]')
      ?? native.querySelector('a[href="/signin"]');
    expect(link).toBeTruthy();
    expect(link!.textContent).toContain('Return to sign in');
  });

  it('contains no public account creation form controls', () => {
    expect(native.querySelector('form')).toBeNull();
    expect(native.querySelector('input')).toBeNull();
    expect(native.querySelector('button[type="submit"]')).toBeNull();
  });

  it('contains no public signup copy or OAuth prompts', () => {
    const text = native.textContent!;
    const banned = [
      'First Name',
      'Last Name',
      'Sign up with Google',
      'Sign up with X',
      'By creating an account',
      'Create account',
      'Register',
    ];
    for (const term of banned) {
      expect(text).not.toContain(term);
    }
  });

  it('contains no signup form component', () => {
    expect(native.querySelector('app-signup-form')).toBeNull();
  });
});
