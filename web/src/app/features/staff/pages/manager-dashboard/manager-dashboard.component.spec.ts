import { ComponentFixture, TestBed } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';

import { ManagerDashboardComponent } from './manager-dashboard.component';

describe('ManagerDashboardComponent', () => {
  let fixture: ComponentFixture<ManagerDashboardComponent>;
  let native: HTMLElement;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ManagerDashboardComponent, RouterTestingModule],
    }).compileComponents();

    fixture = TestBed.createComponent(ManagerDashboardComponent);
    fixture.detectChanges();
    native = fixture.nativeElement;
  });

  it('renders nursery-domain heading', () => {
    expect(native.textContent).toContain('Manager operations');
  });

  it('does not contain template or ecommerce terminology', () => {
    const text = native.textContent!;
    const bannedTerms = [
      'Ecommerce',
      'Sales',
      'Orders',
      'Customers',
      'Products',
      'Revenue',
      'Starter Plan',
      'Pro Plan',
      'Enterprise Plan',
    ];
    for (const term of bannedTerms) {
      expect(text).not.toContain(term);
    }
  });

  it('renders attendance summary with four tiles', () => {
    const labels = ['Checked in today', 'Not in yet', 'Enrollment incomplete', 'Incomplete attendance'];
    for (const label of labels) {
      expect(native.textContent).toContain(label);
    }
  });

  it('renders incomplete attendance triage with today-first items', () => {
    const body = native.querySelector('tbody')!;
    const rows = body.querySelectorAll('tr');
    expect(rows.length).toBeGreaterThanOrEqual(2);

    const firstDate = rows[0].querySelectorAll('td')[1]?.textContent?.trim();
    expect(firstDate).toContain('Today');
  });

  it('renders invoice run status with billing month and counts', () => {
    const text = native.textContent!;
    expect(text).toContain('June 2026');
    expect(text).toContain('Eligible');
    expect(text).toContain('Blocked');
    expect(text).toContain('Drafts');
    expect(text).toContain('Issued');
    expect(text).toContain('Last run');
  });

  it('renders payment follow-up sorted by urgency', () => {
    const sections = native.querySelectorAll('section');
    const paymentSection = Array.from(sections).find((s) =>
      s.querySelector('h2')?.textContent?.includes('Payment follow-up'),
    );
    expect(paymentSection).toBeTruthy();

    const rows = paymentSection!.querySelectorAll('tbody tr');
    expect(rows.length).toBeGreaterThanOrEqual(3);

    const statuses = Array.from(rows).map((row) => {
      const badge = row.querySelector('app-status-badge');
      return badge?.textContent?.trim() ?? '';
    });

    const firstOverdue = statuses.findIndex((s) => s === 'Overdue');
    const firstFailed = statuses.findIndex((s) => s === 'Payment failed');
    const firstIssued = statuses.findIndex((s) => s === 'Issued');

    expect(firstOverdue).toBeLessThan(firstFailed);
    expect(firstFailed).toBeLessThan(firstIssued);
  });

  it('renders GBP amounts with pound symbol and two decimals', () => {
    const text = native.textContent!;
    expect(text).toContain('£450.00');
    expect(text).toContain('£225.00');
  });

  it('renders enabled quick actions as router links to existing routes', () => {
    const quickActionSection = Array.from(native.querySelectorAll('section')).find((s) =>
      s.querySelector('h2')?.textContent?.includes('Quick actions'),
    );
    expect(quickActionSection).toBeTruthy();

    const links = quickActionSection!.querySelectorAll('a');
    const hrefs = Array.from(links).map((a) => a.getAttribute('href') ?? '');

    expect(hrefs.some((h) => h.includes('/manager/attendance'))).toBe(true);
    expect(hrefs.some((h) => h.includes('/manager/children'))).toBe(true);
  });

  it('renders disabled future actions with aria-disabled and no navigation', () => {
    const disabled = native.querySelectorAll('[aria-disabled="true"]');
    expect(disabled.length).toBe(1);

    const disabledText = Array.from(disabled).map((el) => el.textContent?.trim() ?? '');
    expect(disabledText.some((t) => t.includes('Review payment follow-up'))).toBe(true);

    const disabledAnchors = Array.from(disabled).filter((el) => el.tagName === 'A');
    expect(disabledAnchors.length).toBe(0);
  });

});
