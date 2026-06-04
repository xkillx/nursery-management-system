import { TestBed } from '@angular/core/testing';
import { ToastService } from './toast.service';

describe('ToastService', () => {
  let service: ToastService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(ToastService);
  });

  it('creates the service', () => {
    expect(service).toBeTruthy();
  });

  it('adds a success toast', () => {
    const id = service.success('Saved successfully');
    expect(id).toBeTruthy();
    expect(service.state.toasts().length).toBe(1);
    expect(service.state.toasts()[0].variant).toBe('success');
    expect(service.state.toasts()[0].message).toBe('Saved successfully');
  });

  it('adds an error toast', () => {
    service.error('Something failed');
    const toasts = service.state.toasts();
    expect(toasts[0].variant).toBe('error');
  });

  it('adds a warning toast', () => {
    service.warning('Be careful');
    expect(service.state.toasts()[0].variant).toBe('warning');
  });

  it('adds an info toast', () => {
    service.info('FYI');
    expect(service.state.toasts()[0].variant).toBe('info');
  });

  it('dismisses a toast by id', () => {
    const id = service.success('test');
    service.dismiss(id);
    expect(service.state.toasts().length).toBe(0);
  });

  it('clears all toasts', () => {
    service.success('a');
    service.error('b');
    service.info('c');
    expect(service.state.toasts().length).toBe(3);
    service.clear();
    expect(service.state.toasts().length).toBe(0);
  });

  it('respects custom title', () => {
    service.success('msg', { title: 'Title' });
    expect(service.state.toasts()[0].title).toBe('Title');
  });

  it('limits to max visible toasts', () => {
    const ids: string[] = [];
    for (let i = 0; i < 7; i++) {
      ids.push(service.success(`toast ${i}`));
    }
    expect(service.state.toasts().length).toBe(5);
  });

  it('auto-dismisses after duration', (done) => {
    const id = service.success('auto', { durationMs: 100 });
    setTimeout(() => {
      expect(service.state.toasts().length).toBe(0);
      done();
    }, 200);
  });
});
