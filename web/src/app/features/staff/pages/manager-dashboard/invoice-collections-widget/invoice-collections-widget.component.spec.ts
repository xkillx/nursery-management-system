import { ComponentFixture, TestBed } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { of, throwError } from 'rxjs';

import { InvoiceCollectionsWidgetComponent } from './invoice-collections-widget.component';
import { ManagerInvoicesApiService } from '../../../data/manager-invoices-api.service';
import { OverdueSummary } from '../../../models/manager-invoices.models';

const mockOverdueSummary: OverdueSummary = {
  totalOverdueMinor: 67500,
  overdueCount: 3,
  items: [
    { id: 'inv-1', invoiceNumber: 'INV-001', childId: 'c1', childName: 'Alice Smith', outstandingMinor: 22500, dueDate: '2026-06-01', daysOverdue: 44 },
    { id: 'inv-2', invoiceNumber: 'INV-002', childId: 'c2', childName: 'Bob Jones', outstandingMinor: 22500, dueDate: '2026-06-15', daysOverdue: 30 },
    { id: 'inv-3', invoiceNumber: 'INV-003', childId: 'c3', childName: 'Charlie Brown', outstandingMinor: 22500, dueDate: '2026-07-10', daysOverdue: 5 },
  ],
};

const emptySummary: OverdueSummary = {
  totalOverdueMinor: 0,
  overdueCount: 0,
  items: [],
};

function createMockApi(overrides: Partial<ManagerInvoicesApiService> = {}): jasmine.SpyObj<ManagerInvoicesApiService> {
  const mock = jasmine.createSpyObj('ManagerInvoicesApiService', ['getOverdueSummary']);
  mock.getOverdueSummary.and.returnValue(of(mockOverdueSummary));
  return Object.assign(mock, overrides) as jasmine.SpyObj<ManagerInvoicesApiService>;
}

describe('InvoiceCollectionsWidgetComponent', () => {
  let fixture: ComponentFixture<InvoiceCollectionsWidgetComponent>;
  let native: HTMLElement;
  let mockApi: jasmine.SpyObj<ManagerInvoicesApiService>;

  async function configureTestingModule(apiOverrides: Partial<ManagerInvoicesApiService> = {}): Promise<void> {
    mockApi = createMockApi(apiOverrides);
    await TestBed.configureTestingModule({
      imports: [InvoiceCollectionsWidgetComponent, RouterTestingModule],
      providers: [{ provide: ManagerInvoicesApiService, useValue: mockApi }],
    }).compileComponents();

    fixture = TestBed.createComponent(InvoiceCollectionsWidgetComponent);
    fixture.detectChanges();
    native = fixture.nativeElement;
  }

  describe('with overdue data', () => {
    beforeEach(async () => {
      await configureTestingModule();
    });

    it('renders the section heading', () => {
      expect(native.textContent).toContain('Invoice collections');
    });

    it('displays the total overdue amount in GBP', () => {
      expect(native.textContent).toContain('£675.00');
    });

    it('displays the overdue count', () => {
      expect(native.textContent).toContain('3 overdue invoices');
    });

    it('renders a table row for each overdue invoice', () => {
      const rows = native.querySelectorAll('tbody tr');
      expect(rows.length).toBe(3);
    });

    it('displays child names in the table', () => {
      expect(native.textContent).toContain('Alice Smith');
      expect(native.textContent).toContain('Bob Jones');
      expect(native.textContent).toContain('Charlie Brown');
    });

    it('displays invoice numbers', () => {
      expect(native.textContent).toContain('INV-001');
      expect(native.textContent).toContain('INV-002');
      expect(native.textContent).toContain('INV-003');
    });

    it('displays outstanding amounts formatted as GBP', () => {
      const text = native.textContent!;
      expect(text).toContain('£225.00');
    });

    it('displays days overdue chips', () => {
      expect(native.textContent).toContain('44d');
      expect(native.textContent).toContain('30d');
      expect(native.textContent).toContain('5d');
    });

    it('renders the View all overdue link', () => {
      const link = native.querySelector('a[routerLink="/manager/invoices"]');
      expect(link).toBeTruthy();
      expect(link!.textContent).toContain('View all overdue');
    });

    it('rows link to invoice detail pages', () => {
      const rows = native.querySelectorAll('tbody tr');
      expect(rows.length).toBeGreaterThan(0);
    });

    it('applies error styling to 15+ days overdue chips', () => {
      const chips = native.querySelectorAll('tbody tr span.inline-flex');
      const chip44d = Array.from(chips).find((el) => el.textContent?.trim() === '44d');
      expect(chip44d).toBeTruthy();
      expect(chip44d!.className).toContain('error');
    });

    it('applies warning styling to 1-14 days overdue chips', () => {
      const chips = native.querySelectorAll('tbody tr span.inline-flex');
      const chip5d = Array.from(chips).find((el) => el.textContent?.trim() === '5d');
      expect(chip5d).toBeTruthy();
      expect(chip5d!.className).toContain('warning');
    });
  });

  describe('with no overdue invoices', () => {
    beforeEach(async () => {
      await configureTestingModule();
      mockApi.getOverdueSummary.and.returnValue(of(emptySummary));
      fixture = TestBed.createComponent(InvoiceCollectionsWidgetComponent);
      fixture.detectChanges();
      native = fixture.nativeElement;
    });

    it('shows empty state message', () => {
      expect(native.textContent).toContain('No overdue invoices');
    });

    it('does not render the View all overdue link', () => {
      const link = native.querySelector('a[routerLink="/manager/invoices"]');
      expect(link).toBeFalsy();
    });
  });

  describe('when API fails', () => {
    beforeEach(async () => {
      await configureTestingModule();
      mockApi.getOverdueSummary.and.returnValue(throwError(() => new Error('network')));
      fixture = TestBed.createComponent(InvoiceCollectionsWidgetComponent);
      fixture.detectChanges();
      native = fixture.nativeElement;
    });

    it('shows error message', () => {
      expect(native.textContent).toContain('Failed to load overdue invoices.');
    });

    it('shows a retry button', () => {
      const button = native.querySelector('button');
      expect(button).toBeTruthy();
      expect(button!.textContent).toContain('Retry');
    });
  });
});
