import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, of, shareReplay, switchMap } from 'rxjs';

@Injectable({ providedIn: 'root' })
export class ChildPhotoService {
  private readonly http = inject(HttpClient);
  private readonly cache = new Map<string, Observable<string>>();

  getPhotoUrl(apiUrl: string): Observable<string> {
    const cached = this.cache.get(apiUrl);
    if (cached) return cached;

    const blobUrl$ = this.http.get(apiUrl, { responseType: 'blob' }).pipe(
      switchMap((blob) => {
        const objectUrl = URL.createObjectURL(blob);
        return of(objectUrl);
      }),
      shareReplay({ bufferSize: 1, refCount: true }),
    );

    this.cache.set(apiUrl, blobUrl$);
    return blobUrl$;
  }

  invalidate(apiUrl: string): void {
    this.cache.delete(apiUrl);
  }
}
