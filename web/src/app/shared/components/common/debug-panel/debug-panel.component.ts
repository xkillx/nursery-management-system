import { CommonModule } from '@angular/common';
import { Component, Input } from '@angular/core';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroCheck, heroXMark, heroClipboardDocument } from '@ng-icons/heroicons/outline';
import { environment } from '../../../../../environments/environment';

@Component({
  selector: 'app-debug-panel',
  imports: [CommonModule, NgIcon],
  providers: [provideIcons({ heroCheck, heroXMark, heroClipboardDocument })],
  template: `
    @if (!isProduction) {
      <button
        type="button"
        (click)="toggle()"
        class="fixed bottom-4 right-4 z-50 rounded-full border border-gray-300 bg-white px-3 py-1.5 text-xs font-mono shadow-lg hover:bg-gray-100 dark:border-gray-600 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700"
      >
        {{ open ? '✕ Close' : '🔍 Debug' }}
      </button>
    }

    @if (open) {
      <div
        class="fixed bottom-16 right-4 z-50 overflow-auto rounded-lg border border-gray-300 bg-gray-900 shadow-xl"
        [class]="maxHeight + ' ' + maxWidth"
      >
        <div class="flex items-center justify-between border-b border-gray-700 px-3 py-2">
          <span class="font-mono text-xs font-semibold text-gray-400">{{ title }}</span>
          <div class="flex items-center gap-1">
            <button
              type="button"
              (click)="copy()"
              class="rounded p-1 text-gray-400 hover:bg-gray-700 hover:text-white"
              title="Copy JSON"
            >
              @if (copied) {
                <ng-icon name="heroCheck" size="14" class="text-green-400" />
              } @else {
                <ng-icon name="heroClipboardDocument" size="14" />
              }
            </button>
            <button
              type="button"
              (click)="toggle()"
              class="rounded p-1 text-gray-400 hover:bg-gray-700 hover:text-white"
              title="Close"
            >
              <ng-icon name="heroXMark" size="14" />
            </button>
          </div>
        </div>
        <pre class="whitespace-pre-wrap p-4 font-mono text-xs leading-5 text-green-400">{{ formattedJson }}</pre>
      </div>
    }
  `,
})
export class DebugPanelComponent {
  @Input() data: Record<string, unknown> | null = null;
  @Input() title = 'Debug';
  @Input() maxHeight = 'max-h-96';
  @Input() maxWidth = 'w-[480px]';

  protected open = false;
  protected copied = false;
  protected readonly isProduction = environment.production;

  get formattedJson(): string {
    return JSON.stringify(this.data ?? {}, null, 2);
  }

  protected toggle(): void {
    this.open = !this.open;
  }

  protected copy(): void {
    navigator.clipboard.writeText(this.formattedJson);
    this.copied = true;
    setTimeout(() => (this.copied = false), 2000);
  }
}
