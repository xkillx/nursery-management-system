import { Component } from '@angular/core';

@Component({
  selector: 'app-parent-invoices-placeholder',
  standalone: true,
  template: `
    <div class="flex flex-col items-center justify-center p-8">
      <h1 class="text-2xl font-semibold text-gray-800 dark:text-white">Invoices</h1>
      <p class="mt-3 text-gray-500 dark:text-gray-400">Issued nursery invoices will appear here.</p>
    </div>
  `,
})
export class ParentInvoicesPlaceholderComponent {}
