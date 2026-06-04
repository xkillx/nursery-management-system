import { Component } from '@angular/core';

@Component({
  selector: 'app-manager-dashboard',
  standalone: true,
  template: `
    <div class="p-6">
      <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">Manager dashboard</h1>
      <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
        Attendance, invoice runs, and payment status will appear here.
      </p>
    </div>
  `,
})
export class ManagerDashboardComponent {}
