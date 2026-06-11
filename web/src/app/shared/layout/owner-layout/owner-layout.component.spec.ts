import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { OwnerLayoutComponent } from './owner-layout.component';

describe('OwnerLayoutComponent', () => {
  let fixture: ComponentFixture<OwnerLayoutComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [OwnerLayoutComponent],
      providers: [provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(OwnerLayoutComponent);
    fixture.detectChanges();
  });

  it('renders owner overview link', () => {
    const link = fixture.nativeElement.querySelector('[data-testid="owner-link-overview"]');
    expect(link).toBeTruthy();
    expect(link.getAttribute('href')).toContain('/owner');
    expect(link.textContent.trim()).toBe('Overview');
  });

  it('renders manager access link', () => {
    const link = fixture.nativeElement.querySelector('[data-testid="owner-link-manager-access"]');
    expect(link).toBeTruthy();
    expect(link.getAttribute('href')).toContain('/owner/manager-access');
    expect(link.textContent.trim()).toBe('Manager access');
  });

  it('renders router-outlet', () => {
    const outlet = fixture.nativeElement.querySelector('router-outlet');
    expect(outlet).toBeTruthy();
  });

  it('does not render staff sidebar or staff links', () => {
    const sidebar = fixture.nativeElement.querySelector('app-sidebar');
    expect(sidebar).toBeFalsy();

    const staffIds = [
      'staff-link-manager-dashboard',
      'staff-link-manager-children',
      'staff-link-manager-guardians',
      'staff-link-manager-invites',
      'staff-link-practitioner-attendance',
      'staff-link-manager-funding',
      'staff-link-manager-invoice-run',
      'staff-link-manager-invoices',
      'staff-link-manager-attendance-corrections',
    ];

    for (const id of staffIds) {
      expect(fixture.nativeElement.querySelector(`[data-testid="${id}"]`)).toBeFalsy();
    }
  });

  it('does not render parent navigation links', () => {
    const parentLink = fixture.nativeElement.querySelector('[data-testid="parent-link-invoices"]');
    expect(parentLink).toBeFalsy();
  });
});
