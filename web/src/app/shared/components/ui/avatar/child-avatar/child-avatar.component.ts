import { CommonModule } from '@angular/common';
import { Component, effect, inject, input, signal } from '@angular/core';
import { ChildPhotoService } from '../../../../services/child-photo.service';

@Component({
  selector: 'app-child-avatar',
  imports: [CommonModule],
  template: `<div
    class="flex items-center justify-center overflow-hidden"
    [ngClass]="[sizeClass, shapeClass, bgClass]"
  >
    @if (resolvedUrl() && !imageError) {
      <img
        [src]="resolvedUrl()"
        [alt]="name()"
        class="h-full w-full object-cover"
        (error)="onImageError()"
      />
    } @else {
      <span class="font-medium" [ngClass]="textSizeClass">{{ initials }}</span>
    }
  </div>`,
})
export class ChildAvatarComponent {
  private readonly photoService = inject(ChildPhotoService);

  photoUrl = input<string | null>(null);
  name = input('');
  size = input<'sm' | 'md' | 'lg'>('md');
  shape = input<'circle' | 'rounded'>('circle');
  statusColor = input<string | null>(null);

  imageError = false;
  resolvedUrl = signal<string | null>(null);

  constructor() {
    effect((onCleanup) => {
      const url = this.photoUrl();
      if (!url) {
        this.resolvedUrl.set(null);
        return;
      }
      const sub = this.photoService.getPhotoUrl(url).subscribe({
        next: (blobUrl) => this.resolvedUrl.set(blobUrl),
        error: () => this.resolvedUrl.set(null),
      });
      onCleanup(() => sub.unsubscribe());
    });
  }

  get initials(): string {
    const n = this.name();
    if (!n) return '';
    const parts = n.trim().split(/\s+/);
    if (parts.length === 1) return parts[0][0].toUpperCase();
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
  }

  get sizeClass(): string {
    const sizes: Record<string, string> = {
      sm: 'size-8',
      md: 'size-10',
      lg: 'size-12',
    };
    return sizes[this.size()] || sizes['md'];
  }

  get textSizeClass(): string {
    const sizes: Record<string, string> = {
      sm: 'text-xs',
      md: 'text-sm',
      lg: 'text-base',
    };
    return sizes[this.size()] || sizes['md'];
  }

  get shapeClass(): string {
    return this.shape() === 'rounded' ? 'rounded-lg' : 'rounded-full';
  }

  get bgClass(): string {
    const sc = this.statusColor();
    if (sc) return sc;
    return this.deterministicColor;
  }

  private get deterministicColor(): string {
    const colors = [
      'bg-brand-100 text-brand-600',
      'bg-pink-100 text-pink-600',
      'bg-cyan-100 text-cyan-600',
      'bg-orange-100 text-orange-600',
      'bg-green-100 text-green-600',
      'bg-purple-100 text-purple-600',
      'bg-yellow-100 text-yellow-600',
      'bg-error-100 text-error-600',
    ];
    const n = this.name();
    const index = n
      .split('')
      .reduce((acc, char) => acc + char.charCodeAt(0), 0);
    return colors[index % colors.length];
  }

  onImageError(): void {
    this.imageError = true;
  }
}
