import { ComponentFixture, TestBed } from '@angular/core/testing';
import { FormFieldComponent } from './form-field.component';

describe('FormFieldComponent', () => {
  let fixture: ComponentFixture<FormFieldComponent>;
  let component: FormFieldComponent;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [FormFieldComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(FormFieldComponent);
    component = fixture.componentInstance;
  });

  it('creates the component', () => {
    expect(component).toBeTruthy();
  });

  it('renders label text', () => {
    component.label = 'Full name';
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Full name');
  });

  it('renders required marker when required', () => {
    component.label = 'Name';
    component.required = true;
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('*');
  });

  it('renders error message', () => {
    component.error = 'This field is required';
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('This field is required');
  });

  it('renders hint when no error', () => {
    component.hint = 'Enter at least 3 characters';
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Enter at least 3 characters');
  });

  it('does not render hint when error is present', () => {
    component.hint = 'Some hint';
    component.error = 'Error';
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).not.toContain('Some hint');
  });

  it('connects label to input via for attribute', () => {
    component.label = 'Email';
    component.labelFor = 'email-field';
    fixture.detectChanges();
    const label = (fixture.nativeElement as HTMLElement).querySelector('label');
    expect(label?.getAttribute('for')).toBe('email-field');
  });
});
