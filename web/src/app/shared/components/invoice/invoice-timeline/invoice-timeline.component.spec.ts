import { ComponentFixture, TestBed } from '@angular/core/testing';
import { InvoiceTimelineComponent, TimelineEntry } from './invoice-timeline.component';

describe('InvoiceTimelineComponent', () => {
  let component: InvoiceTimelineComponent;
  let fixture: ComponentFixture<InvoiceTimelineComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [InvoiceTimelineComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(InvoiceTimelineComponent);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    fixture.componentRef.setInput('entries', []);
    expect(component).toBeTruthy();
  });

  describe('nodeClasses', () => {
    it('returns solid classes for non-pending success entry', () => {
      const entry: TimelineEntry = {
        key: 'paid',
        icon: 'heroCheckBadge',
        tone: 'success',
        title: 'Paid',
        description: 'Payment received',
        timestamp: '2026-01-15T10:00:00Z',
      };
      expect(component.nodeClasses(entry)).toBe('bg-success-500 text-white');
    });

    it('returns solid classes for non-pending error entry', () => {
      const entry: TimelineEntry = {
        key: 'failed',
        icon: 'heroExclamationCircle',
        tone: 'error',
        title: 'Failed',
        description: 'Payment failed',
        timestamp: null,
      };
      expect(component.nodeClasses(entry)).toBe('bg-error-500 text-white');
    });

    it('returns solid classes for non-pending neutral entry', () => {
      const entry: TimelineEntry = {
        key: 'generated',
        icon: 'heroReceiptPercent',
        tone: 'neutral',
        title: 'Generated',
        description: 'Draft prepared',
        timestamp: '2026-01-10T09:00:00Z',
      };
      expect(component.nodeClasses(entry)).toBe('bg-gray-200 text-gray-700 dark:bg-gray-700 dark:text-gray-200');
    });

    it('returns dashed border classes for pending entry', () => {
      const entry: TimelineEntry = {
        key: 'due',
        icon: 'heroClock',
        tone: 'warning',
        title: 'Due in 5 days',
        description: 'Payment due soon',
        timestamp: null,
        isPending: true,
      };
      const classes = component.nodeClasses(entry);
      expect(classes).toContain('border-2');
      expect(classes).toContain('border-dashed');
      expect(classes).toContain('border-warning-500');
    });

    it('returns dashed border classes for pending neutral entry', () => {
      const entry: TimelineEntry = {
        key: 'draft',
        icon: 'heroReceiptPercent',
        tone: 'neutral',
        title: 'Draft',
        description: 'In draft',
        timestamp: null,
        isPending: true,
      };
      const classes = component.nodeClasses(entry);
      expect(classes).toContain('border-dashed');
      expect(classes).toContain('border-gray-300');
    });
  });

  describe('connectorClasses', () => {
    it('returns tone-matching connector class', () => {
      const entry: TimelineEntry = {
        key: 'paid',
        icon: 'heroCheckBadge',
        tone: 'success',
        title: 'Paid',
        description: '',
        timestamp: null,
      };
      expect(component.connectorClasses(entry)).toBe('bg-success-500');
    });

    it('returns error connector class', () => {
      const entry: TimelineEntry = {
        key: 'failed',
        icon: 'heroExclamationCircle',
        tone: 'error',
        title: 'Failed',
        description: '',
        timestamp: null,
      };
      expect(component.connectorClasses(entry)).toBe('bg-error-500');
    });
  });

  describe('formatTimestamp', () => {
    it('returns empty string for null', () => {
      expect(component.formatTimestamp(null)).toBe('');
    });

    it('formats a valid ISO timestamp', () => {
      const result = component.formatTimestamp('2026-01-15T10:00:00Z');
      expect(result).toBeTruthy();
      expect(result).toContain('2026');
    });
  });
});
