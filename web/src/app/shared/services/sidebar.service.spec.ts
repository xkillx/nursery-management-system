import { TestBed } from '@angular/core/testing';
import { SidebarService } from './sidebar.service';

describe('SidebarService', () => {
  let service: SidebarService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(SidebarService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('accordion state', () => {
    it('initializes with accordion expanded (default state)', (done) => {
      service.accordionExpanded$.subscribe(val => {
        expect(val).toBe(true);
        done();
      });
    });

    it('toggleAccordion flips the state', () => {
      let value = true;
      service.accordionExpanded$.subscribe(val => value = val);

      service.toggleAccordion();
      expect(value).toBe(false);

      service.toggleAccordion();
      expect(value).toBe(true);
    });

    it('setAccordionExpanded(false) collapses the accordion', (done) => {
      service.setAccordionExpanded(false);
      service.accordionExpanded$.subscribe(val => {
        expect(val).toBe(false);
        done();
      });
    });

    it('setAccordionExpanded(true) expands the accordion', (done) => {
      service.setAccordionExpanded(false);
      service.setAccordionExpanded(true);
      service.accordionExpanded$.subscribe(val => {
        expect(val).toBe(true);
        done();
      });
    });
  });
});
