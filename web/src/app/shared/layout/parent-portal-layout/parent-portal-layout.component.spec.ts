import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { ParentPortalLayoutComponent } from './parent-portal-layout.component';

describe('ParentPortalLayoutComponent', () => {
  let fixture: ComponentFixture<ParentPortalLayoutComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ParentPortalLayoutComponent],
      providers: [provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(ParentPortalLayoutComponent);
    fixture.detectChanges();
  });

  it('renders parent invoice link pointing to /app/invoices', () => {
    const link = fixture.nativeElement.querySelector('[data-testid="parent-link-invoices"]');
    expect(link).toBeTruthy();
    expect(link.getAttribute('href')).toContain('/app/invoices');
    expect(link.textContent.trim()).toBe('Invoices');
  });

  it('renders user dropdown', () => {
    const dropdown = fixture.nativeElement.querySelector('app-user-dropdown');
    expect(dropdown).toBeTruthy();
  });

  it('renders theme toggle', () => {
    const toggle = fixture.nativeElement.querySelector('app-theme-toggle-button');
    expect(toggle).toBeTruthy();
  });

  it('renders router-outlet', () => {
    const outlet = fixture.nativeElement.querySelector('router-outlet');
    expect(outlet).toBeTruthy();
  });

  it('does not render staff sidebar or sidebar toggle', () => {
    const sidebar = fixture.nativeElement.querySelector('app-sidebar');
    const backdrop = fixture.nativeElement.querySelector('app-backdrop');
    const toggleBtn = fixture.nativeElement.querySelector('[aria-label="Toggle Sidebar"]');

    expect(sidebar).toBeFalsy();
    expect(backdrop).toBeFalsy();
    expect(toggleBtn).toBeFalsy();
  });

  it('does not render staff navigation links', () => {
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

  it('contains no disabled future navigation sections', () => {
    const text = fixture.nativeElement.textContent;
    const futureLabels = ['Settings', 'Reports', 'Calendar', 'Messages', 'Profile'];

    for (const label of futureLabels) {
      expect(text).not.toContain(label);
    }
  });
});
