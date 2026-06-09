import { InvoiceRunMockService } from './invoice-run-mock.service';

describe('InvoiceRunMockService', () => {
  let service: InvoiceRunMockService;

  beforeEach(() => {
    service = new InvoiceRunMockService();
  });

  describe('loadPreflight', () => {
    it('returns preflight with eligible and blocked children for 2026-05', (done) => {
      service.loadPreflight('2026-05').subscribe((preflight) => {
        expect(preflight.billingMonth).toBe('2026-05');
        expect(preflight.summary.eligibleChildren).toBe(4);
        expect(preflight.summary.blockedChildren).toBe(2);
        expect(preflight.eligibleChildren.length).toBe(4);
        expect(preflight.blockedChildren.length).toBe(2);
        done();
      });
    });

    it('returns all eligible for 2026-03', (done) => {
      service.loadPreflight('2026-03').subscribe((preflight) => {
        expect(preflight.summary.blockedChildren).toBe(0);
        expect(preflight.eligibleChildren.length).toBe(2);
        done();
      });
    });

    it('returns no eligible for 2026-01', (done) => {
      service.loadPreflight('2026-01').subscribe((preflight) => {
        expect(preflight.summary.eligibleChildren).toBe(0);
        expect(preflight.blockedChildren.length).toBe(1);
        done();
      });
    });
  });

  describe('generateDrafts', () => {
    it('generates drafts only for eligible children', (done) => {
      service.generateDrafts('2026-05').subscribe((result) => {
        expect(result.generatedCount).toBe(4);
        expect(result.blockedCount).toBe(2);
        done();
      });
    });

    it('includes a zero-total draft', (done) => {
      service.generateDrafts('2026-05').subscribe(() => {
        service.listDrafts('2026-05').subscribe((drafts) => {
          const zeroDraft = drafts.find(d => d.childName === 'Noah Williams');
          expect(zeroDraft).toBeDefined();
          expect(zeroDraft!.netDueMinor).toBe(0);
          done();
        });
      });
    });

    it('produces deterministic line totals', (done) => {
      service.generateDrafts('2026-05').subscribe(() => {
        service.listDrafts('2026-05').subscribe((drafts) => {
          const amira = drafts.find(d => d.childName === 'Amira Hassan');
          expect(amira).toBeDefined();
          expect(amira!.netDueMinor).toBe(33600);
          expect(amira!.lines.length).toBe(3);
          done();
        });
      });
    });

    it('marks regeneration as updated', (done) => {
      service.generateDrafts('2026-05').subscribe(() => {
        service.generateDrafts('2026-05').subscribe((result) => {
          expect(result.updatedCount).toBe(4);
          expect(result.generatedCount).toBe(0);
          done();
        });
      });
    });

    it('does not mutate issued drafts on regeneration', (done) => {
      service.generateDrafts('2026-05').subscribe(() => {
        service.listDrafts('2026-05').subscribe((drafts) => {
          const amiraId = drafts.find(d => d.childName === 'Amira Hassan')!.invoiceId;
          service.bulkIssue('2026-05', [amiraId]).subscribe(() => {
            service.generateDrafts('2026-05').subscribe(() => {
              service.listDrafts('2026-05').subscribe((updated) => {
                const amira = updated.find(d => d.invoiceId === amiraId);
                expect(amira!.status).toBe('issued');
                done();
              });
            });
          });
        });
      });
    });
  });

  describe('bulkIssue', () => {
    it('assigns invoice numbers in child-name order', (done) => {
      service.generateDrafts('2026-05').subscribe(() => {
        service.listDrafts('2026-05').subscribe((drafts) => {
          const ids = drafts.filter(d => d.netDueMinor > 0).map(d => d.invoiceId);
          service.bulkIssue('2026-05', ids).subscribe((result) => {
            expect(result.issued.length).toBe(3);
            expect(result.issued[0].childName).toBe('Amira Hassan');
            expect(result.issued[0].invoiceNumber).toBe('INV-202605-0001');
            expect(result.issued[1].childName).toBe('Arjun Patel');
            expect(result.issued[1].invoiceNumber).toBe('INV-202605-0002');
            expect(result.issued[2].childName).toBe('Emma Chen');
            expect(result.issued[2].invoiceNumber).toBe('INV-202605-0003');
            done();
          });
        });
      });
    });

    it('issues zero-total draft', (done) => {
      service.generateDrafts('2026-05').subscribe(() => {
        service.listDrafts('2026-05').subscribe((drafts) => {
          const zeroId = drafts.find(d => d.childName === 'Noah Williams')!.invoiceId;
          service.bulkIssue('2026-05', [zeroId]).subscribe((result) => {
            expect(result.issued.length).toBe(1);
            expect(result.issued[0].totalMinor).toBe(0);
            expect(result.issued[0].invoiceNumber).toBe('INV-202605-0001');
            done();
          });
        });
      });
    });

    it('skips deselected drafts', (done) => {
      service.generateDrafts('2026-05').subscribe(() => {
        service.listDrafts('2026-05').subscribe((drafts) => {
          const amiraId = drafts.find(d => d.childName === 'Amira Hassan')!.invoiceId;
          service.bulkIssue('2026-05', [amiraId]).subscribe((result) => {
            expect(result.issued.length).toBe(1);
            done();
          });
        });
      });
    });
  });

  describe('issueOne', () => {
    it('issues single draft by id', (done) => {
      service.generateDrafts('2026-05').subscribe(() => {
        service.listDrafts('2026-05').subscribe((drafts) => {
          const arjunId = drafts.find(d => d.childName === 'Arjun Patel')!.invoiceId;
          service.issueOne(arjunId).subscribe((result) => {
            expect(result.issued.length).toBe(1);
            expect(result.issued[0].childName).toBe('Arjun Patel');
            done();
          });
        });
      });
    });

    it('returns skip for unknown invoice', (done) => {
      service.issueOne('inv-unknown').subscribe((result) => {
        expect(result.issuedCount).toBe(0);
        expect(result.skipped.length).toBe(1);
        done();
      });
    });
  });

  describe('resetMonth', () => {
    it('clears state for a billing month', (done) => {
      service.generateDrafts('2026-05').subscribe(() => {
        service.resetMonth('2026-05');
        service.listDrafts('2026-05').subscribe((drafts) => {
          expect(drafts.length).toBe(0);
          done();
        });
      });
    });
  });
});
