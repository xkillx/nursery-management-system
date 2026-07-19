import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ChildAvatarComponent } from './child-avatar.component';
import { ChildPhotoService } from '../../../../services/child-photo.service';
import { of, throwError } from 'rxjs';

describe('ChildAvatarComponent', () => {
  let fixture: ComponentFixture<ChildAvatarComponent>;
  let component: ChildAvatarComponent;
  let photoServiceMock: jasmine.SpyObj<ChildPhotoService>;

  beforeEach(async () => {
    photoServiceMock = jasmine.createSpyObj('ChildPhotoService', ['getPhotoUrl', 'invalidate']);
    photoServiceMock.getPhotoUrl.and.callFake((url: string) => of(url));

    await TestBed.configureTestingModule({
      imports: [ChildAvatarComponent],
      providers: [{ provide: ChildPhotoService, useValue: photoServiceMock }],
    }).compileComponents();

    fixture = TestBed.createComponent(ChildAvatarComponent);
    component = fixture.componentInstance;
  });

  it('creates the component', () => {
    expect(component).toBeTruthy();
  });

  describe('initials', () => {
    it('renders correct initials from first and last name', () => {
      fixture.componentRef.setInput('name', 'John Smith');
      expect(component.initials).toBe('JS');
    });

    it('uppercases initials', () => {
      fixture.componentRef.setInput('name', 'jane doe');
      expect(component.initials).toBe('JD');
    });

    it('returns single initial for single-word name', () => {
      fixture.componentRef.setInput('name', 'Alice');
      expect(component.initials).toBe('A');
    });

    it('uses first and last word for multi-word names', () => {
      fixture.componentRef.setInput('name', 'Mary Jane Watson');
      expect(component.initials).toBe('MW');
    });

    it('returns empty string for empty name', () => {
      fixture.componentRef.setInput('name', '');
      expect(component.initials).toBe('');
    });
  });

  describe('photo display', () => {
    it('renders photo when photoUrl is provided', () => {
      fixture.componentRef.setInput('photoUrl', 'https://example.com/photo.jpg');
      fixture.componentRef.setInput('name', 'John Smith');
      fixture.detectChanges();

      const img = fixture.nativeElement.querySelector('img');
      expect(img).toBeTruthy();
      expect(photoServiceMock.getPhotoUrl).toHaveBeenCalledWith('https://example.com/photo.jpg');
    });

    it('renders initials when photoUrl is null', () => {
      fixture.componentRef.setInput('photoUrl', null);
      fixture.componentRef.setInput('name', 'John Smith');
      fixture.detectChanges();

      const img = fixture.nativeElement.querySelector('img');
      const span = fixture.nativeElement.querySelector('span');
      expect(img).toBeFalsy();
      expect(span.textContent.trim()).toBe('JS');
    });

    it('falls back to initials when image fails to load', () => {
      fixture.componentRef.setInput('photoUrl', 'https://example.com/photo.jpg');
      fixture.componentRef.setInput('name', 'John Smith');
      fixture.detectChanges();

      const img = fixture.nativeElement.querySelector('img');
      img.dispatchEvent(new Event('error'));
      fixture.detectChanges();

      const span = fixture.nativeElement.querySelector('span');
      expect(fixture.nativeElement.querySelector('img')).toBeFalsy();
      expect(span.textContent.trim()).toBe('JS');
    });

    it('falls back to initials when photo fetch fails', () => {
      photoServiceMock.getPhotoUrl.and.returnValue(throwError(() => new Error('fetch failed')));
      fixture.componentRef.setInput('photoUrl', '/api/v1/children/123/photo');
      fixture.componentRef.setInput('name', 'John Smith');
      fixture.detectChanges();

      const span = fixture.nativeElement.querySelector('span');
      expect(fixture.nativeElement.querySelector('img')).toBeFalsy();
      expect(span.textContent.trim()).toBe('JS');
    });
  });

  describe('statusColor', () => {
    it('applies statusColor class when provided', () => {
      fixture.componentRef.setInput('name', 'John Smith');
      fixture.componentRef.setInput('statusColor', 'bg-blue-500 text-white');
      fixture.detectChanges();

      const div = fixture.nativeElement.querySelector('div');
      expect(div.classList.contains('bg-blue-500')).toBeTrue();
      expect(div.classList.contains('text-white')).toBeTrue();
    });

    it('applies deterministic color when no statusColor', () => {
      fixture.componentRef.setInput('name', 'John Smith');
      fixture.componentRef.setInput('statusColor', null);
      fixture.detectChanges();

      const div = fixture.nativeElement.querySelector('div');
      const classes = Array.from(div.classList) as string[];
      const hasColor = classes.some(
        (c: string) => c.startsWith('bg-') && c !== 'bg-transparent'
      );
      expect(hasColor).toBeTrue();
    });
  });

  describe('size', () => {
    it('applies sm size class', () => {
      fixture.componentRef.setInput('name', 'John Smith');
      fixture.componentRef.setInput('size', 'sm');
      fixture.detectChanges();

      const div = fixture.nativeElement.querySelector('div');
      expect(div.classList.contains('size-8')).toBeTrue();
    });

    it('applies md size class by default', () => {
      fixture.componentRef.setInput('name', 'John Smith');
      fixture.detectChanges();

      const div = fixture.nativeElement.querySelector('div');
      expect(div.classList.contains('size-10')).toBeTrue();
    });

    it('applies lg size class', () => {
      fixture.componentRef.setInput('name', 'John Smith');
      fixture.componentRef.setInput('size', 'lg');
      fixture.detectChanges();

      const div = fixture.nativeElement.querySelector('div');
      expect(div.classList.contains('size-12')).toBeTrue();
    });

    it('applies full size class and sets host classes', () => {
      fixture.componentRef.setInput('name', 'John Smith');
      fixture.componentRef.setInput('size', 'full');
      fixture.detectChanges();

      const div = fixture.nativeElement.querySelector('div');
      expect(div.classList.contains('w-full')).toBeTrue();
      expect(div.classList.contains('aspect-square')).toBeTrue();

      const host = fixture.nativeElement;
      expect(host.classList.contains('w-full')).toBeTrue();
      expect(host.classList.contains('block')).toBeTrue();
    });
  });

  describe('shape', () => {
    it('applies circle shape by default', () => {
      fixture.componentRef.setInput('name', 'John Smith');
      fixture.detectChanges();

      const div = fixture.nativeElement.querySelector('div');
      expect(div.classList.contains('rounded-full')).toBeTrue();
    });

    it('applies rounded shape when set', () => {
      fixture.componentRef.setInput('name', 'John Smith');
      fixture.componentRef.setInput('shape', 'rounded');
      fixture.detectChanges();

      const div = fixture.nativeElement.querySelector('div');
      expect(div.classList.contains('rounded-lg')).toBeTrue();
    });
  });
});
