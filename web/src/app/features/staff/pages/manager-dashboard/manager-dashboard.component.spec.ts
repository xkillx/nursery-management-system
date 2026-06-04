import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ManagerDashboardComponent } from './manager-dashboard.component';

describe('ManagerDashboardComponent', () => {
  let fixture: ComponentFixture<ManagerDashboardComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ManagerDashboardComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(ManagerDashboardComponent);
    fixture.detectChanges();
  });

  it('renders nursery-domain heading', () => {
    expect(fixture.nativeElement.textContent).toContain('Manager dashboard');
  });

  it('does not contain TailAdmin demo terminology', () => {
    const text = fixture.nativeElement.textContent;
    const demoTerms = ['Ecommerce', 'Sales', 'Orders', 'Customers'];
    for (const term of demoTerms) {
      expect(text).not.toContain(term);
    }
  });
});
